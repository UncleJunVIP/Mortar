package main

import (
	"mortar/ui"
	"mortar/utils"
	"os"
	"strings"
)

func init() {
	utils.SetLogLevel("DEBUG")
	utils.ConfigureEnvironment()

	config, err := utils.LoadConfig()
	if err != nil {
		ui.ShowMessage("Unable to parse config.yml! Quitting!", "3")
		utils.LogStandardFatal("Error loading config", err)
	}

	utils.SetConfig(config)

	appState := utils.GetAppState()

	appState.HostIndices = make(map[string]int)
	for idx, host := range appState.Config.Hosts {
		appState.HostIndices[host.DisplayName] = idx
	}

	if len(appState.Config.Hosts) == 1 {
		appState.CurrentScreen = ui.Screens.SectionSelection
		appState.CurrentHost = appState.Config.Hosts[0]
	} else {
		appState.CurrentScreen = ui.Screens.MainMenu
	}

	utils.UpdateAppState(appState)
}

func cleanup() {
	utils.CloseLogger()
}

func main() {
	defer cleanup()

	for {
		appState := utils.GetAppState()

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
				utils.SetHost(appState.Config.Hosts[idx])
			case 1, 2:
				os.Exit(0)
			}

		case ui.Screens.SectionSelection:
			switch selection.Code {
			case 0:
				ui.SetScreen(ui.Screens.Loading)
				idx := appState.CurrentHost.GetSectionIndices()[strings.TrimSpace(selection.Value)]
				utils.SetSection(appState.CurrentHost.Sections[idx])
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
						utils.SetSelectedFile(item.Filename)
						break
					}
				}

				ui.SetScreen(ui.Screens.Download)
			case 2:
				if appState.SearchFilter != "" {
					utils.SetSearchFilter("")
				} else {
					ui.SetScreen(ui.Screens.SectionSelection)
				}
			case 4:
				ui.SetScreen(ui.Screens.SearchBox)
			case 404:
				if appState.SearchFilter != "" {
					ui.ShowMessage("No results found for \""+appState.SearchFilter+"\"", "3")
					utils.SetSearchFilter("")
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
				utils.SetSearchFilter(selection.Value)
			case 1, 2, 3:
				utils.SetSearchFilter("")
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
				ui.ShowMessage("Found art! :)", "3")
			case 1:
				ui.ShowMessage("Could not find art :(", "3")
			}
			ui.SetScreen(ui.Screens.ItemList)
		}
	}
}
