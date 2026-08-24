package config

import (
	"os"
	"path/filepath"
)

// Paths holds sshmark configuration locations.
type Paths struct {
	Base     string
	Projects string
}

// DefaultPaths returns config paths, honoring SSHMARK_CONFIG when set.
func DefaultPaths() Paths {
	base := os.Getenv("SSHMARK_CONFIG")
	if base == "" {
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				configHome = "."
			} else {
				configHome = filepath.Join(home, ".config")
			}
		}
		base = filepath.Join(configHome, "sshmark")
	}
	return Paths{
		Base:     base,
		Projects: filepath.Join(base, "projects"),
	}
}

func (p Paths) KnownHostsFile() string {
	return filepath.Join(p.Base, "known_hosts")
}

func (p Paths) ProjectFile(name string) string {
	return filepath.Join(p.Projects, name+".toml")
}

// EnsureDirs creates the config directory tree.
func (p Paths) EnsureDirs() error {
	if err := os.MkdirAll(p.Base, 0o755); err != nil {
		return err
	}
	return os.MkdirAll(p.Projects, 0o755)
}
