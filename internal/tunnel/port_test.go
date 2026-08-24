package tunnel

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortAvailable(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	err = CheckLocalPort(port)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already in use")
}

func TestCheckProjectPortsErrorIncludesProject(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	err = CheckProjectPorts("shop", []int{port}, func(p int) string {
		return fmt.Sprintf("localhost:%d -> remote localhost:%d", p, p)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), `project "shop"`)
	require.Contains(t, err.Error(), "localhost:")
}
