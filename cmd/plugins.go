package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/nikhiljohn10/uagplugin/internal/utils"
	"github.com/nikhiljohn10/uagplugin/logger"
	"github.com/spf13/cobra"
)

func pluginInstall(cmd *cobra.Command, args []string) {
	pluginName, _ := cmd.Flags().GetString("name")
	token, _ := cmd.Flags().GetString("token")
	u, err := url.ParseRequestURI(args[0])
	if err != nil || u.Scheme == "" || u.Host == "" {
		logger.Error("Invalid repository URL")
		return
	}
	pluginRepoInstall(cmd.Context(), pluginName, token, u)
}

func pluginInstallDir(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		logger.Error("Please provide the directory path of the plugin to install.")
		return
	}

	if args[0] == "" {
		logger.Error("Directory path cannot be empty.")
		return
	}

	srcDir := args[0]
	if srcDir == "." || srcDir == "./" {
		cwd, err := os.Getwd()
		if err != nil {
			logger.Error("Failed to get current working directory: %v", err)
			return
		}
		srcDir = cwd
	}

	if srcDir == ".." || srcDir == "../" {
		cwd, err := os.Getwd()
		if err != nil {
			logger.Error("Failed to get current working directory: %v", err)
			return
		}
		srcDir = filepath.Dir(cwd)
	}

	pluginName, _ := cmd.Flags().GetString("name")
	pluginDirInstall(cmd.Context(), pluginName, srcDir)
}

func getBaseAndBuildDir() (baseDir, buildDir string, err error) {
	if !logger.IsDebugMode() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", "", err
		}
		baseDir = filepath.Join(homeDir, ".uag")
	} else {
		baseDir = ".uag"
	}
	pluginsDir := filepath.Join(baseDir, "plugins")
	buildDir = filepath.Join(pluginsDir, "build")
	return baseDir, buildDir, nil
}

func buildAndLog(ctx context.Context, pluginName, srcDir, buildDir, source string) {
	logger.Info("Installing plugin from %s: %s", source, srcDir)
	err := buildPlugin(ctx, pluginName, srcDir, buildDir)
	if err != nil {
		logger.Error("Failed to build plugin: %v", err)
		return
	}
	logger.Info("Plugin installed successfully from %s: %s", source, srcDir)
}

func pluginDirInstall(ctx context.Context, pluginName, srcDir string) {
	stat, err := os.Stat(srcDir)
	if err != nil || !stat.IsDir() {
		logger.Error("Invalid directory specified: %s", srcDir)
		return
	}
	if pluginName == "" {
		pluginName = strings.ToLower(filepath.Base(filepath.Clean(srcDir)))
		pluginName = strings.TrimPrefix(strings.TrimSuffix(pluginName, filepath.Ext(pluginName)), "uag-")
	}
	_, buildDir, err := getBaseAndBuildDir()
	if err != nil {
		logger.Error("Failed to get build directory: %v", err)
		return
	}
	buildAndLog(ctx, pluginName, srcDir, buildDir, "directory")
}

func pluginRepoInstall(ctx context.Context, pluginName, token string, repoURL *url.URL) {
	if repoURL.Host != "github.com" {
		logger.Error("Only GitHub repositories are supported.")
		return
	}

	parts := strings.Split(repoURL.Path, "/")
	if pluginName == "" {
		pluginName = parts[len(parts)-1]
	}
	baseDir, buildDir, err := getBaseAndBuildDir()
	if err != nil {
		logger.Error("Failed to get build directory: %v", err)
		return
	}
	pluginsDir := filepath.Join(baseDir, "plugins")
	srcDir := filepath.Join(pluginsDir, "pkgs", pluginName)

	if _, err := os.Stat(srcDir); err == nil {
		if err := os.RemoveAll(srcDir); err != nil {
			logger.Error("Failed to remove plugin directory: %v", err)
		}
	}

	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	owner := parts[len(parts)-2]
	apiURL := "https://api.github.com/repos/" + owner + "/" + pluginName
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, 6*time.Second)
		defer cancel()
	}
	isPublic := utils.IsRepoPublic(ctx, apiURL, token)

	cloneOptions := &git.CloneOptions{
		URL:      repoURL.String(),
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
	_, err = git.PlainClone(srcDir, false, cloneOptions)
	if err != nil {
		logger.Error("Failed to clone plugin: %v", err)
		return
	}

	buildAndLog(ctx, pluginName, srcDir, buildDir, "repo")
}

func buildPlugin(ctx context.Context, pluginName, sourceDir, buildDir string, cleanup ...bool) error {
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
		return fmt.Errorf("not a UAG plugin")
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
	// Enforce a build timeout to avoid indefinite hangs
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
	}
	cmdBuild := exec.CommandContext(ctx, "go", "build", "-buildmode=plugin", "-o", soFile, ".")
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
