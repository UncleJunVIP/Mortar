package ui

import (
	"github.com/UncleJunVIP/gabagool/ui"
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
	footerHelpItems := []ui.FooterHelpItem{
		{ButtonName: "B", HelpText: "No"},
		{ButtonName: "A", HelpText: "Yes"},
	}

	for _, game := range a.Games {
		artPath := utils.FindArt(a.Platform, game, a.DownloadType)

		result, err := ui.NewBlockingMessage("Confirmation", "Are you sure you want to proceed?", footerHelpItems, artPath)
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
