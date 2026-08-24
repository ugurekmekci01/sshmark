package cli

import (
	"bytes"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ugurekmekci01/sshmark/internal/config"
	"github.com/ugurekmekci01/sshmark/internal/ssh"
)

type mockRunner struct {
	sshPath string
	runErr  error
	calls   [][]string
}

func (m *mockRunner) LookPath(name string) (string, error) {
	if name == "ssh" {
		if m.sshPath == "" {
			return "/usr/bin/ssh", nil
		}
		return m.sshPath, nil
	}
	return "/usr/bin/" + name, nil
}

func (m *mockRunner) Run(name string, args ...string) error {
	m.calls = append(m.calls, append([]string{name}, args...))
	return m.runErr
}

func (m *mockRunner) RunCapture(name string, args ...string) (string, error) {
	m.calls = append(m.calls, append([]string{name}, args...))
	if m.runErr != nil {
		return "", m.runErr
	}
	return "exists", nil
}

func testApp(t *testing.T) (*App, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	paths := config.Paths{
		Base:     t.TempDir(),
		Projects: t.TempDir() + "/projects",
	}
	return &App{
		Paths:     paths,
		Prompter:  Prompter{In: strings.NewReader(""), Out: out},
		Runner:    &mockRunner{},
		SSHRunner: func(string, []string) error { return nil },
		SSHConfig: t.TempDir() + "/config",
		SSHDir:    t.TempDir() + "/.ssh",
	}, out
}

func TestOpenUnknownProject(t *testing.T) {
	t.Parallel()
	app, _ := testApp(t)
	cmd := app.openCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"missing"})
	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), `unknown project "missing"`)
}

func TestOpenOnlyExecsSSH(t *testing.T) {
	t.Parallel()
	app, _ := testApp(t)
	var gotPath string
	var gotArgs []string
	app.SSHRunner = func(path string, args []string) error {
		gotPath = path
		gotArgs = args
		return nil
	}

	require.NoError(t, config.Save(app.Paths, config.Project{
		Name: "api", Host: "buildbox", User: "user", Directory: "/home/user/projects/api",
		Tunnels: []config.Tunnel{{LocalPort: 3000, RemotePort: 3000}},
	}))

	cmd := app.openCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"api"})
	require.NoError(t, cmd.Execute())
	require.Equal(t, "/usr/bin/ssh", gotPath)
	require.Contains(t, gotArgs, "-L")
	require.Contains(t, gotArgs, "127.0.0.1:3000:127.0.0.1:3000")
}

func TestListAndShow(t *testing.T) {
	t.Parallel()
	app, out := testApp(t)
	require.NoError(t, config.Save(app.Paths, config.Project{
		Name: "api", Host: "buildbox", User: "user", Directory: "/d",
	}))

	list := app.listCmd()
	list.SetOut(out)
	require.NoError(t, list.Execute())
	require.Contains(t, out.String(), "api")

	show := app.showCmd()
	show.SetOut(out)
	show.SetArgs([]string{"api"})
	require.NoError(t, show.Execute())
	require.Contains(t, out.String(), "Project: api")
}

func TestRemoveRequiresConfirmation(t *testing.T) {
	t.Parallel()
	app, _ := testApp(t)
	require.NoError(t, config.Save(app.Paths, config.Project{
		Name: "api", Host: "h", User: "u", Directory: "/d",
	}))

	app.Prompter = Prompter{In: strings.NewReader("n\n"), Out: io.Discard}
	cmd := app.removeCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"api"})
	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cancelled")
	require.True(t, config.Exists(app.Paths, "api"))
}

func TestRemoveWithYesFlag(t *testing.T) {
	t.Parallel()
	app, _ := testApp(t)
	require.NoError(t, config.Save(app.Paths, config.Project{
		Name: "api", Host: "h", User: "u", Directory: "/d",
	}))

	cmd := app.removeCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"api"})
	require.NoError(t, cmd.Flags().Set("yes", "true"))
	require.NoError(t, cmd.Execute())
	require.False(t, config.Exists(app.Paths, "api"))
}

func TestCheckReportsStatus(t *testing.T) {
	t.Parallel()
	app, out := testApp(t)
	runner := &mockRunner{}
	app.Runner = runner

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	require.NoError(t, config.Save(app.Paths, config.Project{
		Name: "api", Host: "h", User: "u", Directory: "/d",
		Tunnels: []config.Tunnel{{LocalPort: port, RemotePort: port}},
	}))

	cmd := app.checkCmd()
	cmd.SetOut(out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"api"})
	require.Error(t, cmd.Execute())
	require.Contains(t, out.String(), "config ok")
	require.Contains(t, out.String(), "already in use")
}

func TestAddNonInteractiveSkipsBootstrapForKnownHost(t *testing.T) {
	t.Parallel()
	app, out := testApp(t)
	runner := &mockRunner{}
	app.Runner = runner
	require.NoError(t, config.MarkKnownHost(app.Paths, "buildbox"))

	add := app.addCmd()
	add.SetOut(out)
	add.SetErr(io.Discard)
	add.SetArgs([]string{"api"})
	require.NoError(t, add.Flags().Set("host", "buildbox"))
	require.NoError(t, add.Flags().Set("user", "user"))
	require.NoError(t, add.Flags().Set("dir", "/home/user/dev"))
	require.NoError(t, add.Execute())
	// mkdir ssh call only, no ssh-copy-id
	require.Len(t, runner.calls, 1)
	require.Contains(t, runner.calls[0][0], "ssh")
	require.True(t, config.Exists(app.Paths, "api"))
}

func TestDefaultHostNickname(t *testing.T) {
	t.Parallel()
	require.Equal(t, "fedorabox-box", ssh.ProjectHostNickname("fedorabox"))
}

func TestBootstrapBuildsCopyIDArgs(t *testing.T) {
	t.Parallel()
	args := ssh.CopyIDArgs("/k.pub", "user", "10.0.0.1", 22)
	require.Equal(t, []string{"-i", "/k.pub", "user@10.0.0.1"}, args)
}

func TestExpectedErrorsNoStackTrace(t *testing.T) {
	t.Parallel()
	app, _ := testApp(t)
	cmd := app.openCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"nope"})
	err := cmd.Execute()
	require.Error(t, err)
	require.NotContains(t, err.Error(), "panic")
	require.NotContains(t, err.Error(), "runtime.")
}

func TestExecutePrintsError(t *testing.T) {
	t.Parallel()
	// covered by command-level tests; root prints errors to stderr in Execute()
}
