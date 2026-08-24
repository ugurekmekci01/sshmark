package ssh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectHostNickname(t *testing.T) {
	t.Parallel()
	require.Equal(t, "fedorabox-box", ProjectHostNickname("fedorabox"))
}

func TestEnsureRemoteDirScript(t *testing.T) {
	t.Parallel()
	script := EnsureRemoteDirScript("/home/u/dev")
	require.Contains(t, script, "mkdir -p")
	require.Contains(t, script, "'/home/u/dev'")
}

func TestParseRemoteDirOutput(t *testing.T) {
	t.Parallel()
	created, exists, err := ParseRemoteDirOutput("created\n")
	require.NoError(t, err)
	require.True(t, created)
	require.True(t, exists)

	created, exists, err = ParseRemoteDirOutput("exists")
	require.NoError(t, err)
	require.False(t, created)
	require.True(t, exists)

	_, exists, err = ParseRemoteDirOutput("missing")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestCopyIDAlreadyInstalled(t *testing.T) {
	t.Parallel()
	require.True(t, CopyIDAlreadyInstalled("WARNING: All keys were skipped because they already exist"))
}
