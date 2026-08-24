package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// IsKnownHost reports whether a host nickname was already bootstrapped.
func IsKnownHost(paths Paths, host string) (bool, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return false, nil
	}

	f, err := os.Open(paths.KnownHostsFile())
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == host {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// MarkKnownHost records a bootstrapped host nickname.
func MarkKnownHost(paths Paths, host string) error {
	host = strings.TrimSpace(host)
	if host == "" {
		return fmt.Errorf("host nickname is required")
	}

	known, err := IsKnownHost(paths, host)
	if err != nil {
		return err
	}
	if known {
		return nil
	}

	if err := paths.EnsureDirs(); err != nil {
		return err
	}

	f, err := os.OpenFile(paths.KnownHostsFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("update known hosts: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, host); err != nil {
		return fmt.Errorf("write known host: %w", err)
	}
	return nil
}
