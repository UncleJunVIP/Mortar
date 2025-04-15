package utils

import (
	"context"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/disintegration/imaging"
	"go.uber.org/zap"
	"mortar/clients"
	"mortar/state"
	"path/filepath"
	"strings"
)

func MapTagsToDirectories(items shared.Items) map[string]string {
	mapping := make(map[string]string)

	for _, entry := range items {
		if entry.IsDirectory {

			path := filepath.Join(common.RomDirectory, entry.Filename)
			mapping[entry.Tag] = path

		}
	}

	return mapping
}

func FindArt() bool {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	if appState.CurrentHost.HostType == shared.HostTypes.ROMM {
		// Skip all this silliness and grab the art from RoMM
		client, err := clients.BuildClient(appState.CurrentHost)
		if err != nil {
			return false
		}

		var selectedItem shared.Item

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

		LastSavedArtPath, err := client.DownloadFileRename(artSubdirectory,
			mediaPath, artFilename, appState.SelectedFile)

		appState.LastSavedArtPath = LastSavedArtPath

		state.UpdateAppState(appState)

		if err != nil {
			return false
		}

		return true
	}

	tag := common.TagRegex.FindStringSubmatch(appState.CurrentSection.LocalDirectory)

	if tag == nil {
		return false
	}

	client := common.NewThumbnailClient(appState.Config.ArtDownloadType)
	section := client.BuildThumbnailSection(tag[1])

	artList, err := client.ListDirectory(section)

	if err != nil {
		logger.Info("Unable to fetch artlist", zap.Error(err))
		return false
	}

	noExtension := strings.TrimSuffix(appState.SelectedFile, filepath.Ext(appState.SelectedFile))

	var matched shared.Item

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
		lastSavedArtPath, err := client.DownloadFileRename(section.HostSubdirectory,
			filepath.Join(appState.CurrentSection.LocalDirectory, ".media"), matched.Filename, appState.SelectedFile)

		if err != nil {
			return false
		}

		src, err := imaging.Open(lastSavedArtPath)
		if err != nil {
			logger.Error("Unable to open last saved art", zap.Error(err))
			return false
		}

		dst := imaging.Resize(src, 400, 0, imaging.Lanczos)

		err = imaging.Save(dst, lastSavedArtPath)
		if err != nil {
			logger.Error("Unable to save resized last saved art", zap.Error(err))
			return false
		}

		appState.LastSavedArtPath = lastSavedArtPath

		state.UpdateAppState(appState)

		return true
	}

	return false
}

func DownloadFile(cancel context.CancelFunc) (string, error) {
	defer cancel()

	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	client, err := clients.BuildClient(appState.CurrentHost)
	if err != nil {
		return "", err
	}

	defer func(client shared.Client) {
		err := client.Close()
		if err != nil {
			logger.Error("Unable to close client", zap.Error(err))
		}
	}(client)

	var hostSubdirectory string

	if appState.CurrentHost.HostType == shared.HostTypes.ROMM {
		var selectedItem shared.Item
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
