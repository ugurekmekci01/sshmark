package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ugurekmekci01/sshmark/internal/config"
	"github.com/ugurekmekci01/sshmark/internal/ssh"
)

// Runner executes external commands (for tests).
type Runner interface {
	LookPath(name string) (string, error)
	Run(name string, args ...string) error
	RunCapture(name string, args ...string) (string, error)
}

// SSHRunner runs an interactive ssh session.
type SSHRunner func(sshPath string, args []string) error

type execRunner struct{}

func (execRunner) LookPath(name string) (string, error) { return ssh.LookPath(name) }

func (execRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (execRunner) RunCapture(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := stdout.String() + stderr.String()
	return out, err
}

func defaultSSHRunner(sshPath string, args []string) error {
	cmd := exec.Command(sshPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// App holds CLI dependencies.
type App struct {
	Paths     config.Paths
	Verbose   bool
	Prompter  Prompter
	Runner    Runner
	SSHRunner SSHRunner
	SSHConfig string
	SSHDir    string
}

func newApp() *App {
	return &App{
		Paths:     config.DefaultPaths(),
		Prompter:  defaultPrompter(),
		Runner:    execRunner{},
		SSHRunner: defaultSSHRunner,
		SSHConfig: defaultSSHConfigPath(),
		SSHDir:    defaultSSHDir(),
	}
}

func parseTunnelFlag(s string) (config.Tunnel, error) {
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 1:
		port, err := strconv.Atoi(parts[0])
		if err != nil {
			return config.Tunnel{}, fmt.Errorf("invalid tunnel %q", s)
		}
		return config.Tunnel{LocalPort: port, RemoteHost: "127.0.0.1", RemotePort: port}, nil
	case 2:
		local, err := strconv.Atoi(parts[0])
		if err != nil {
			return config.Tunnel{}, fmt.Errorf("invalid tunnel %q", s)
		}
		remote, err := strconv.Atoi(parts[1])
		if err != nil {
			return config.Tunnel{}, fmt.Errorf("invalid tunnel %q", s)
		}
		return config.Tunnel{LocalPort: local, RemoteHost: "127.0.0.1", RemotePort: remote}, nil
	default:
		return config.Tunnel{}, fmt.Errorf("invalid tunnel %q (use PORT or LOCAL:REMOTE)", s)
	}
}

func looksLikeAddress(host string) bool {
	if strings.Count(host, ".") == 3 {
		return true
	}
	return strings.Contains(host, ":") || strings.Contains(host, ".")
}

func defaultRemoteDirectory(user string) string {
	return "/home/" + user + "/dev"
}

func (a *App) bootstrapHost(nickname, address, user string, port int, pubKey ssh.PublicKey) error {
	if _, err := a.Runner.LookPath("ssh-copy-id"); err != nil {
		return err
	}
	if _, err := a.Runner.LookPath("ssh"); err != nil {
		return err
	}

	copyArgs := ssh.CopyIDArgs(pubKey.Path, user, address, port)
	out, err := a.Runner.RunCapture("ssh-copy-id", copyArgs...)
	if err != nil && !ssh.CopyIDAlreadyInstalled(out) {
		return fmt.Errorf("copy SSH key to %s: %w", nickname, err)
	}
	if ssh.CopyIDAlreadyInstalled(out) {
		_, _ = fmt.Fprintln(a.Prompter.Out, "SSH key: already on remote")
	} else {
		_, _ = fmt.Fprintln(a.Prompter.Out, "SSH key: copied to remote")
	}

	block := ssh.HostBlock(nickname, address, user, port, pubKey.Private)
	if _, err := ssh.AppendHostBlock(a.SSHConfig, block, nickname); err != nil {
		return err
	}
	return config.MarkKnownHost(a.Paths, nickname)
}

func (a *App) ensureRemoteDirectory(p config.Project) (bool, error) {
	sshPath, err := a.Runner.LookPath("ssh")
	if err != nil {
		return false, err
	}
	params := openParamsFromProject(p)
	args := ssh.RemoteCommandArgs(params, ssh.EnsureRemoteDirScript(p.Directory))
	out, err := a.Runner.RunCapture(sshPath, args...)
	if err != nil {
		return false, fmt.Errorf("prepare remote directory %q on %s: %w", p.Directory, p.Host, err)
	}
	created, _, err := ssh.ParseRemoteDirOutput(out)
	if err != nil {
		return false, fmt.Errorf("prepare remote directory %q on %s: %w", p.Directory, p.Host, err)
	}
	return created, nil
}

func (a *App) reportRemoteDirectory(p config.Project, created bool) {
	if created {
		_, _ = fmt.Fprintf(a.Prompter.Out, "Created %s on %s\n", p.Directory, p.Host)
	}
}

func pathInstallHint() string {
	if _, err := exec.LookPath("sshmark"); err == nil {
		return ""
	}

	exe, err := os.Executable()
	if err != nil {
		return ""
	}

	dir, err := filepath.EvalSymlinks(filepath.Dir(exe))
	if err != nil {
		dir = filepath.Dir(exe)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	localBin, err := filepath.EvalSymlinks(filepath.Join(home, ".local", "bin"))
	if err != nil {
		localBin = filepath.Join(home, ".local", "bin")
	}

	if dir == localBin {
		return fmt.Sprintf("Add to PATH: export PATH=%q:$PATH", localBin)
	}

	return ""
}

func (a *App) selectPublicKey(keys []ssh.PublicKey) (ssh.PublicKey, error) {
	if len(keys) == 0 {
		ok, err := a.Prompter.confirm("No SSH public key found. Create a new ed25519 key?")
		if err != nil {
			return ssh.PublicKey{}, err
		}
		if !ok {
			return ssh.PublicKey{}, fmt.Errorf("SSH key required to connect to a new machine")
		}
		if _, err := a.Runner.LookPath("ssh-keygen"); err != nil {
			return ssh.PublicKey{}, err
		}
		priv := filepath.Join(a.SSHDir, "id_ed25519")
		pub := priv + ".pub"
		if err := a.Runner.Run("ssh-keygen", "-t", "ed25519", "-f", priv, "-N", ""); err != nil {
			return ssh.PublicKey{}, fmt.Errorf("create SSH key: %w", err)
		}
		return ssh.PublicKey{Path: pub, Private: priv}, nil
	}
	if len(keys) == 1 {
		return keys[0], nil
	}

	_, _ = fmt.Fprintln(a.Prompter.Out, "SSH public keys:")
	for i, k := range keys {
		_, _ = fmt.Fprintf(a.Prompter.Out, "  %d) %s\n", i+1, k.Path)
	}
	choice, err := a.Prompter.readLine("Select key", "1")
	if err != nil {
		return ssh.PublicKey{}, err
	}
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(keys) {
		return ssh.PublicKey{}, fmt.Errorf("invalid key selection")
	}
	return keys[idx-1], nil
}

func (a *App) ensureHostReady(nickname, address, user string, port int) error {
	known, err := config.IsKnownHost(a.Paths, nickname)
	if err != nil {
		return err
	}
	if known {
		return nil
	}

	keys, err := ssh.DiscoverPublicKeys(a.SSHDir)
	if err != nil {
		return err
	}
	pubKey, err := a.selectPublicKey(keys)
	if err != nil {
		return err
	}
	return a.bootstrapHost(nickname, address, user, port, pubKey)
}

func formatProjectSummary(p config.Project, verbose bool) string {
	if !verbose {
		tunnels := make([]string, 0, len(p.Tunnels))
		for _, t := range p.Tunnels {
			tunnels = append(tunnels, strconv.Itoa(t.LocalPort))
		}
		tun := ""
		if len(tunnels) > 0 {
			tun = " (tunnels: " + strings.Join(tunnels, ", ") + ")"
		}
		return fmt.Sprintf("%s → %s:%s%s", p.Name, p.Host, p.Directory, tun)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Project: %s\n", p.Name)
	fmt.Fprintf(&b, "Host: %s\n", p.Host)
	fmt.Fprintf(&b, "Directory: %s\n", p.Directory)
	if len(p.Tunnels) > 0 {
		fmt.Fprintln(&b, "Tunnels:")
		for _, t := range p.Tunnels {
			remoteHost := t.RemoteHost
			if remoteHost == "" {
				remoteHost = "127.0.0.1"
			}
			fmt.Fprintf(&b, "  localhost:%d -> remote %s:%d\n", t.LocalPort, remoteHost, t.RemotePort)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func openParamsFromProject(p config.Project) ssh.OpenParams {
	tunnels := make([]ssh.TunnelParam, 0, len(p.Tunnels))
	for _, t := range p.Tunnels {
		tunnels = append(tunnels, ssh.TunnelParam{
			LocalPort:  t.LocalPort,
			RemoteHost: t.RemoteHost,
			RemotePort: t.RemotePort,
		})
	}
	return ssh.OpenParams{
		Host:      p.Host,
		User:      p.User,
		Port:      p.Port,
		Directory: p.Directory,
		Tunnels:   tunnels,
	}
}
