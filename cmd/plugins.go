package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/nikhiljohn10/uagplugin/internal/utils"
	"github.com/nikhiljohn10/uagplugin/logger"
	"github.com/spf13/cobra"
)

func pluginInstall(cmd *cobra.Command, args []string) {
	repoURL := args[0]
	if !strings.HasPrefix(repoURL, "https://github.com") && !strings.HasPrefix(repoURL, "github.com") {
		logger.Error("Invalid repository URL")
		return
	}
	if strings.HasPrefix(repoURL, "github.com/") {
		repoURL = "https://" + repoURL
	}
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	pluginName := parts[len(parts)-1]

	// Determine base directory based on environment
	var baseDir string
	if !logger.IsDebugMode() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Error("Failed to get user home directory: %v", err)
			return
		}
		baseDir = filepath.Join(homeDir, ".uag")
	} else {
		baseDir = ".uag"
	}

	pluginsDir := filepath.Join(baseDir, "plugins")
	buildDir := filepath.Join(pluginsDir, "build")
	srcDir := filepath.Join(pluginsDir, "pkgs", pluginName)

	logger.Info("Installing plugin: %s", pluginName)
	if _, err := os.Stat(srcDir); err == nil {
		if err := os.RemoveAll(srcDir); err != nil {
			logger.Error("Failed to remove plugin directory: %v", err)
		}
	}

	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	owner := parts[len(parts)-2]
	apiURL := "https://api.github.com/repos/" + owner + "/" + pluginName
	isPublic := utils.IsRepoPublic(apiURL, token)

	cloneOptions := &git.CloneOptions{
		URL:      repoURL,
		Progress: nil,
	}
	if !isPublic {
		if token == "" {
			logger.Warn("No token provided, cannot clone private repository.")
			return
		}
		auth := &http.BasicAuth{Username: "access_token", Password: token}
		cloneOptions.Auth = auth
	}
	_, err := git.PlainClone(srcDir, false, cloneOptions)
	if err != nil {
		logger.Error("Failed to clone plugin: %v", err)
		return
	}

	goModFile := filepath.Join(srcDir, "go.mod")
	if _, err := os.Stat(goModFile); os.IsNotExist(err) {
		logger.Error("This repository is not a UAG plugin.")
		if err := os.RemoveAll(srcDir); err != nil {
			logger.Error("Failed to remove plugin directory: %v", err)
		}
		return
	}

	logger.Info("Building plugin as shared object file...")
	err = buildPlugin(pluginName, srcDir, buildDir)
	if err != nil {
		logger.Error("Failed to build plugin: %v", err)
		return
	}
	logger.Info("Plugin installed successfully.")
	logger.Info("Installed Location: %s", buildDir)
}

func buildPlugin(pluginName, sourceDir, buildDir string) error {
	soFile, err := filepath.Abs(filepath.Join(buildDir, pluginName+".so"))
	if err != nil {
		logger.Error("Failed to resolve absolute path for plugin .so file: %v", err)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(soFile), 0755); err != nil {
		logger.Error("Failed to create compiled plugins directory: %v", err)
		return err
	}
	cmdBuild := exec.Command("go", "build", "-buildmode=plugin", "-o", soFile, ".")
	cmdBuild.Dir = sourceDir
	cmdBuild.Stdout = os.Stdout
	cmdBuild.Stderr = os.Stderr
	if err := cmdBuild.Run(); err != nil {
		logger.Error("Failed to build plugin as .so file: %v", err)
		return err
	}
	return nil
}
