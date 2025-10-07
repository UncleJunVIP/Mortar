package ui

import (
	"encoding/base64"
	"mortar/clients"
	"mortar/models"
	"mortar/state"
	"mortar/utils"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"qlova.tech/sum"
)

type DownloadScreen struct {
	Platform      models.Platform
	Games         shared.Items
	SelectedGames shared.Items
	SearchFilter  string
}

func InitDownloadScreen(platform models.Platform, games shared.Items, selectedGames shared.Items, searchFilter string) DownloadScreen {
	return DownloadScreen{
		Platform:      platform,
		Games:         games,
		SelectedGames: selectedGames,
		SearchFilter:  searchFilter,
	}
}

func (d DownloadScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Download
}

func (d DownloadScreen) Draw() (value interface{}, exitCode int, e error) {
	logger := gaba.GetLoggerInstance()

	downloads := BuildDownload(d.Platform, d.SelectedGames)

	headers := make(map[string]string)

	if d.Platform.Host.HostType == shared.HostTypes.ROMM {
		auth := d.Platform.Host.Username + ":" + d.Platform.Host.Password
		authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		headers["Authorization"] = authHeader

		logger.Debug("RomM Auth Header", "header", authHeader)
	}

	slices.SortFunc(downloads, func(a, b gaba.Download) int {
		return strings.Compare(strings.ToLower(a.DisplayName), strings.ToLower(b.DisplayName))
	})

	logger.Debug("Starting ROM download", "downloads", downloads)

	res, err := gaba.DownloadManager(downloads, headers, state.GetAppState().Config.DownloadArt)
	if err != nil {
		logger.Error("Error downloading", "error", err)
		return nil, -1, err
	}

	if len(res.FailedDownloads) > 0 {
		for _, g := range downloads {
			if slices.Contains(res.FailedDownloads, g) {
				common.DeleteFile(g.Location)
			}
		}
	}

	exitCode = 0

	if len(res.CompletedDownloads) == 0 {
		exitCode = 1
	}

	var downloadedGames []shared.Item

	for _, g := range d.Games {
		if slices.ContainsFunc(res.CompletedDownloads, func(d gaba.Download) bool {
			return d.DisplayName == g.DisplayName
		}) {
			downloadedGames = append(downloadedGames, g)
		}
	}

	return downloadedGames, exitCode, err
}

func BuildDownload(platform models.Platform, games shared.Items) []gaba.Download {
	var downloads []gaba.Download
	for _, g := range games {

		var downloadLocation string
		if utils.IsDev() {
			romDirectory := strings.ReplaceAll(platform.LocalDirectory, common.RomDirectory, utils.GetRomDirectory())
			downloadLocation = filepath.Join(romDirectory, g.Filename)
		} else {
			downloadLocation = filepath.Join(platform.LocalDirectory, g.Filename)
		}

		root := platform.Host.RootURI

		if platform.Host.Port != 0 {
			root = root + ":" + strconv.Itoa(platform.Host.Port)
		}

		var sourceURL string

		if platform.Host.HostType == shared.HostTypes.ROMM {
			client := clients.NewRomMClient(platform.Host.RootURI, platform.Host.Port, platform.Host.Username, platform.Host.Password)
			sourceURL, _ = client.BuildDownloadURL(g.RomID, g.Filename)
		} else {
			sourceURL, _ = url.JoinPath(root, platform.HostSubdirectory, g.Filename)
		}

		downloads = append(downloads, gaba.Download{
			URL:         sourceURL,
			Location:    downloadLocation,
			DisplayName: g.DisplayName,
		})
	}

	return downloads
}
