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
	gaba.InitSDL("Mortar")
	common.SetLogLevel("ERROR")

	config, err := state.LoadConfig()
	if err != nil {
		_, err := gaba.Message("Setup Required!", "Scan the QR Code for Instructions", []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
		}, "setup-qr.png")
		cleanup()
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
		cleanup()
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

	defer gaba.CloseSDL()
	defer cleanup()

	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	logger.Info("Starting Mortar")

	var screen models.Screen

	quitOnBack := len(appState.Config.Hosts) == 1

	if quitOnBack {
		screen = ui.InitPlatformSelection(appState.Config.Hosts[0], quitOnBack)
	} else {
		screen = ui.InitMainMenu(appState.Config.Hosts)
	}

	for {
		res, code, _ := screen.Draw()

		switch screen.Name() {
		case ui.Screens.MainMenu:
			switch code {
			case 0:
				host := res.(models.Host)
				screen = ui.InitPlatformSelection(host, quitOnBack)
			case 4:
				screen = ui.InitSettingsScreen()
			case 1, 2:
				os.Exit(0)
			}
		case ui.Screens.Settings:
			if code != 404 {
				if len(appState.Config.Hosts) == 1 {
					screen = ui.InitPlatformSelection(appState.Config.Hosts[0], quitOnBack)
				} else {
					screen = ui.InitMainMenu(appState.Config.Hosts)
				}
			}
		case ui.Screens.PlatformSelection:
			switch code {
			case 0:
				platform := res.(models.Platform)
				screen = ui.InitGamesList(platform, shared.Items{}, "", 0)
			case 1, 2:
				if quitOnBack {
					os.Exit(0)
				}
				screen = ui.InitMainMenu(appState.Config.Hosts)
			case 4:
				screen = ui.InitSettingsScreen()
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
					screen = ui.InitGamesList(gl.Platform, state.GetAppState().CurrentFullGamesList, "", 0) // Clear search filter
				} else {
					screen = ui.InitPlatformSelection(gl.Platform.Host, quitOnBack)
				}

			case 4:
				screen = ui.InitSearch(gl.Platform)

			case 404:
				if gl.SearchFilter != "" {
					screen = ui.InitGamesList(gl.Platform, shared.Items{}, "", 0)
				} else {
					screen = ui.InitPlatformSelection(gl.Platform.Host, quitOnBack)
				}
			}
		case ui.Screens.SearchBox:
			sb := screen.(ui.Search)
			switch code {
			case 0:
				query := res.(string)
				screen = ui.InitGamesList(sb.Platform, state.GetAppState().CurrentFullGamesList, query, 0)
			default:
				screen = ui.InitGamesList(sb.Platform, state.GetAppState().CurrentFullGamesList, "", 0)
			}
		case ui.Screens.Download:
			ds := screen.(ui.DownloadScreen)
			switch code {
			case 0:
				if appState.Config.DownloadArt {
					downloadedGames := res.([]shared.Item)
					screen = ui.InitDownloadArtScreen(ds.Platform, downloadedGames, appState.Config.ArtDownloadType, ds.SearchFilter)
				} else {
					screen = ui.InitGamesList(ds.Platform, state.GetAppState().CurrentFullGamesList, ds.SearchFilter, state.GetAppState().LastSelectedIndex)
				}
			case 1:
				screen = ui.InitGamesList(ds.Platform, state.GetAppState().CurrentFullGamesList, ds.SearchFilter, state.GetAppState().LastSelectedIndex)
			default:
				screen = ui.InitGamesList(ds.Platform, state.GetAppState().CurrentFullGamesList, ds.SearchFilter, state.GetAppState().LastSelectedIndex)
			}
		case ui.Screens.DownloadArt:
			da := screen.(ui.DownloadArtScreen)
			switch code {
			case 404:
			}

			screen = ui.InitGamesList(da.Platform, state.GetAppState().CurrentFullGamesList, da.SearchFilter, state.GetAppState().LastSelectedIndex)
		}
	}
}
