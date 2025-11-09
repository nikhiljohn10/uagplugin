package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/nikhiljohn10/uagplugin/internal/utils"
	"github.com/nikhiljohn10/uagplugin/logger"
	"github.com/spf13/cobra"
)

type semVer struct {
	major, minor, patch int
}

func (s semVer) isGreaterThan(other semVer) bool {
	if s.major != other.major {
		return s.major > other.major
	}
	if s.minor != other.minor {
		return s.minor > other.minor
	}
	return s.patch > other.patch
}

func parseVersion(v string) (semVer, error) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return semVer{}, fmt.Errorf("invalid version format: %s", v)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semVer{}, fmt.Errorf("invalid major version: %s", parts[0])
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semVer{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semVer{}, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return semVer{major, minor, patch}, nil
}

func pluginInstall(cmd *cobra.Command, args []string) {
	pluginName, _ := cmd.Flags().GetString("name")
	token, _ := cmd.Flags().GetString("token")
	u, err := url.ParseRequestURI(args[0])
	if err != nil || u.Host == "" {
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
	if pluginName == "" {
		pluginName = filepath.Base(srcDir)
	}

	buildDir, err := utils.GetPluginBuildDir(pluginName)
	if err != nil {
		logger.Error("Failed to get plugin build directory: %v", err)
		return
	}

	err = os.RemoveAll(buildDir)
	if err != nil {
		logger.Error("Failed to remove existing build directory: %v", err)
		return
	}

	err = os.MkdirAll(buildDir, 0755)
	if err != nil {
		logger.Error("Failed to create build directory: %v", err)
		return
	}

	err = utils.CopyDir(srcDir, buildDir)
	if err != nil {
		logger.Error("Failed to copy plugin source to build directory: %v", err)
		return
	}

	utils.BuildAndLog(cmd.Context(), pluginName, buildDir, "dir")
}

func pluginRepoInstall(ctx context.Context, pluginName, token string, u *url.URL) {
	if u.Host != "github.com" {
		logger.Error("Only GitHub repositories are supported.")
		return
	}

	parts := strings.Split(u.Path, "/")
	if pluginName == "" {
		pluginName = parts[len(parts)-1]
	}
	baseDir, _, err := utils.GetBaseAndBuildDir()
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
		ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
	}
	isPublic := utils.IsRepoPublic(ctx, apiURL, token)

	cloneOptions := &git.CloneOptions{
		URL:      u.String(),
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
	repo, err := git.PlainClone(srcDir, false, cloneOptions)
	if err != nil {
		logger.Error("Failed to clone plugin: %v", err)
		return
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		logger.Error("Failed to get tags: %v", err)
		return
	}

	var latestVersion semVer
	var latestTag *plumbing.Reference
	var isLatestVersionSet bool
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tagName := strings.TrimPrefix(ref.Name().String(), "refs/tags/")
		v, err := parseVersion(tagName)
		if err != nil {
			return nil // Ignore invalid tags
		}
		if !isLatestVersionSet || v.isGreaterThan(latestVersion) {
			latestVersion = v
			latestTag = ref
			isLatestVersionSet = true
		}
		return nil
	})
	if err != nil {
		logger.Error("Failed to iterate over tags: %v", err)
		return
	}

	if latestTag != nil {
		w, err := repo.Worktree()
		if err != nil {
			logger.Error("Failed to get worktree: %v", err)
			return
		}
		err = w.Checkout(&git.CheckoutOptions{
			Hash: latestTag.Hash(),
		})
		if err != nil {
			logger.Error("Failed to checkout latest tag: %v", err)
			return
		}
		logger.Info("Checked out latest version: v%d.%d.%d", latestVersion.major, latestVersion.minor, latestVersion.patch)
	}

	utils.BuildAndLog(ctx, pluginName, srcDir, "repo")
}
