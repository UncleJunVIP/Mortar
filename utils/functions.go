package utils

import (
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
)

func GetRomDirectory() string {
	if os.Getenv("DEVELOPMENT") == "true" {
		return "/Users/btk/Desktop/Roms"
	}

	return common.RomDirectory
}

func MapTagsToDirectories(items shared.Items) map[string]string {
	mapping := make(map[string]string)

	for _, entry := range items {
		if entry.IsDirectory {
			tag := strings.ReplaceAll(entry.Tag, "(", "")
			tag = strings.ReplaceAll(tag, ")", "")
			path := filepath.Join(common.RomDirectory, entry.Filename)
			mapping[tag] = path
		}
	}

	return mapping
}

func CachedMegaThreadJsonFilename(hostName, platformName string) string {
	return strings.ReplaceAll(fmt.Sprintf("%s_%s_%s.json",
		hostName, platformName, "megathread"), " ", "")
}

func CacheFolderExists() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	cachePath := filepath.Join(cwd, ".cache")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func DeleteCache() error {
	logger := common.GetLoggerInstance()
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(cwd, ".cache"))
	if err != nil {
		logger.Error("Unable to delete cache", zap.Error(err))
		return err
	}

	logger.Info("Cache deleted")
	return nil
}
