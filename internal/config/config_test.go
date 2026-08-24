package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func testPaths(t *testing.T) Paths {
	t.Helper()
	base := filepath.Join(t.TempDir(), "sshmark")
	return Paths{
		Base:     base,
		Projects: filepath.Join(base, "projects"),
	}
}

func TestParseValidProjectTOML(t *testing.T) {
	t.Parallel()
	paths := testPaths(t)
	require.NoError(t, paths.EnsureDirs())

	p := Project{
		Name:      "api",
		Host:      "buildbox",
		User:      "user",
		Directory: "/home/user/projects/api",
		Tunnels: []Tunnel{
			{LocalPort: 3000, RemoteHost: "127.0.0.1", RemotePort: 3000},
		},
	}
	require.NoError(t, Save(paths, p))

	loaded, err := Load(paths, "api")
	require.NoError(t, err)
	require.Equal(t, p.Name, loaded.Name)
	require.Equal(t, p.Host, loaded.Host)
	require.Len(t, loaded.Tunnels, 1)
}

func TestRejectMissingRequiredFields(t *testing.T) {
	t.Parallel()
	cases := []Project{
		{Host: "h", User: "u", Directory: "/d"},
		{Name: "n", User: "u", Directory: "/d"},
		{Name: "n", Host: "h", Directory: "/d"},
		{Name: "n", Host: "h", User: "u"},
	}
	for _, p := range cases {
		err := Validate(p)
		require.Error(t, err)
	}
}

func TestRejectDuplicateLocalPorts(t *testing.T) {
	t.Parallel()
	p := Project{
		Name:      "api",
		Host:      "buildbox",
		User:      "user",
		Directory: "/home/user/projects/api",
		Tunnels: []Tunnel{
			{LocalPort: 3000, RemotePort: 3000},
			{LocalPort: 3000, RemotePort: 8080},
		},
	}
	err := Validate(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate local port 3000")
}

func TestSaveLoadRoundtrip(t *testing.T) {
	t.Parallel()
	paths := testPaths(t)
	p := Project{
		Name:      "shop",
		Host:      "gpu",
		User:      "dev",
		Port:      2222,
		Directory: "/srv/shop",
		Tunnels: []Tunnel{
			{LocalPort: 13000, RemoteHost: "127.0.0.1", RemotePort: 3000},
		},
	}
	require.NoError(t, Save(paths, p))
	loaded, err := Load(paths, "shop")
	require.NoError(t, err)
	require.Equal(t, p, loaded)
}

func TestAutoCreateConfigDirOnFirstWrite(t *testing.T) {
	t.Parallel()
	base := filepath.Join(t.TempDir(), "new", "sshmark")
	paths := Paths{Base: base, Projects: filepath.Join(base, "projects")}
	_, err := os.Stat(paths.Projects)
	require.True(t, os.IsNotExist(err))

	require.NoError(t, Save(paths, Project{
		Name: "api", Host: "h", User: "u", Directory: "/d",
	}))
	_, err = os.Stat(paths.Projects)
	require.NoError(t, err)
}

func TestNoSecretFieldsInSchema(t *testing.T) {
	t.Parallel()
	paths := testPaths(t)
	raw := []byte(`name = "api"
host = "buildbox"
user = "user"
directory = "/d"
password = "secret"
private_key = "-----BEGIN"
`)
	require.NoError(t, paths.EnsureDirs())
	require.NoError(t, os.WriteFile(paths.ProjectFile("api"), raw, 0o644))

	p, err := Load(paths, "api")
	require.NoError(t, err)
	require.Equal(t, "api", p.Name)
}

func TestKnownHosts(t *testing.T) {
	t.Parallel()
	paths := testPaths(t)

	known, err := IsKnownHost(paths, "buildbox")
	require.NoError(t, err)
	require.False(t, known)

	require.NoError(t, MarkKnownHost(paths, "buildbox"))
	known, err = IsKnownHost(paths, "buildbox")
	require.NoError(t, err)
	require.True(t, known)

	require.NoError(t, MarkKnownHost(paths, "buildbox"))
	data, err := os.ReadFile(paths.KnownHostsFile())
	require.NoError(t, err)
	require.Equal(t, "buildbox\n", string(data))
}

func TestListProjects(t *testing.T) {
	t.Parallel()
	paths := testPaths(t)
	names, err := List(paths)
	require.NoError(t, err)
	require.Empty(t, names)

	for _, name := range []string{"api", "shop"} {
		require.NoError(t, Save(paths, Project{
			Name: name, Host: "h", User: "u", Directory: "/d",
		}))
	}
	names, err = List(paths)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"api", "shop"}, names)
}
