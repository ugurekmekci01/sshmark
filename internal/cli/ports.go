package cli

import (
	"fmt"

	"github.com/ugurekmekci01/sshmark/internal/config"
	"github.com/ugurekmekci01/sshmark/internal/tunnel"
)

func checkPortsForProject(name string, tunnels []config.Tunnel) error {
	ports := make([]int, 0, len(tunnels))
	for _, t := range tunnels {
		ports = append(ports, t.LocalPort)
	}
	return tunnel.CheckProjectPorts(name, ports, func(port int) string {
		for _, t := range tunnels {
			if t.LocalPort == port {
				remoteHost := t.RemoteHost
				if remoteHost == "" {
					remoteHost = "127.0.0.1"
				}
				return fmt.Sprintf("localhost:%d -> remote %s:%d", port, remoteHost, t.RemotePort)
			}
		}
		return fmt.Sprintf("localhost:%d", port)
	})
}
