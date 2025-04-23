package ui

import (
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
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

func (m MainMenu) Draw() (host models.ScreenReturn, exitCode int, e error) {
	var extraArgs []string
	extraArgs = append(extraArgs, "--cancel-text", "QUIT")

	selection, err := cui.DisplayList(m.Hosts, "Mortar", "", extraArgs...)
	if err != nil {
		return models.Host{}, -1, err
	}

	if selection.ExitCode == 0 {
		hostIdx := m.HostIndices[selection.SelectedValue]
		return m.Hosts[hostIdx], selection.ExitCode, nil
	}

	return models.Host{}, selection.ExitCode, nil
}
