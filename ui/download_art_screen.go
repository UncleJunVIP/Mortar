package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/ui"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"mortar/models"
	"mortar/utils"
	"qlova.tech/sum"
	"time"
)

type DownloadArtScreen struct {
	Platform     models.Platform
	Games        shared.Items
	DownloadType sum.Int[shared.ArtDownloadType]
	SearchFilter string
}

func InitDownloadArtScreen(platform models.Platform, games shared.Items, downloadType sum.Int[shared.ArtDownloadType], searchFilter string) models.Screen {
	return DownloadArtScreen{
		Platform:     platform,
		Games:        games,
		DownloadType: downloadType,
		SearchFilter: searchFilter,
	}
}

func (a DownloadArtScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.DownloadArt
}

func (a DownloadArtScreen) Draw() (value interface{}, exitCode int, e error) {
	footerHelpItems := []ui.FooterHelpItem{
		{ButtonName: "B", HelpText: "I'll Find My Own"},
		{ButtonName: "A", HelpText: "Use It!"},
	}

	for _, game := range a.Games {
		process, _ := ui.BlockingProcess(fmt.Sprintf("Finding art for %s...", game.DisplayName), false, func() (interface{}, error) {
			artPath := utils.FindArt(a.Platform, game, a.DownloadType)
			return artPath, nil
		})

		artPath := process.Result.(string)
		if artPath == "" {
			_, _ = ui.BlockingProcess(fmt.Sprintf("No art found for %s!", game.DisplayName), false, func() (interface{}, error) {
				time.Sleep(time.Millisecond * 1500)
				return nil, nil
			})
			continue
		}

		result, err := ui.Message("", "Found This Art!", footerHelpItems, artPath)
		if err != nil {
			return nil, -1, err
		}

		if result.IsNone() {
			common.DeleteFile(artPath)
		}
	}

	return nil, 2, nil
}
