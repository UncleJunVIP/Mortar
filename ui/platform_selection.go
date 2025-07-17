package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
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

	var menuItems []gabagool.MenuItem

	//if ps.Host.HostType == shared.HostTypes.MEGATHREAD {
	//	menuItems = append(menuItems, gabagool.MenuItem{
	//		Text:     "All Games",
	//		Selected: false,
	//		Focused:  false,
	//		Metadata: models.Platform{
	//			Name: "All Games",
	//			Host: ps.Host,
	//		},
	//	})
	//}

	for _, platform := range ps.Host.Platforms {
		platform.Host = ps.Host
		menuItems = append(menuItems, gabagool.MenuItem{
			Text:     platform.Name,
			Selected: false,
			Focused:  false,
			Metadata: platform,
		})
	}

	var fhi []gabagool.FooterHelpItem

	if ps.QuitOnBack {
		fhi = []gabagool.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
			{ButtonName: "X", HelpText: "Settings"},
			{ButtonName: "A", HelpText: "Select"},
		}
	} else {
		fhi = []gabagool.FooterHelpItem{
			{ButtonName: "B", HelpText: "Back"},
			{ButtonName: "A", HelpText: "Select"},
		}
	}

	title := ps.Host.DisplayName

	if ps.QuitOnBack {
		title = fmt.Sprintf("Mortar | %s", ps.Host.DisplayName)
	}

	options := gabagool.DefaultListOptions(title, menuItems)
	options.EnableAction = ps.QuitOnBack
	options.FooterHelpItems = fhi

	selection, err := gabagool.List(options)

	if err != nil {
		return models.Platform{}, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered && ps.QuitOnBack {
		return nil, 4, nil
	} else if selection.IsSome() && selection.Unwrap().SelectedIndex != -1 {
		if selection.Unwrap().SelectedItem.Metadata.(models.Platform).Name == "All Games" {
			return models.Platform{
				Name: "All Games",
				Host: ps.Host,
			}, 5, nil
		}
		return selection.Unwrap().SelectedItem.Metadata.(models.Platform), 0, nil
	}

	return nil, 2, nil

}
