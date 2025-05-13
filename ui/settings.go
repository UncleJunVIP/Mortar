package ui

import (
	"fmt"
	gabamod "github.com/UncleJunVIP/gabagool/models"
	"github.com/UncleJunVIP/gabagool/ui"
	"mortar/models"
	"qlova.tech/sum"
)

type SettingsScreen struct {
}

func InitSettingsScreen() SettingsScreen {
	return SettingsScreen{}
}

func (s SettingsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Settings
}

func (s SettingsScreen) Draw() (settings interface{}, exitCode int, e error) {
	items := []ui.ItemWithOptions{
		{
			Item: gabamod.MenuItem{
				Text: "Download Art",
			},
			Options: []ui.Option{
				{DisplayName: "True", Value: true},
				{DisplayName: "False", Value: false},
			},
		},
		{
			Item: gabamod.MenuItem{
				Text: "Art Type",
			},
			Options: []ui.Option{
				{DisplayName: "Box Art", Value: "BOX_ART"},
				{DisplayName: "Title Screen", Value: "TITLE_SCREEN"},
				{DisplayName: "Logos", Value: "LOGOS"},
				{DisplayName: "Screenshots", Value: "SCREENSHOTS"},
			},
			SelectedOption: 1,
		},
	}

	footerHelpItems := []ui.FooterHelpItem{
		{ButtonName: "B", HelpText: "Cancel"},
		{ButtonName: "←→", HelpText: "Change option"},
		{ButtonName: "A", HelpText: "Confirm"},
	}

	result, err := ui.NewBlockingOptionsList(
		"Settings",
		items,
		footerHelpItems,
	)

	if err != nil {
		fmt.Println("Error showing options list:", err)
		return
	}

	return result, 0, nil
}
