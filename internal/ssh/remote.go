package ssh

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	remoteDirExists  = "exists"
	remoteDirCreated = "created"
)

// ProjectHostNickname returns the default SSH config alias for a project.
func ProjectHostNickname(projectName string) string {
	return projectName + "-box"
}

// EnsureRemoteDirScript returns a shell script that mkdirs if needed and prints status.
func EnsureRemoteDirScript(directory string) string {
	q := ShellQuoteSingle(directory)
	return fmt.Sprintf(`if [ -d %s ]; then echo %s; else mkdir -p %s && echo %s; fi`, q, remoteDirExists, q, remoteDirCreated)
}

// CheckRemoteDirScript returns a shell script that prints exists or missing.
func CheckRemoteDirScript(directory string) string {
	q := ShellQuoteSingle(directory)
	return fmt.Sprintf(`if [ -d %s ]; then echo %s; else echo missing; fi`, q, remoteDirExists)
}

// RemoteCommandArgs builds ssh argv for a remote shell command.
func RemoteCommandArgs(p OpenParams, script string) []string {
	args := []string{}
	if p.Port > 0 && p.Port != 22 {
		args = append(args, "-p", strconv.Itoa(p.Port))
	}
	args = append(args, "-o", "BatchMode=yes", fmt.Sprintf("%s@%s", p.User, p.Host), script)
	return args
}

// ParseRemoteDirOutput interprets ensure/check remote directory script output.
func ParseRemoteDirOutput(output string) (created bool, exists bool, err error) {
	line := strings.TrimSpace(output)
	switch line {
	case remoteDirExists:
		return false, true, nil
	case remoteDirCreated:
		return true, true, nil
	case "missing":
		return false, false, nil
	case "":
		return false, false, fmt.Errorf("empty response from remote host")
	default:
		return false, false, fmt.Errorf("unexpected remote response: %s", line)
	}
}

// CopyIDAlreadyInstalled reports whether ssh-copy-id output means keys are already present.
func CopyIDAlreadyInstalled(output string) bool {
	lower := strings.ToLower(output)
	return strings.Contains(lower, "already exist") ||
		strings.Contains(lower, "were skipped because they are already installed")
}
