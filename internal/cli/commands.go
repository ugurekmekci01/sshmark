package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ugurekmekci01/sshmark/internal/config"
	"github.com/ugurekmekci01/sshmark/internal/ssh"
)

func (a *App) listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List project bookmarks",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := config.List(a.Paths)
			if err != nil {
				return err
			}
			sort.Strings(names)
			if len(names) == 0 {
				_, _ = fmt.Fprintln(a.Prompter.Out, "No projects configured.")
				return nil
			}
			for _, name := range names {
				p, err := config.Load(a.Paths, name)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(a.Prompter.Out, formatProjectSummary(p, false))
			}
			return nil
		},
	}
}

func (a *App) showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <project>",
		Short: "Show a project bookmark",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := config.Load(a.Paths, args[0])
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(a.Prompter.Out, formatProjectSummary(p, true))
			return nil
		},
	}
}

func (a *App) checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <project>",
		Short: "Validate a project without connecting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := config.Load(a.Paths, args[0])
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(a.Prompter.Out, "Project %q: config ok\n", p.Name)
			if err := checkPortsForProject(p.Name, p.Tunnels); err != nil {
				_, _ = fmt.Fprintln(a.Prompter.Out, err.Error())
				return err
			}
			if len(p.Tunnels) > 0 {
				_, _ = fmt.Fprintln(a.Prompter.Out, "Local ports: available")
			}

			sshPath, err := a.Runner.LookPath("ssh")
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(a.Prompter.Out, "OpenSSH: available")

			params := openParamsFromProject(p)
			sshArgs := ssh.RemoteCommandArgs(params, ssh.CheckRemoteDirScript(p.Directory))
			out, err := a.Runner.RunCapture(sshPath, sshArgs...)
			if err != nil {
				_, _ = fmt.Fprintf(a.Prompter.Out, "Remote directory: unreachable (%v)\n", err)
				return fmt.Errorf("check remote directory for project %q: %w", p.Name, err)
			}
			_, exists, parseErr := ssh.ParseRemoteDirOutput(out)
			if parseErr != nil {
				return parseErr
			}
			if exists {
				_, _ = fmt.Fprintln(a.Prompter.Out, "Remote directory: exists")
			} else {
				_, _ = fmt.Fprintln(a.Prompter.Out, "Remote directory: missing (will be created on open)")
			}
			return nil
		},
	}
}

func (a *App) editCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <project>",
		Short: "Edit a project bookmark in $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.Exists(a.Paths, args[0]) {
				return fmt.Errorf("unknown project %q", args[0])
			}
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			path := config.ProjectPath(a.Paths, args[0])
			parts := strings.Fields(editor)
			editArgs := append(parts[1:], path)
			return a.Runner.Run(parts[0], editArgs...)
		},
	}
}

func (a *App) removeCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "remove <project>",
		Short: "Delete a project bookmark",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.Exists(a.Paths, args[0]) {
				return fmt.Errorf("unknown project %q", args[0])
			}
			if !yes {
				ok, err := a.Prompter.confirm(fmt.Sprintf("Delete project %q?", args[0]))
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("remove cancelled")
				}
			}
			return config.Remove(a.Paths, args[0])
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	return cmd
}
