package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoverPublicKeys(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "id_ed25519.pub"), []byte("pub"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "id_rsa.pub"), []byte("pub"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README"), []byte("x"), 0o644))

	keys, err := DiscoverPublicKeys(dir)
	require.NoError(t, err)
	require.Len(t, keys, 2)
}

func TestCopyIDArgs(t *testing.T) {
	t.Parallel()
	args := CopyIDArgs("/home/u/.ssh/id_ed25519.pub", "user", "192.168.1.40", 22)
	require.Equal(t, []string{"-i", "/home/u/.ssh/id_ed25519.pub", "user@192.168.1.40"}, args)

	args = CopyIDArgs("/k.pub", "user", "host", 2222)
	require.Equal(t, []string{"-i", "/k.pub", "-p", "2222", "user@host"}, args)
}

func TestAppendHostBlockNoDuplicate(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config")
	block := HostBlock("buildbox", "192.168.1.40", "user", 22, "/home/u/.ssh/id_ed25519")

	added, err := AppendHostBlock(path, block, "buildbox")
	require.NoError(t, err)
	require.True(t, added)

	added, err = AppendHostBlock(path, block, "buildbox")
	require.NoError(t, err)
	require.False(t, added)
}

func TestBuildOpenArgs(t *testing.T) {
	t.Parallel()
	args := BuildOpenArgs(OpenParams{
		Host:      "buildbox",
		User:      "user",
		Port:      22,
		Directory: "/home/user/projects/api",
		Tunnels: []TunnelParam{
			{LocalPort: 3000, RemoteHost: "127.0.0.1", RemotePort: 3000},
		},
	})
	require.Contains(t, args, "-L")
	require.Contains(t, args, "127.0.0.1:3000:127.0.0.1:3000")
	require.Contains(t, args, "-t")
	require.Contains(t, args, "user@buildbox")
	require.Contains(t, args, "cd '/home/user/projects/api' && exec $SHELL")
	require.False(t, ContainsInsecureSSHFlags(args))
}

func TestBuildOpenArgsCustomPort(t *testing.T) {
	t.Parallel()
	args := BuildOpenArgs(OpenParams{
		Host: "h", User: "u", Port: 2222, Directory: "/d",
	})
	require.Contains(t, args, "-p")
	require.Contains(t, args, "2222")
}

func TestBuildOpenArgsEscapesDirectory(t *testing.T) {
	t.Parallel()
	args := BuildOpenArgs(OpenParams{
		Host: "h", User: "u", Directory: "/path/with spaces",
	})
	joined := stringsJoin(args)
	require.Contains(t, joined, "'/path/with spaces'")
}

func TestHostAliasPassedThrough(t *testing.T) {
	t.Parallel()
	args := BuildOpenArgs(OpenParams{
		Host: "buildbox", User: "user", Directory: "/d",
	})
	require.Contains(t, args, "user@buildbox")
}

func stringsJoin(parts []string) string {
	out := ""
	for _, p := range parts {
		out += p + " "
	}
	return out
}
