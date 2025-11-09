package utils

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nikhiljohn10/uagplugin/logger"
)

// BuildAndLog is a wrapper around BuildPlugin that logs errors.
func BuildAndLog(ctx context.Context, pluginName, srcDir, buildFrom string) {
	buildDir, err := GetBuildDir()
	if err != nil {
		logger.Error("Failed to get plugin build dir: %v", err)
		return
	}
	err = BuildPlugin(ctx, pluginName, srcDir, buildDir, buildFrom == "repo")
	if err != nil {
		logger.Error("Failed to build plugin: %v", err)
	}
}

// BuildPlugin builds a Go plugin from source.
func BuildPlugin(ctx context.Context, pluginName, sourceDir, buildDir string, cleanup ...bool) error {
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

	projectRoot, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to locate project root: %v", err)
		return err
	}
	projectRoot, err = filepath.Abs(projectRoot)
	if err != nil {
		logger.Error("Failed to resolve project root path: %v", err)
		return err
	}

	if err := ensureLocalModule(ctx, sourceDir, projectRoot); err != nil {
		return err
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
	logger.Info("Done.\nPlugin Installed Location: %s", soFile)
	return nil
}

func ensureLocalModule(ctx context.Context, sourceDir, projectRoot string) error {
	modulePath := "github.com/nikhiljohn10/uagplugin"

	dropCmd := exec.CommandContext(ctx, "go", "mod", "edit", "-dropreplace="+modulePath)
	dropCmd.Dir = sourceDir
	_ = dropCmd.Run()

	moduleRoot, err := findModuleRoot(projectRoot, modulePath)
	if err != nil {
		logger.Error("Failed to inspect local module: %v", err)
		return err
	}

	if moduleRoot != "" {
		replaceArg := fmt.Sprintf("-replace=%s=%s", modulePath, moduleRoot)
		replaceCmd := exec.CommandContext(ctx, "go", "mod", "edit", replaceArg)
		replaceCmd.Dir = sourceDir
		if output, err := replaceCmd.CombinedOutput(); err != nil {
			logger.Error("Failed to align plugin module: %v", err)
			if len(output) > 0 {
				logger.Error("%s", string(output))
			}
			return err
		}
	} else if logger.IsDebugMode() {
		logger.Debug("Local module %q not found; relying on remote module", modulePath)
	}

	tidyCmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	tidyCmd.Dir = sourceDir
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		logger.Error("Failed to tidy plugin module: %v", err)
		return err
	}
	return nil
}

func findModuleRoot(startDir, modulePath string) (string, error) {
	dir := startDir
	for {
		modFile := filepath.Join(dir, "go.mod")
		contents, err := os.ReadFile(modFile)
		if err == nil {
			if extractModulePath(contents) == modulePath {
				return dir, nil
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", nil
}

func extractModulePath(content []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			if idx := strings.Index(line, "//"); idx >= 0 {
				line = strings.TrimSpace(line[:idx])
			}
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}
