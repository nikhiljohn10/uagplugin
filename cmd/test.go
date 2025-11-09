package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/nikhiljohn10/uagplugin/internal/plugintest"
	"github.com/nikhiljohn10/uagplugin/logger"
	"github.com/nikhiljohn10/uagplugin/models"
	"github.com/spf13/cobra"
)

// showMetadata shows metadata of a .so plugin file
var showMetadata = func(filePath string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		logger.Error("Failed to resolve absolute path: %v", err)
		return
	}

	st, err := os.Stat(absPath)
	if err != nil {
		logger.Error("File not found: %s", absPath)
		return
	}

	var pluginList []string

	if st.IsDir() {
		entries, err := os.ReadDir(absPath)
		if err != nil {
			logger.Error("Failed to read directory: %v", err)
			return
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.ToLower(entry.Name()) != ".so" && strings.HasSuffix(strings.ToLower(entry.Name()), ".so") {
				pluginList = append(pluginList, filepath.Join(absPath, entry.Name()))
			}
		}
		if len(pluginList) == 0 {
			logger.Info("No .so plugin files found in directory: %s", absPath)
			return
		}
	} else {
		if !strings.HasSuffix(strings.ToLower(absPath), ".so") {
			logger.Info("Provided path is not a .so file: %s", absPath)
			return
		}
		pluginList = []string{absPath}
	}

	if len(pluginList) > 0 {
		println("Plugins:")
	}
	for _, pluginPath := range pluginList {
		meta, err := plugintest.GetPluginMetadata(pluginPath)
		if err != nil {
			logger.Error("Failed to get metadata for %s: %v", pluginPath, err)
			continue
		}

		println(fmt.Sprintf("  Plugin: %s", pluginPath))
		println(fmt.Sprintf("    Name:        %s", meta.Name))
		println(fmt.Sprintf("    Version:     %s", meta.Version))
		println(fmt.Sprintf("    Author:      %s", meta.Author))
		println(fmt.Sprintf("    Description: %s", meta.Description))
		println(fmt.Sprintf("    Contract:    %s", meta.ContractVersion))
	}
}

var testPlugins = func(cmd *cobra.Command, args []string) {
	// Ensure test env markers
	_ = os.Setenv("UAG_ENV", "test")
	_ = os.Setenv("UAG_TEST", "1")

	// Load env-file if provided
	envFile, _ := cmd.Flags().GetString("env-file")
	if strings.TrimSpace(envFile) != "" {
		if err := godotenv.Load(envFile); err != nil {
			logger.Warn("Failed to load env file %s: %v", envFile, err)
		}
	}

	// Resolve dirs
	baseDir, buildDir, err := getBaseAndBuildDir()
	if err != nil {
		logger.Fatal("Failed to resolve plugin directories: %v", err)
		return
	}

	// Flags
	timeoutSec, _ := cmd.Flags().GetInt("timeout")
	mode, _ := cmd.Flags().GetString("mode")
	jsonOut, _ := cmd.Flags().GetBool("json")

	// Parse auth/params
	var auth models.AuthCredentials = models.AuthCredentials{}
	if s, _ := cmd.Flags().GetString("auth"); strings.TrimSpace(s) != "" {
		if err := json.Unmarshal([]byte(s), &auth); err != nil {
			logger.Warn("Invalid --auth JSON: %v", err)
		}
	}
	var contact_params models.ContactQueryParams
	if s, _ := cmd.Flags().GetString("contact_params"); strings.TrimSpace(s) != "" {
		if err := json.Unmarshal([]byte(s), &contact_params); err != nil {
			logger.Warn("Invalid --contact_params JSON: %v", err)
		}
	}
	var ledger_params models.LedgerQueryParams
	if s, _ := cmd.Flags().GetString("ledger_params"); strings.TrimSpace(s) != "" {
		if err := json.Unmarshal([]byte(s), &ledger_params); err != nil {
			logger.Warn("Invalid --ledger_params JSON: %v", err)
		}
	}

	timeout := time.Duration(timeoutSec) * time.Second

	// Resolve target files/dirs
	var files []string
	var searchDirs []string
	if len(args) == 1 {
		p := strings.TrimSpace(args[0])
		if p == "" {
			logger.Fatal("invalid path argument")
			return
		}
		abs, _ := filepath.Abs(p)
		st, err := os.Stat(abs)
		if err != nil {
			logger.Fatal("path not found: %s", p)
			return
		}
		if st.IsDir() {
			searchDirs = []string{abs}
		} else {
			if !strings.HasSuffix(strings.ToLower(abs), ".so") {
				logger.Fatal("file must be a .so shared object: %s", abs)
				return
			}
			files = []string{abs}
		}
	} else {
		// No args: search default locations
		home, herr := os.UserHomeDir()
		if herr != nil || strings.TrimSpace(home) == "" {
			logger.Warn("could not resolve home dir: %v", herr)
		}
		defaults := []string{filepath.Join(".uag", "plugins", "build")}
		if home != "" {
			defaults = append(defaults, filepath.Join(home, ".uag", "plugins", "build"))
		}
		// prefer cwd ./.uag if exists else fallback to env-based buildDir
		var foundDirs []string
		for _, d := range defaults {
			if st, err := os.Stat(d); err == nil && st.IsDir() {
				foundDirs = append(foundDirs, d)
			}
		}
		if len(foundDirs) == 0 {
			// fall back to resolved buildDir from getBaseAndBuildDir
			if st, err := os.Stat(buildDir); err == nil && st.IsDir() {
				foundDirs = append(foundDirs, buildDir)
			}
		}
		if len(foundDirs) == 0 {
			logger.Info("No default plugin build directories found. Checked: %s, %s", filepath.Join(".uag", "plugins", "build"), fmt.Sprintf("%s/.uag/plugins/build", home))
		}
		searchDirs = foundDirs
	}

	// Run
	res := plugintest.Run(cmd.Context(), plugintest.RunConfig{
		BaseDir:       baseDir,
		BuildDir:      buildDir,
		Files:         files,
		SearchDirs:    searchDirs,
		Timeout:       timeout,
		Mode:          plugintest.ModeFromString(mode),
		Auth:          auth,
		ContactParams: contact_params,
		LedgerParams:  ledger_params,
		JSON:          jsonOut,
	})

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return
	}

	plugintest.PrintHuman(res)
	if res.Failures > 0 {
		os.Exit(1)
	}
}
