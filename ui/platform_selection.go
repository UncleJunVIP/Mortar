package ui

import (
	"fmt"
	gabamod "github.com/UncleJunVIP/gabagool/models"
	gaba "github.com/UncleJunVIP/gabagool/ui"
	"mortar/models"
	"qlova.tech/sum"
)

type PlatformSelection struct {
	Host       models.Host
	QuitOnBack bool
}

func InitPlatformSelection(host models.Host, quitOnBack bool) PlatformSelection {
	return PlatformSelection{
		Host:       host,
		QuitOnBack: quitOnBack,
	}
}

func (ps PlatformSelection) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlatformSelection
}

func (ps PlatformSelection) Draw() (p interface{}, exitCode int, e error) {
	if len(ps.Host.Platforms) == 0 {
		return models.Platform{}, 404, nil
	}

	var menuItems []gabamod.MenuItem
	for _, platform := range ps.Host.Platforms {
		platform.Host = ps.Host
		menuItems = append(menuItems, gabamod.MenuItem{
			Text:     platform.Name,
			Selected: false,
			Focused:  false,
			Metadata: platform,
		})
	}

	var fhi []gaba.FooterHelpItem

	if ps.QuitOnBack {
		fhi = []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
			{ButtonName: "X", HelpText: "Settings"},
			{ButtonName: "A", HelpText: "Select"},
		}
	} else {
		fhi = []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Back"},
			{ButtonName: "A", HelpText: "Select"},
		}
	}

	title := ps.Host.DisplayName

	if ps.QuitOnBack {
		title = fmt.Sprintf("Mortar | %s", ps.Host.DisplayName)
	}

	selection, err := gaba.List(title, menuItems,
		gaba.ListOptions{
			FooterHelpItems:   fhi,
			EnableAction:      false,
			EnableMultiSelect: false,
			EnableReordering:  false,
			SelectedIndex:     0,
		})

	if err != nil {
		return models.Platform{}, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered && ps.QuitOnBack {
		return nil, 4, nil
	} else if selection.IsSome() {
		return selection.Unwrap().SelectedItem.Metadata.(models.Platform), 0, nil
	}

	return nil, 2, nil

}
