package ui

import (
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
		artPath := gaba.WithProcessMessage("Downloading Art", func() interface{} {
			return utils.FindArt(a.Platform, game, a.DownloadType)
		})

		result, err := gaba.NewBlockingMessage(game.DisplayName, "Found This Art!", footerHelpItems, artPath.(string))
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
