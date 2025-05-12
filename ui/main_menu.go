package ui

import (
	gabamod "github.com/UncleJunVIP/gabagool/models"
	gaba "github.com/UncleJunVIP/gabagool/ui"
	"mortar/models"
	"qlova.tech/sum"
)

type MainMenu struct {
	Hosts       models.Hosts
	HostIndices map[string]int
}

func InitMainMenu(hosts models.Hosts) MainMenu {
	hostIndices := make(map[string]int)
	for idx, host := range hosts {
		hostIndices[host.DisplayName] = idx
	}

	return MainMenu{
		Hosts:       hosts,
		HostIndices: hostIndices,
	}
}

func (m MainMenu) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.MainMenu
}

func (m MainMenu) Draw() (host interface{}, exitCode int, e error) {

	var menuItems []gabamod.MenuItem
	for _, host := range m.Hosts {
		menuItems = append(menuItems, gabamod.MenuItem{
			Text:     host.DisplayName,
			Selected: false,
			Focused:  false,
			Metadata: host,
		})
	}

	fhi := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Quit"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gaba.NewBlockingList("Mortar", menuItems, "", fhi, false, false, false)
	if err != nil {
		return models.Host{}, -1, err
	}

	if selection.IsSome() {
		return selection.Unwrap().SelectedItem.Metadata.(models.Host), 0, nil
	}

	return models.Host{}, 2, nil
}
