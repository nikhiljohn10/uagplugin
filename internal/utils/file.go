package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// GetPluginBuildDir returns the directory where a specific plugin will be built.
// func GetPluginBuildDir(pluginName string) (string, error) {
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get user home directory: %w", err)
// 	}
// 	return filepath.Join(homeDir, ".uag", "plugins", "build", pluginName), nil
// }

// CopyFile copies a single file from src to dst.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Ignore if source file doesn't exist (like go.sum)
		}
		return err
	}
	defer sourceFile.Close()

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// CopyDir recursively copies a directory from src to dst.
func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	directory, err := os.Open(src)
	if err != nil {
		return err
	}
	defer directory.Close()

	objects, err := directory.Readdir(-1)
	if err != nil {
		return err
	}

	for _, obj := range objects {
		srcFile := filepath.Join(src, obj.Name())
		dstFile := filepath.Join(dst, obj.Name())

		if obj.IsDir() {
			err = CopyDir(srcFile, dstFile)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(srcFile, dstFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetBaseDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".uag", "plugins"), nil
}

func GetBuildDir() (string, error) {
	baseDir, err := GetBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, "build"), nil
}

func GetBaseAndBuildDir() (string, string, error) {
	baseDir, err := GetBaseDir()
	if err != nil {
		return "", "", err
	}
	return baseDir, filepath.Join(baseDir, "build"), nil
}
