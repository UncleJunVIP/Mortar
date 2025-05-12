package ui

import (
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
	// TODO add clear cache back here

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

	fhi := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gaba.NewBlockingList(ps.Host.DisplayName, menuItems, "", fhi, false, false, false)
	if err != nil {
		return models.Platform{}, -1, err
	}

	if selection.IsSome() {
		return selection.Unwrap().SelectedItem.Metadata.(models.Platform), 0, nil
	}

	return models.Platform{Host: ps.Host}, 2, nil

}
