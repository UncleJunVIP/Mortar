package ui

import (
	"github.com/UncleJunVIP/gabagool/ui"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"mortar/models"
	"qlova.tech/sum"
)

type DownloadArtScreen struct {
	Platform     models.Platform
	Game         shared.Item
	DownloadType sum.Int[shared.ArtDownloadType]
	SearchFilter string
}

func InitDownloadArtScreen(platform models.Platform, game shared.Item,
	downloadType sum.Int[shared.ArtDownloadType], searchFilter string) DownloadArtScreen {
	return DownloadArtScreen{
		Platform:     platform,
		Game:         game,
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

	result, err := ui.NewBlockingMessage("Confirmation", "Are you sure you want to proceed?", footerHelpItems, "path/to/image.bmp")
	if err != nil {
		return nil, -1, err
	}

	if result.IsSome() && result.Unwrap().ButtonName == "Yes" {
		return a.Game, 0, nil
	}

	return nil, 2, nil
}
