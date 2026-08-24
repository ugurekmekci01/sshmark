package tunnel

import (
	"fmt"
	"net"
)

// CheckLocalPort reports whether a TCP port is available on 127.0.0.1.
func CheckLocalPort(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("local port %d is already in use", port)
	}
	_ = ln.Close()
	return nil
}

// CheckProjectPorts validates all tunnel local ports for a project.
func CheckProjectPorts(projectName string, localPorts []int, formatTunnel func(port int) string) error {
	for _, port := range localPorts {
		if err := CheckLocalPort(port); err != nil {
			msg := fmt.Sprintf("Cannot open project %q:\n%s", projectName, err.Error())
			if formatTunnel != nil {
				msg += "\n\nRequested tunnel:\n" + formatTunnel(port)
			}
			return fmt.Errorf("%s", msg)
		}
	}
	return nil
}
