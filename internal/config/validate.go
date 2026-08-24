package config

import (
	"fmt"
	"strings"
)

// Validate checks a project bookmark for required fields and tunnel rules.
func Validate(p Project) error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("project name is required")
	}
	if strings.TrimSpace(p.Host) == "" {
		return fmt.Errorf("host is required")
	}
	if strings.TrimSpace(p.User) == "" {
		return fmt.Errorf("user is required")
	}
	if strings.TrimSpace(p.Directory) == "" {
		return fmt.Errorf("directory is required")
	}
	if p.Port < 0 || p.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	seen := make(map[int]struct{}, len(p.Tunnels))
	for i, tun := range p.Tunnels {
		if tun.LocalPort < 1 || tun.LocalPort > 65535 {
			return fmt.Errorf("tunnel %d: local_port must be between 1 and 65535", i+1)
		}
		if tun.RemotePort < 1 || tun.RemotePort > 65535 {
			return fmt.Errorf("tunnel %d: remote_port must be between 1 and 65535", i+1)
		}
		if _, ok := seen[tun.LocalPort]; ok {
			return fmt.Errorf("duplicate local port %d in project %q", tun.LocalPort, p.Name)
		}
		seen[tun.LocalPort] = struct{}{}
	}
	return nil
}
