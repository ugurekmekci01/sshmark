package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionPrintsNonEmpty(t *testing.T) {
	cmd := newRootCmd(newApp())
	cmd.SetArgs([]string{"version"})

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()
	require.NoError(t, err)

	require.NotEmpty(t, out.String())
}

func TestVersionStdoutNotStderr(t *testing.T) {
	cmd := newRootCmd(newApp())
	cmd.SetArgs([]string{"version"})

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	cmd.SetOut(os.Stdout)
	cmd.SetErr(io.Discard)

	execErr := cmd.Execute()
	w.Close()
	os.Stdout = oldStdout

	require.NoError(t, execErr)

	var buf bytes.Buffer
	_, readErr := io.Copy(&buf, r)
	require.NoError(t, readErr)
	require.NotEmpty(t, buf.String())
}
