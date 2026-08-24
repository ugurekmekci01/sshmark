package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PublicKey represents a discovered SSH public key.
type PublicKey struct {
	Path    string
	Private string
}

// DiscoverPublicKeys lists *.pub files in sshDir (default ~/.ssh).
func DiscoverPublicKeys(sshDir string) ([]PublicKey, error) {
	if sshDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("home directory: %w", err)
		}
		sshDir = strings.TrimSuffix(home, "/") + "/.ssh"
	}

	entries, err := os.ReadDir(sshDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", sshDir, err)
	}

	var keys []PublicKey
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pub") {
			continue
		}
		if strings.HasSuffix(e.Name(), "-cert.pub") {
			continue
		}
		pub := strings.TrimSuffix(e.Name(), ".pub")
		keys = append(keys, PublicKey{
			Path:    sshDir + "/" + e.Name(),
			Private: sshDir + "/" + pub,
		})
	}
	return keys, nil
}

// CopyIDArgs builds argv for ssh-copy-id.
func CopyIDArgs(pubKeyPath, user, address string, port int) []string {
	target := fmt.Sprintf("%s@%s", user, address)
	args := []string{"-i", pubKeyPath}
	if port > 0 && port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", port))
	}
	args = append(args, target)
	return args
}

// HostBlock formats an SSH config Host block.
func HostBlock(nickname, address, user string, port int, identityFile string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Host %s\n", nickname)
	fmt.Fprintf(&b, "    HostName %s\n", address)
	fmt.Fprintf(&b, "    User %s\n", user)
	if port > 0 && port != 22 {
		fmt.Fprintf(&b, "    Port %d\n", port)
	}
	if identityFile != "" {
		fmt.Fprintf(&b, "    IdentityFile %s\n", identityFile)
	}
	return b.String()
}

// HasHostEntry reports whether sshConfig contains a Host block for nickname.
func HasHostEntry(sshConfig, nickname string) bool {
	marker := "Host " + nickname
	for _, line := range strings.Split(sshConfig, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == marker || strings.HasPrefix(trimmed, marker+" ") {
			return true
		}
	}
	return false
}

// AppendHostBlock appends a Host block if it does not already exist.
func AppendHostBlock(sshConfigPath, block, nickname string) (bool, error) {
	data, err := os.ReadFile(sshConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read ssh config: %w", err)
	}
	content := string(data)
	if HasHostEntry(content, nickname) {
		return false, nil
	}

	f, err := os.OpenFile(sshConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return false, fmt.Errorf("write ssh config: %w", err)
	}
	defer f.Close()

	prefix := ""
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		prefix = "\n"
	}
	if _, err := fmt.Fprintf(f, "%s%s\n", prefix, strings.TrimRight(block, "\n")); err != nil {
		return false, err
	}
	return true, nil
}

// ShellQuoteSingle wraps s in single quotes for remote shell use.
func ShellQuoteSingle(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// BuildOpenArgs builds argv for sshmark open.
func BuildOpenArgs(p OpenParams) []string {
	args := []string{}
	if p.Port > 0 && p.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", p.Port))
	}
	for _, tun := range p.Tunnels {
		remoteHost := tun.RemoteHost
		if remoteHost == "" {
			remoteHost = "127.0.0.1"
		}
		spec := fmt.Sprintf("127.0.0.1:%d:%s:%d", tun.LocalPort, remoteHost, tun.RemotePort)
		args = append(args, "-L", spec)
	}
	remote := fmt.Sprintf("cd %s && exec $SHELL", ShellQuoteSingle(p.Directory))
	args = append(args, "-t", fmt.Sprintf("%s@%s", p.User, p.Host), remote)
	return args
}

// OpenParams holds inputs for building an ssh open command.
type OpenParams struct {
	Host      string
	User      string
	Port      int
	Directory string
	Tunnels   []TunnelParam
}

// TunnelParam is a tunnel for BuildOpenArgs.
type TunnelParam struct {
	LocalPort  int
	RemoteHost string
	RemotePort int
}

// ContainsInsecureSSHFlags reports whether args disable host key checking.
func ContainsInsecureSSHFlags(args []string) bool {
	joined := strings.Join(args, " ")
	return strings.Contains(joined, "StrictHostKeyChecking=no") ||
		strings.Contains(joined, "UserKnownHostsFile=/dev/null")
}

// LookPath wraps exec.LookPath for ssh binaries.
func LookPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH: install OpenSSH", name)
	}
	return path, nil
}
