package main

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	commonUI "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"mortar/state"
	"mortar/ui"
	"mortar/utils"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	common.SetLogLevel("ERROR")
	common.ConfigureEnvironment()

	config, err := state.LoadConfig()
	if err != nil {
		_, _ = commonUI.ShowMessage("Unable to parse config.yml! Quitting!", "3")
		common.LogStandardFatal("Error loading config", err)
	}

	common.SetLogLevel(config.LogLevel)

	logger := common.GetLoggerInstance()

	if config.RawArtDownloadType == "" {
		config.RawArtDownloadType = "BOX_ART"
	}

	if val, ok := shared.ArtDownloadTypeFromString[config.RawArtDownloadType]; ok {
		config.ArtDownloadType = val
	} else {
		logger.Info("Invalid art download type provided... defaulting to BOX_ART")
		config.ArtDownloadType = shared.ArtDownloadTypes.BOX_ART
	}

	fb := filebrowser.NewFileBrowser(logger)
	err = fb.CWD(common.RomDirectory)
	if err != nil {
		_, _ = commonUI.ShowMessage("Unable to fetch ROM directories! Quitting!", "3")
		common.LogStandardFatal("Error loading fetching ROM directories", err)
	}

	romDirectories := utils.MapTagsToDirectories(fb.Items)

	for hostIdx, host := range config.Hosts {
		for sectionIdx, section := range host.Sections {
			if section.SystemTag != "" {
				config.Hosts[hostIdx].Sections[sectionIdx].LocalDirectory = romDirectories[section.SystemTag]
			}
		}
	}

	//logger.Debug("Config Loaded", zap.Object("config", config))

	state.SetConfig(config)

	appState := state.GetAppState()

	if len(appState.Config.Hosts) == 1 {
		appState.CurrentScreen = ui.Screens.SectionSelection
		appState.CurrentHost = appState.Config.Hosts[0]
	} else {
		appState.CurrentScreen = ui.Screens.MainMenu
	}

	state.UpdateAppState(appState)
}

func cleanup() {
	common.CloseLogger()
}

func main() {
	defer cleanup()

	logger := common.GetLoggerInstance()

	for {
		appState := state.GetAppState()

		selection, err := ui.ScreenFuncs[appState.CurrentScreen]()

		if err != nil {
			logger.Error("Error loading screen")
		}

		// Hacky way to handle bad input on deep sleep
		if strings.Contains(selection.Value, "SetRawBrightness") ||
			strings.Contains(selection.Value, "nSetRawVolume") {
			continue
		}

		switch appState.CurrentScreen {
		case ui.Screens.MainMenu:
			switch selection.ExitCode {
			case 0:
				ui.SetScreen(ui.Screens.SectionSelection)
				idx := appState.HostIndices[strings.TrimSpace(selection.Value)]
				state.SetHost(appState.Config.Hosts[idx])
			case 1, 2:
				os.Exit(0)
			}

		case ui.Screens.SectionSelection:
			switch selection.ExitCode {
			case 0:
				ui.SetScreen(ui.Screens.Loading)
				idx := appState.CurrentHost.GetSectionIndices()[strings.TrimSpace(selection.Value)]
				state.SetSection(appState.CurrentHost.Sections[idx])
			case 1, 2:
				if len(appState.Config.Hosts) == 1 {
					os.Exit(0)
				}
				ui.SetScreen(ui.Screens.MainMenu)
			}

		case ui.Screens.ItemList:
			switch selection.ExitCode {
			case 0:
				selectedItem := strings.TrimSpace(selection.Value)
				for _, item := range appState.CurrentItemsList {
					itemWithoutExt := strings.ReplaceAll(item.Filename, filepath.Ext(item.Filename), "")
					if selectedItem == itemWithoutExt {
						state.SetSelectedFile(item.Filename)
						break
					}
				}

				ui.SetScreen(ui.Screens.Download)
			case 2:
				if appState.SearchFilter != "" {
					state.SetSearchFilter("")
				} else {
					ui.SetScreen(ui.Screens.SectionSelection)
				}
			case 4:
				ui.SetScreen(ui.Screens.SearchBox)
			case 404:
				if appState.SearchFilter != "" {
					_, _ = commonUI.ShowMessage("No results found for \""+appState.SearchFilter+"\"", "3")
					state.SetSearchFilter("")
					ui.SetScreen(ui.Screens.SearchBox)
				} else {
					_, _ = commonUI.ShowMessage("This section contains no items", "3")
					ui.SetScreen(ui.Screens.SectionSelection)
				}
			}

		case ui.Screens.Loading:
			switch selection.ExitCode {
			case 0:
				ui.SetScreen(ui.Screens.ItemList)
			case 1:
				_, _ = commonUI.ShowMessage("Unable to download item listing from source", "3")
				ui.SetScreen(ui.Screens.MainMenu)
			}

		case ui.Screens.SearchBox:
			switch selection.ExitCode {
			case 0:
				state.SetSearchFilter(selection.Value)
			case 1, 2, 3:
				state.SetSearchFilter("")
			}

			ui.SetScreen(ui.Screens.ItemList)

		case ui.Screens.Download:
			switch selection.ExitCode {
			case 0:
				if appState.Config.DownloadArt {
					ui.SetScreen(ui.Screens.DownloadArt)
				} else {
					ui.SetScreen(ui.Screens.ItemList)
				}

			case 1:
				_, _ = commonUI.ShowMessage("Unable to download "+appState.SelectedFile, "3")
				ui.SetScreen(ui.Screens.ItemList)

			default:
				ui.SetScreen(ui.Screens.ItemList)
			}

		case ui.Screens.DownloadArt:
			switch selection.ExitCode {
			case 0:
				logger := common.GetLoggerInstance()

				logger.Info("Art Path", zap.String("lsap", state.GetAppState().LastSavedArtPath))

				code, _ := commonUI.ShowMessageWithOptions("　　　　　　　　　　　　　　　　　　　　　　　　　", "0",
					"--background-image", state.GetAppState().LastSavedArtPath,
					"--confirm-text", "Use",
					"--confirm-show", "true",
					"--action-button", "X",
					"--action-text", "I'll Find My Own",
					"--action-show", "true",
					"--message-alignment", "bottom")

				if code == 2 || code == 4 {
					common.DeleteFile(state.GetAppState().LastSavedArtPath)
				}
			case 1:
				_, _ = commonUI.ShowMessage("Could not find art :(", "3")
			}
			ui.SetScreen(ui.Screens.ItemList)
		}
	}
}
