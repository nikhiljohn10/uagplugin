package cmd

import (
	"github.com/nikhiljohn10/uagplugin/internal/version"
	"github.com/nikhiljohn10/uagplugin/logger"
	"github.com/nikhiljohn10/uagplugin/typing"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "uagplugin [.so file]",
	Short: "UAG Plugin Tool is a cli application to manage plugins",
	Long: `UAG Plugin Tool is a CLI application used to manage plugins that includes installing, testing and updating plugins.
This application provides various commands to interact with your system.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			logger.Info("Welcome to UAG Plugin Tool! Use --help for more information.")
			return
		}

		// If an argument is provided, show metadata of the .so file provided
		showMetadata(args[0])
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
	Use:   "install [repository url]",
	Short: "Install a plugin from a GitHub repository",
	Args:  cobra.ExactArgs(1),
	Run:   pluginInstall,
}

var pluginInstallDirCmd = &cobra.Command{
	Use:   "dir [source directory]",
	Short: "Install a plugin from a local directory",
	Args:  cobra.ExactArgs(1),
	Run:   pluginInstallDir,
}

var testCmd = &cobra.Command{
	Use:   "test [plugin .so file or directory]",
	Short: "Test installed plugins (.so) and optionally source tests",
	Long:  "Run smoke tests against compiled plugins in .uag/plugins/build and optionally run 'go test' in source directories.",
	Args:  cobra.MaximumNArgs(1),
	Run:   testPlugins,
}

func init() {
	Root.AddCommand(versionCmd)
	Root.Version = version.Version

	pluginInstallDirCmd.Flags().String("name", "", "Name of the plugin")
	pluginInstallCmd.AddCommand(pluginInstallDirCmd)

	pluginInstallCmd.Flags().String("token", "", "GitHub Personal Access Token for cloning private repositories")
	Root.AddCommand(pluginInstallCmd)

	testCmd.Flags().Int("timeout", 5, "Per-call timeout in seconds")
	testCmd.Flags().String("env-file", "", "Optional .env file to load before testing")
	testCmd.Flags().String("auth", "", "JSON object for AuthCredentials passed to plugin functions")
	testCmd.Flags().String("params", "", "JSON object for Params passed to plugin functions")
	testCmd.Flags().String("mode", "smoke", "Test mode: smoke|source|all")
	testCmd.Flags().Bool("json", false, "Output JSON report")
	Root.AddCommand(testCmd)
}
