package main

import (
	gaba "github.com/UncleJunVIP/gabagool/ui"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
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
		// TODO QR CODE HERE
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
	err = fb.CWD(utils.GetRomDirectory(), false)
	if err != nil {
		// TODO DISPLAY ERROR HERE
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
	gaba.InitSDL("Mortar")
	defer gaba.CloseSDL()
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
				host := res.(models.Host)
				screen = ui.InitPlatformSelection(host, false)
			case 1, 2:
				os.Exit(0)
			}
		case ui.Screens.PlatformSelection:
			platform := res.(models.Platform)
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
				} else {
				}
				screen = ui.InitPlatformSelection(platform.Host, len(appState.Config.Hosts) == 0)
			case 404:
				screen = ui.InitMainMenu(appState.Config.Hosts)
			case -1:
				screen = ui.InitMainMenu(appState.Config.Hosts)
			}
		case ui.Screens.GameList:
			gl := screen.(ui.GameList)

			switch code {
			case 0:
				games := res.(shared.Items)
				screen = ui.InitDownloadScreen(gl.Platform, gl.Games, games, gl.SearchFilter)

			case 2:
				if gl.SearchFilter != "" {
					screen = ui.InitGamesList(gl.Platform, state.GetAppState().CurrentFullGamesList, "") // Clear search filter
				} else {
					screen = ui.InitPlatformSelection(gl.Platform.Host, len(appState.Config.Hosts) == 0)
				}

			case 4:
				screen = ui.InitSearch(gl.Platform)

			case 404:
				if gl.SearchFilter != "" {
					screen = ui.InitGamesList(gl.Platform, shared.Items{}, "")
				} else {
					screen = ui.InitPlatformSelection(gl.Platform.Host, len(appState.Config.Hosts) == 0)
				}
			}
		case ui.Screens.SearchBox:
			sb := screen.(ui.Search)
			switch code {
			case 0:
				query := res.(string)
				screen = ui.InitGamesList(sb.Platform, state.GetAppState().CurrentFullGamesList, query)
			default:
				screen = ui.InitGamesList(sb.Platform, state.GetAppState().CurrentFullGamesList, "")
			}
		case ui.Screens.Download:
			ds := screen.(ui.DownloadScreen)
			switch code {
			case 0:
				if appState.Config.DownloadArt {
					screen = ui.InitDownloadArtScreen(ds.Platform, ds.SelectedGames, appState.Config.ArtDownloadType, ds.SearchFilter)
				} else {
					screen = ui.InitGamesList(ds.Platform, state.GetAppState().CurrentFullGamesList, ds.SearchFilter)
				}
			case 1:
				screen = ui.InitGamesList(ds.Platform, state.GetAppState().CurrentFullGamesList, ds.SearchFilter)
			default:
				screen = ui.InitGamesList(ds.Platform, state.GetAppState().CurrentFullGamesList, ds.SearchFilter)
			}
		case ui.Screens.DownloadArt:
			da := screen.(ui.DownloadArtScreen)
			switch code {
			case 404:
			}

			screen = ui.InitGamesList(da.Platform, state.GetAppState().CurrentFullGamesList, da.SearchFilter)
		}
	}
}
