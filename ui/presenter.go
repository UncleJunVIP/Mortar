package ui

import (
	"context"
	"go.uber.org/zap"
	"mortar/clients"
	"mortar/models"
	"mortar/utils"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func ShowMessage(message string, timeout string) {
	args := []string{"--message", message, "--timeout", timeout}
	cmd := exec.Command("minui-presenter", args...)
	err := cmd.Run()

	if err != nil && cmd.ProcessState.ExitCode() != 124 {
		utils.LogStandardFatal("Failed to run minui-presenter", err)
	}
}

func findArt() bool {
	logger := utils.GetLoggerInstance()
	appState := utils.GetAppState()

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

		err = client.DownloadFileRename(artSubdirectory,
			filepath.Join(appState.CurrentSection.LocalDirectory, ".media"), artFilename, appState.SelectedFile)

		if err != nil {
			return false
		}

		return true
	}

	re := regexp.MustCompile(`\((.*?)\)`)
	tag := re.FindStringSubmatch(appState.CurrentSection.LocalDirectory)

	if len(tag) < 2 {
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

func downloadFile(cancel context.CancelFunc) error {
	defer cancel()

	logger := utils.GetLoggerInstance()
	appState := utils.GetAppState()

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
