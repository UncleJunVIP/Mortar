package main

import (
	"go.uber.org/zap"
	"mortar/common"
	"mortar/ui"
	"mortar/utils"
	"os"
	"strings"
)

func init() {
	common.SetLogLevel("DEBUG")
	common.ConfigureEnvironment()

	config, err := common.LoadConfig()
	if err != nil {
		ui.ShowMessage("Unable to parse config.yml! Quitting!", "3")
		common.LogStandardFatal("Error loading config", err)
	}

	logger := common.GetLoggerInstance()

	romDirectories, err := utils.FetchRomDirectories()
	if err != nil {
		logger.Error("Issue fetching rom directories", zap.Error(err))
	} else {
		for hostIdx, host := range config.Hosts {
			for sectionIdx, section := range host.Sections {
				if section.SystemTag != "" {
					config.Hosts[hostIdx].Sections[sectionIdx].LocalDirectory = romDirectories[section.SystemTag]
				}
			}
		}
	}

	logger.Debug("Config Loaded",
		zap.Object("config", config))

	common.SetConfig(config)

	appState := common.GetAppState()

	if len(appState.Config.Hosts) == 1 {
		appState.CurrentScreen = ui.Screens.SectionSelection
		appState.CurrentHost = appState.Config.Hosts[0]
	} else {
		appState.CurrentScreen = ui.Screens.MainMenu
	}

	common.UpdateAppState(appState)
}

func cleanup() {
	common.CloseLogger()
}

func main() {
	defer cleanup()

	for {
		appState := common.GetAppState()

		selection := ui.ScreenFuncs[appState.CurrentScreen]()

		// Hacky way to handle bad input on deep sleep
		if strings.Contains(selection.Value, "SetRawBrightness") ||
			strings.Contains(selection.Value, "nSetRawVolume") {
			continue
		}

		switch appState.CurrentScreen {
		case ui.Screens.MainMenu:
			switch selection.Code {
			case 0:
				ui.SetScreen(ui.Screens.SectionSelection)
				idx := appState.HostIndices[strings.TrimSpace(selection.Value)]
				common.SetHost(appState.Config.Hosts[idx])
			case 1, 2:
				os.Exit(0)
			}

		case ui.Screens.SectionSelection:
			switch selection.Code {
			case 0:
				ui.SetScreen(ui.Screens.Loading)
				idx := appState.CurrentHost.GetSectionIndices()[strings.TrimSpace(selection.Value)]
				common.SetSection(appState.CurrentHost.Sections[idx])
			case 1, 2:
				if len(appState.Config.Hosts) == 1 {
					os.Exit(0)
				}
				ui.SetScreen(ui.Screens.MainMenu)
			}

		case ui.Screens.ItemList:
			switch selection.Code {
			case 0:
				for _, item := range appState.CurrentItemsList {
					if strings.Contains(item.Filename, strings.TrimSpace(selection.Value)) {
						common.SetSelectedFile(item.Filename)
						break
					}
				}

				ui.SetScreen(ui.Screens.Download)
			case 2:
				if appState.SearchFilter != "" {
					common.SetSearchFilter("")
				} else {
					ui.SetScreen(ui.Screens.SectionSelection)
				}
			case 4:
				ui.SetScreen(ui.Screens.SearchBox)
			case 404:
				if appState.SearchFilter != "" {
					ui.ShowMessage("No results found for \""+appState.SearchFilter+"\"", "3")
					common.SetSearchFilter("")
					ui.SetScreen(ui.Screens.SearchBox)
				} else {
					ui.ShowMessage("This section contains no items", "3")
					ui.SetScreen(ui.Screens.SectionSelection)
				}
			}

		case ui.Screens.Loading:
			switch selection.Code {
			case 0:
				ui.SetScreen(ui.Screens.ItemList)
			case 1:
				ui.ShowMessage("Unable to download item listing from source", "3")
				ui.SetScreen(ui.Screens.MainMenu)
			}

		case ui.Screens.SearchBox:
			switch selection.Code {
			case 0:
				common.SetSearchFilter(selection.Value)
			case 1, 2, 3:
				common.SetSearchFilter("")
			}

			ui.SetScreen(ui.Screens.ItemList)

		case ui.Screens.Download:
			switch selection.Code {
			case 0:
				if appState.Config.DownloadArt {
					ui.SetScreen(ui.Screens.DownloadArt)
				} else {
					ui.SetScreen(ui.Screens.ItemList)
				}

			case 1:
				ui.ShowMessage("Unable to download "+appState.SelectedFile, "3")
				ui.SetScreen(ui.Screens.ItemList)

			default:
				ui.SetScreen(ui.Screens.ItemList)
			}

		case ui.Screens.DownloadArt:
			switch selection.Code {
			case 0:
				code := ui.ShowMessageWithOptions("　　　　　　　　　　　　　　　　　　　　　　　　　", "0",
					"--background-image", common.GetAppState().LastSavedArtPath,
					"--confirm-text", "Use",
					"--confirm-show", "true",
					"--action-button", "X",
					"--action-text", "I'll Find My Own",
					"--action-show", "true",
					"--message-alignment", "bottom")

				if code == 2 || code == 4 {
					utils.DeleteFile(common.GetAppState().LastSavedArtPath)
				}
			case 1:
				ui.ShowMessage("Could not find art :(", "3")
			}
			ui.SetScreen(ui.Screens.ItemList)
		}
	}
}
