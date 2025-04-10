package utils

import (
	"context"
	"go.uber.org/zap"
	"mortar/clients"
	"mortar/common"
	"mortar/models"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const romDirectory = "/mnt/SDCARD/Roms"

var tagRegex = regexp.MustCompile(`\((.*?)\)`)

func FetchRomDirectories() (map[string]string, error) {
	dirs := make(map[string]string)

	entries, err := os.ReadDir(romDirectory)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			tag := tagRegex.FindStringSubmatch(entry.Name())
			if tag == nil {
				continue
			}

			path := filepath.Join(romDirectory, entry.Name())
			dirs[tag[1]] = path

		}
	}

	return dirs, nil
}

func DeleteFile(path string) {
	logger := common.GetLoggerInstance()

	err := os.Remove(path)
	if err != nil {
		logger.Error("Issue removing file",
			zap.String("path", path),
			zap.Error(err))
	} else {
		logger.Debug("Removed file", zap.String("path", path))
	}
}

func FindArt() bool {
	logger := common.GetLoggerInstance()
	appState := common.GetAppState()

	if appState.CurrentHost.HostType == models.HostTypes.ROMM {
		// Skip all this silliness and grab the art from RoMM
		client, err := clients.BuildClient(appState.CurrentHost)
		if err != nil {
			return false
		}

		var selectedItem models.Item

		for _, item := range appState.CurrentItemsList {
			if item.Filename == appState.SelectedFile {
				selectedItem = item
				break
			}
		}

		if selectedItem.ArtURL == "" {
			return false
		}

		slashIdx := strings.LastIndex(selectedItem.ArtURL, "/")
		artSubdirectory, artFilename := selectedItem.ArtURL[:slashIdx], selectedItem.ArtURL[slashIdx+1:]

		artFilename = strings.Split(artFilename, "?")[0] // For the query string caching stuff

		mediaPath := filepath.Join(appState.CurrentSection.LocalDirectory, ".media")

		err = client.DownloadFileRename(artSubdirectory,
			mediaPath, artFilename, appState.SelectedFile)

		if err != nil {
			return false
		}

		return true
	}

	tag := tagRegex.FindStringSubmatch(appState.CurrentSection.LocalDirectory)

	if tag == nil {
		return false
	}

	client := clients.NewThumbnailClient()
	section := client.BuildThumbnailSection(tag[1])

	artList, err := client.ListDirectory(section)

	if err != nil {
		logger.Info("Unable to fetch artlist", zap.Error(err))
		return false
	}

	noExtension := strings.TrimSuffix(appState.SelectedFile, filepath.Ext(appState.SelectedFile))

	var matched models.Item

	// naive search first
	for _, art := range artList {
		if strings.Contains(strings.ToLower(art.Filename), strings.ToLower(noExtension)) {
			matched = art
			break
		}
	}

	if matched.Filename == "" {
		// TODO Levenshtein Distance support at some point
	}

	if matched.Filename != "" {
		err = client.DownloadFileRename(section.HostSubdirectory,
			filepath.Join(appState.CurrentSection.LocalDirectory, ".media"), matched.Filename, appState.SelectedFile)

		if err != nil {
			return false
		}

		return true
	}

	return false
}

func DownloadFile(cancel context.CancelFunc) error {
	defer cancel()

	logger := common.GetLoggerInstance()
	appState := common.GetAppState()

	client, err := clients.BuildClient(appState.CurrentHost)
	if err != nil {
		return err
	}

	defer func(client models.Client) {
		err := client.Close()
		if err != nil {
			logger.Error("Unable to close client", zap.Error(err))
		}
	}(client)

	var hostSubdirectory string

	if appState.CurrentHost.HostType == models.HostTypes.ROMM {
		var selectedItem models.Item
		for _, item := range appState.CurrentItemsList {
			if item.Filename == appState.SelectedFile {
				selectedItem = item
				break
			}
		}
		hostSubdirectory = selectedItem.RomID
	} else {
		hostSubdirectory = appState.CurrentSection.HostSubdirectory
	}

	return client.DownloadFile(hostSubdirectory,
		appState.CurrentSection.LocalDirectory, appState.SelectedFile)
}
