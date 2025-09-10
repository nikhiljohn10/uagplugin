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

func pluginDirInstall(cmd *cobra.Command, args []string) {
	var baseDir string
	srcDir, _ := cmd.Flags().GetString("dir")
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

	if srcDir == "" {
		logger.Error("No directory specified. Use --dir to provide a plugin directory.")
		return
	}
	stat, err := os.Stat(srcDir)
	if err != nil || !stat.IsDir() {
		logger.Error("Invalid directory specified: %s", srcDir)
		return
	}
	pluginName := filepath.Base(filepath.Clean(srcDir))
	logger.Info("Installing plugin from directory: %s", srcDir)
	err = buildPlugin(pluginName, srcDir, buildDir)
	if err != nil {
		logger.Error("Failed to build plugin: %v", err)
		return
	}
	logger.Info("Plugin installed successfully from directory: %s", srcDir)
}

func pluginInstall(cmd *cobra.Command, args []string) {
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

	repoURL := args[0]
	if !strings.HasPrefix(repoURL, "https://github.com") && !strings.HasPrefix(repoURL, "github.com") {
		logger.Error("Invalid repository URL")
		return
	}
	if strings.HasPrefix(repoURL, "github.com/") {
		repoURL = "https://" + repoURL
	}
	repoURL = strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(repoURL, "/")
	pluginName := parts[len(parts)-1]

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

	err = buildPlugin(pluginName, srcDir, buildDir)
	if err != nil {
		logger.Error("Failed to build plugin: %v", err)
		return
	}
	logger.Info("Plugin installed successfully from repo: %s", strings.TrimPrefix(repoURL, "https://"))
}

func buildPlugin(pluginName, sourceDir, buildDir string, cleanup ...bool) error {
	logger.Info("Building plugin as shared object file...")

	_, err1 := os.Stat(filepath.Join(sourceDir, "go.mod"))
	_, err2 := os.Stat(filepath.Join(sourceDir, "plugin.go"))
	if os.IsNotExist(err1) || os.IsNotExist(err2) {
		logger.Error("This repository is not a UAG plugin.")
		if len(cleanup) > 0 && cleanup[0] {
			if err := os.RemoveAll(sourceDir); err != nil {
				logger.Error("Failed to remove plugin directory: %v", err)
				return err
			}
		}
	}
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
	logger.Info("Done.\nPlugin Installed Location: %s", buildDir)
	return nil
}
