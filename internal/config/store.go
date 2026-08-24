package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Load reads and validates a project bookmark by name.
func Load(paths Paths, name string) (Project, error) {
	data, err := os.ReadFile(paths.ProjectFile(name))
	if err != nil {
		if os.IsNotExist(err) {
			return Project{}, fmt.Errorf("unknown project %q", name)
		}
		return Project{}, fmt.Errorf("read project %q: %w", name, err)
	}

	var p Project
	if err := toml.Unmarshal(data, &p); err != nil {
		return Project{}, fmt.Errorf("parse project %q: %w", name, err)
	}
	if p.Name == "" {
		p.Name = name
	}
	if err := Validate(p); err != nil {
		return Project{}, fmt.Errorf("invalid project %q: %w", name, err)
	}
	return p, nil
}

// Save writes a project bookmark to disk, creating directories as needed.
func Save(paths Paths, p Project) error {
	if err := Validate(p); err != nil {
		return err
	}
	if err := paths.EnsureDirs(); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := toml.Marshal(p)
	if err != nil {
		return fmt.Errorf("encode project %q: %w", p.Name, err)
	}

	path := paths.ProjectFile(p.Name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write project %q: %w", p.Name, err)
	}
	return nil
}

// List returns sorted project names.
func List(paths Paths) ([]string, error) {
	if err := paths.EnsureDirs(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(paths.Projects)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".toml"))
	}
	return names, nil
}

// Remove deletes a project bookmark file.
func Remove(paths Paths, name string) error {
	path := paths.ProjectFile(name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("unknown project %q", name)
		}
		return fmt.Errorf("remove project %q: %w", name, err)
	}
	return nil
}

// Exists reports whether a project bookmark file exists.
func Exists(paths Paths, name string) bool {
	_, err := os.Stat(paths.ProjectFile(name))
	return err == nil
}

// ProjectPath returns the absolute path to a project file.
func ProjectPath(paths Paths, name string) string {
	p, _ := filepath.Abs(paths.ProjectFile(name))
	return p
}
