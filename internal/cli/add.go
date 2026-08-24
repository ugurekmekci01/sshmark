package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/ugurekmekci01/sshmark/internal/config"
	"github.com/ugurekmekci01/sshmark/internal/ssh"
)

func (a *App) addCmd() *cobra.Command {
	var (
		hostFlag    string
		addressFlag string
		userFlag    string
		dirFlag     string
		portFlag    int
		tunnelFlags []string
	)

	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Create a project bookmark",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			flagMode := hostFlag != "" && userFlag != ""

			if name == "" {
				var err error
				name, err = a.Prompter.readLine("Project name", "")
				if err != nil {
					return err
				}
			}

			host := hostFlag
			if host == "" {
				var err error
				host, err = a.Prompter.readLine("Machine (nickname or IP)", "")
				if err != nil {
					return err
				}
			}

			address := addressFlag
			nickname := host
			if looksLikeAddress(host) {
				address = host
				nickname = ssh.ProjectHostNickname(name)
				if !flagMode {
					var err error
					nickname, err = a.Prompter.readLine("Host nickname", nickname)
					if err != nil {
						return err
					}
				}
			} else if address == "" {
				known, err := config.IsKnownHost(a.Paths, nickname)
				if err != nil {
					return err
				}
				if !known {
					var err2 error
					address, err2 = a.Prompter.readLine("Machine address for "+nickname, "")
					if err2 != nil {
						return err2
					}
				}
			}

			user := userFlag
			if user == "" {
				var err error
				user, err = a.Prompter.readLine("SSH user", defaultUser())
				if err != nil {
					return err
				}
			}

			port := portFlag
			if !flagMode && !cmd.Flags().Changed("port") {
				portStr, err := a.Prompter.readLine("SSH port", "22")
				if err != nil {
					return err
				}
				parsed, err := strconv.Atoi(portStr)
				if err != nil || parsed < 1 || parsed > 65535 {
					return fmt.Errorf("invalid SSH port %q", portStr)
				}
				port = parsed
			}

			directory := dirFlag
			if directory == "" {
				if flagMode {
					directory = defaultRemoteDirectory(user)
				} else {
					var err error
					directory, err = a.Prompter.readLine("Remote directory", defaultRemoteDirectory(user))
					if err != nil {
						return err
					}
				}
			}

			var tunnels []config.Tunnel
			if len(tunnelFlags) > 0 {
				for _, tf := range tunnelFlags {
					tun, err := parseTunnelFlag(tf)
					if err != nil {
						return err
					}
					tunnels = append(tunnels, tun)
				}
			} else if !flagMode {
				ok, err := a.Prompter.confirm("Add a port tunnel?")
				if err != nil {
					return err
				}
				if ok {
					for {
						portStr, err := a.Prompter.readLine("Local port (REMOTE or LOCAL:REMOTE)", "")
						if err != nil {
							return err
						}
						tun, err := parseTunnelFlag(portStr)
						if err != nil {
							return err
						}
						tunnels = append(tunnels, tun)
						more, err := a.Prompter.confirm("Add another tunnel?")
						if err != nil {
							return err
						}
						if !more {
							break
						}
					}
				}
			}

			if address != "" {
				if err := a.ensureHostReady(nickname, address, user, port); err != nil {
					return err
				}
			}

			project := config.Project{
				Name:      name,
				Host:      nickname,
				User:      user,
				Port:      port,
				Directory: directory,
				Tunnels:   tunnels,
			}
			if config.Exists(a.Paths, name) {
				return fmt.Errorf("project %q already exists", name)
			}

			created, err := a.ensureRemoteDirectory(project)
			if err != nil {
				return err
			}
			a.reportRemoteDirectory(project, created)

			if err := config.Save(a.Paths, project); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(a.Prompter.Out, "Saved project %q\n", name)
			_, _ = fmt.Fprintf(a.Prompter.Out, "Open with: sshmark open %s\n", name)
			_, _ = fmt.Fprintln(a.Prompter.Out, pathInstallHint())
			return nil
		},
	}

	cmd.Flags().StringVar(&hostFlag, "host", "", "Machine nickname or address")
	cmd.Flags().StringVar(&addressFlag, "address", "", "Machine address when host is a nickname")
	cmd.Flags().StringVar(&userFlag, "user", "", "SSH user")
	cmd.Flags().StringVar(&dirFlag, "dir", "", "Remote project directory")
	cmd.Flags().IntVar(&portFlag, "port", 22, "SSH port")
	cmd.Flags().StringArrayVar(&tunnelFlags, "tunnel", nil, "Tunnel PORT or LOCAL:REMOTE (repeatable)")

	return cmd
}

func (a *App) openCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open <project>",
		Short: "Open a project over SSH",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := config.Load(a.Paths, args[0])
			if err != nil {
				return err
			}

			if err := checkPortsForProject(project.Name, project.Tunnels); err != nil {
				return err
			}

			created, err := a.ensureRemoteDirectory(project)
			if err != nil {
				return err
			}
			if created && a.Verbose {
				a.reportRemoteDirectory(project, true)
			}

			if a.Verbose {
				_, _ = fmt.Fprintln(os.Stdout, formatProjectSummary(project, true))
			} else {
				_, _ = fmt.Fprintln(os.Stdout, formatProjectSummary(project, false))
			}

			sshPath, err := a.Runner.LookPath("ssh")
			if err != nil {
				return err
			}

			argv := ssh.BuildOpenArgs(openParamsFromProject(project))
			if ssh.ContainsInsecureSSHFlags(argv) {
				return fmt.Errorf("refusing insecure SSH options")
			}

			if err := a.SSHRunner(sshPath, argv); err != nil {
				return fmt.Errorf("ssh connection failed for project %q: %w", project.Name, err)
			}
			return nil
		},
	}
	return cmd
}
