package cmd

import (
	"github.com/nikhiljohn10/uagplugin/internal/version"
	"github.com/nikhiljohn10/uagplugin/logger"
	"github.com/nikhiljohn10/uagplugin/typing"
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
		logger.Info("UAG %s (commit %s, built %s) | contract %s (min %s)", version.Version, version.Commit, version.Date, typing.ContractVersion, typing.MinSupportedContractVersion)
	},
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a plugin from a GitHub repository",
	Args:  cobra.NoArgs,
	Run:   func(cmd *cobra.Command, args []string) { pluginInstall(cmd.Context(), cmd, args) },
}

func init() {
	Root.AddCommand(versionCmd)
	Root.Version = version.Version

	pluginInstallCmd.Flags().String("url", "", "GitHub repository URL")
	pluginInstallCmd.Flags().String("token", "", "GitHub Personal Access Token for cloning private repositories")
	pluginInstallCmd.Flags().String("dir", "", "Directory of the plugin to install (if installing from local directory)")
	pluginInstallCmd.Flags().String("name", "", "Name of the plugin")
	Root.AddCommand(pluginInstallCmd)
}
