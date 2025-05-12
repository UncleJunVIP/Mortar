package ui

import (
	"encoding/base64"
	gabamod "github.com/UncleJunVIP/gabagool/models"
	"github.com/UncleJunVIP/gabagool/ui"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"mortar/models"
	"mortar/utils"
	"net/url"
	"os"
	"path/filepath"
	"qlova.tech/sum"
	"strconv"
	"strings"
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
	logger := common.GetLoggerInstance()

	downloads := BuildDownload(d.Platform, d.SelectedGames)

	headers := make(map[string]string)

	if d.Platform.Host.HostType == shared.HostTypes.ROMM {
		auth := d.Platform.Host.Username + ":" + d.Platform.Host.Password
		authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		headers["Authorization"] = authHeader
	}

	_, err := ui.NewBlockingDownload(downloads, headers)
	if err != nil {
		logger.Error("Error downloading", zap.Error(err))
		return shared.Item{}, -1, err
	}

	// TODO finish this

	return nil, 0, err
}

func BuildDownload(platform models.Platform, games shared.Items) []gabamod.Download {
	var downloads []gabamod.Download
	for _, g := range games {

		var downloadLocation string
		if os.Getenv("DEVELOPMENT") == "true" {
			romDirectory := strings.ReplaceAll(platform.LocalDirectory, common.RomDirectory, utils.GetRomDirectory())
			downloadLocation = filepath.Join(romDirectory, g.Filename)
		} else {
			downloadLocation = filepath.Join(platform.LocalDirectory, g.Filename)
		}

		root := platform.Host.RootURI

		if platform.Host.Port != 0 {
			root = root + ":" + strconv.Itoa(platform.Host.Port)
		}

		sourceURL, _ := url.JoinPath(root, platform.HostSubdirectory, g.Filename)
		downloads = append(downloads, gabamod.Download{
			URL:         sourceURL,
			Location:    downloadLocation,
			DisplayName: g.DisplayName,
		})
	}

	return downloads
}
