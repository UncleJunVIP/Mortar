package main

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"mortar/models"
	"mortar/state"
	"mortar/ui"
	"mortar/utils"
	"os"
)

func init() {
	common.SetLogLevel("ERROR")

	config, err := state.LoadConfig()
	if err != nil {
		options := []string{
			"--background-image", "setup-qr.png",
			"--confirm-text", "EXIT",
			"--confirm-show", "true",
			"--message-alignment", "bottom"}

		_, err = cui.ShowMessageWithOptions("Setup Required!", "0", options...)
		common.LogStandardFatal("Setup Required", err)
	}

	if config.LogLevel != "" {
		common.SetLogLevel(config.LogLevel)
	}

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
	err = fb.CWD(common.RomDirectory, false)
	if err != nil {
		_, _ = cui.ShowMessage("Unable to fetch ROM directories! Quitting!", "3")
		logger.Fatal("Error loading fetching ROM directories", zap.Error(err))
	}

	romDirectories := utils.MapTagsToDirectories(fb.Items)

	for hostIdx, host := range config.Hosts {
		for sectionIdx, section := range host.Platforms {
			if section.SystemTag != "" {
				config.Hosts[hostIdx].Platforms[sectionIdx].LocalDirectory = romDirectories[section.SystemTag]
			}
		}
	}

	state.SetConfig(config)
}

func cleanup() {
	common.CloseLogger()
}

func main() {
	defer cleanup()

	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	logger.Info("Starting Mortar")

	var screen models.Screen

	if len(appState.Config.Hosts) == 1 {
		screen = ui.InitPlatformSelection(appState.Config.Hosts[0], true)
	} else {
		screen = ui.InitMainMenu(appState.Config.Hosts)
	}

	for {
		res, code, _ := screen.Draw() // TODO figure out error handling

		switch screen.Name() {
		case ui.Screens.MainMenu:
			switch code {
			case 0:
				host := res.Value().(models.Host)
				screen = ui.InitPlatformSelection(host, false)
			case 1, 2:
				os.Exit(0)
			}
		case ui.Screens.PlatformSelection:
			platform := res.Value().(models.Platform)
			switch code {
			case 0:
				screen = ui.InitGamesList(platform, shared.Items{}, "")
			case 1, 2:
				if screen.(ui.PlatformSelection).QuitOnBack {
					os.Exit(0)
				}
				screen = ui.InitMainMenu(appState.Config.Hosts)
			case 4:
				err := utils.DeleteCache()
				if err != nil {
					_, _ = cui.ShowMessage("Unable to delete cache!", "3")
				} else {
					_, _ = cui.ShowMessage("Cache deleted!", "3")
				}
				screen = ui.InitPlatformSelection(platform.Host, len(appState.Config.Hosts) == 0)
			case 404:
				_, _ = cui.ShowMessage("No platforms configured for \""+platform.Host.DisplayName+"\"", "3")
				screen = ui.InitMainMenu(appState.Config.Hosts)
			case -1:
				_, _ = cui.ShowMessage("Unable to display platforms for \""+platform.Host.DisplayName+"\"", "3")
				screen = ui.InitMainMenu(appState.Config.Hosts)
			}
		case ui.Screens.GameList:
			gl := screen.(ui.GameList)

			switch code {
			case 0:
				game := res.Value().(shared.Item)
				screen = ui.InitDownloadScreen(gl.Platform, gl.Games, game, gl.SearchFilter)

			case 2:
				if gl.SearchFilter != "" {
					screen = ui.InitGamesList(gl.Platform, shared.Items{}, "") // Clear search filter
				} else {
					screen = ui.InitPlatformSelection(gl.Platform.Host, len(appState.Config.Hosts) == 0)
				}

			case 4:
				screen = ui.InitSearch(gl.Platform)

			case 404:
				if gl.SearchFilter != "" {
					_, _ = cui.ShowMessage("No results found for \""+gl.SearchFilter+"\"", "3")
					screen = ui.InitGamesList(gl.Platform, shared.Items{}, "")
				} else {
					_, _ = cui.ShowMessage("This section contains no items", "3")
					screen = ui.InitPlatformSelection(gl.Platform.Host, len(appState.Config.Hosts) == 0)
				}
			}
		case ui.Screens.SearchBox:
			sb := screen.(ui.Search)
			switch code {
			case 0:
				query := res.Value().(string)
				screen = ui.InitGamesList(sb.Platform, shared.Items{}, query)
			default:
				screen = ui.InitGamesList(sb.Platform, shared.Items{}, "")
			}
		case ui.Screens.Download:
			ds := screen.(ui.DownloadScreen)
			switch code {
			case 0:
				if appState.Config.DownloadArt {
					screen = ui.InitDownloadArtScreen(ds.Platform, ds.Game, appState.Config.ArtDownloadType, ds.SearchFilter)
				} else {
					screen = ui.InitGamesList(ds.Platform, shared.Items{}, "")
				}
			case 1:
				_, _ = cui.ShowMessage("Unable to download "+ds.Game.DisplayName, "3")
				screen = ui.InitGamesList(ds.Platform, shared.Items{}, "")
			default:
				screen = ui.InitGamesList(ds.Platform, shared.Items{}, "")
			}
		case ui.Screens.DownloadArt:
			da := screen.(ui.DownloadArtScreen)
			switch code {
			case 404:
				_, _ = cui.ShowMessage("Could not find art :(", "3")
			}

			screen = ui.InitGamesList(da.Platform, shared.Items{}, da.SearchFilter)
		}
	}
}
