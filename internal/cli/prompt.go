package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Prompter reads interactive input.
type Prompter struct {
	In  io.Reader
	Out io.Writer
}

func (p Prompter) readLine(prompt, defaultVal string) (string, error) {
	if defaultVal != "" {
		_, _ = fmt.Fprintf(p.Out, "%s [%s]: ", prompt, defaultVal)
	} else {
		_, _ = fmt.Fprintf(p.Out, "%s: ", prompt)
	}
	reader := bufio.NewReader(p.In)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal, nil
	}
	return line, nil
}

func (p Prompter) confirm(prompt string) (bool, error) {
	answer, err := p.readLine(prompt+" (y/N)", "")
	if err != nil {
		return false, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

func defaultPrompter() Prompter {
	return Prompter{In: os.Stdin, Out: os.Stdout}
}

func defaultUser() string {
	u := os.Getenv("USER")
	if u == "" {
		return "user"
	}
	return u
}

func defaultSSHConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ssh/config"
	}
	return home + "/.ssh/config"
}

func defaultSSHDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ssh"
	}
	return home + "/.ssh"
}
