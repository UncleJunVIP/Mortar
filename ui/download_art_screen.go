package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/ui"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"mortar/models"
	"mortar/utils"
	"qlova.tech/sum"
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
	footerHelpItems := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "I'll Find My Own"},
		{ButtonName: "A", HelpText: "Use It!"},
	}

	for _, game := range a.Games {
		process, _ := gaba.NewBlockingProcess(fmt.Sprintf("Downloading Art for %s...", game.DisplayName), func() (interface{}, error) {
			artPath := utils.FindArt(a.Platform, game, a.DownloadType)
			return artPath, nil
		})

		artPath := process.Result.(string)

		result, err := gaba.NewBlockingMessage("", "Found This Art!", footerHelpItems, artPath)
		if err != nil {
			return nil, -1, err
		}

		if result.IsSome() && result.Unwrap().ButtonName == "Yes" {
			// TODO keep art
		} else {
			// TODO delete art
		}
	}

	return nil, 2, nil
}
