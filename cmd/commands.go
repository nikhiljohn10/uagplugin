package cmd

import (
	"github.com/nikhiljohn10/uagplugin/internal/version"
	"github.com/nikhiljohn10/uagplugin/logger"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "uagplugin",
	Short: "UAG Plugin Tool is a cli application to manage plugins",
	Long: `UAG Plugin Tool is a CLI application used to manage plugins that includes installing, testing and updating plugins.
This application provides various commands to interact with your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Welcome to UAG Plugin Tool! Use --help for more information.")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of UAG.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("UAG %s (commit %s, built %s)", version.Version, version.Commit, version.Date)
	},
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install [repo-url]",
	Short: "Install a plugin from a GitHub repository",
	Args:  cobra.ExactArgs(1),
	Run:   pluginInstall,
}

var pluginDirInstallCmd = &cobra.Command{
	Use:   "installDir [plugin-dir]",
	Short: "Install a plugin from a directory",
	Args:  cobra.ExactArgs(1),
	Run:   pluginDirInstall,
}

func init() {
	Root.AddCommand(versionCmd)
	Root.Version = version.Version

	pluginInstallCmd.Flags().String("token", "", "GitHub Personal Access Token for cloning private repositories")
	Root.AddCommand(pluginInstallCmd)

	pluginDirInstallCmd.Flags().String("name", "", "Name of the plugin")
	Root.AddCommand(pluginDirInstallCmd)
}
