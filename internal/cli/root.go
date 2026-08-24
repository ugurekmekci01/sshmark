package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ugurekmekci01/sshmark/internal/config"
)

var version = "dev"

// Execute runs the sshmark CLI.
func Execute() error {
	app := newApp()
	root := newRootCmd(app)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	return nil
}

func newRootCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:           "sshmark",
		Short:         "Open remote dev projects over SSH with one command",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVarP(&app.Verbose, "verbose", "v", false, "Verbose output")
	configDir := ""
	root.PersistentFlags().StringVar(&configDir, "config-dir", "", "Override sshmark config directory")
	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if configDir != "" {
			app.Paths = config.Paths{
				Base:     configDir,
				Projects: filepath.Join(configDir, "projects"),
			}
		}
	}

	root.AddCommand(
		app.addCmd(),
		app.openCmd(),
		app.listCmd(),
		app.showCmd(),
		app.checkCmd(),
		app.editCmd(),
		app.removeCmd(),
		newVersionCmd(),
	)

	root.RunE = func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	}

	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	}
}
