package ui

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"mortar/models"
	"mortar/utils"
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

func (ps PlatformSelection) Draw() (p models.ScreenReturn, exitCode int, e error) {
	var extraArgs []string
	if ps.QuitOnBack {
		extraArgs = append(extraArgs, "--cancel-text", "QUIT")
	}

	actionText := ""
	if ps.Host.HostType == shared.HostTypes.MEGATHREAD && utils.CacheFolderExists() {
		actionText = "CLEAR CACHE"
	}

	if len(ps.Host.Platforms) == 0 {
		return models.Platform{}, 404, nil
	}

	selection, err := cui.DisplayList(ps.Host.Platforms, ps.Host.DisplayName, actionText, extraArgs...)
	if err != nil {
		return models.Platform{}, -1, err
	}

	if selection.ExitCode == 0 {
		idx := ps.Host.GetPlatformIndices()[selection.SelectedValue]
		platform := ps.Host.Platforms[idx]
		platform.Host = ps.Host

		return platform, selection.ExitCode, nil
	}

	backExitCode := selection.ExitCode
	if ps.QuitOnBack {
		backExitCode = 2
	}

	return models.Platform{Host: ps.Host}, backExitCode, nil

}
