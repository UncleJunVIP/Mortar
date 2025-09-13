package main

import (
	"fmt"
	"mortar/models"
	"mortar/state"
	"mortar/ui"
	"mortar/utils"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/UncleJunVIP/certifiable"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
)

func init() {
	gaba.InitSDL(gaba.GabagoolOptions{
		WindowTitle:    "Mortar",
		ShowBackground: true,
	})

	common.SetLogLevel("ERROR")

	common.InitIncludes()

	if !utils.IsConnectedToInternet() {
		fmt.Println("No Internet Connection")
		_, err := gaba.ConfirmationMessage("No Internet Connection!\nMake sure you are connected to Wi-Fi.", []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
		}, gaba.MessageOptions{})
		defer cleanup()
		common.LogStandardFatal("No Internet Connection", err)
	}

	config, err := state.LoadConfig()
	if err != nil {
		fmt.Println("Setup Required")
		_, err := gaba.ConfirmationMessage("Setup Required!\nScan the QR Code for Instructions", []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
		}, gaba.MessageOptions{ImagePath: "resources/setup-qr.png"})
		defer cleanup()
		common.LogStandardFatal("Setup Required", err)
	}

	if config.LogLevel != "" {
		common.SetLogLevel(config.LogLevel)
	}

	logger := common.GetLoggerInstance()

	logger.Info("Configuration Loaded!", "config", config)

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
		defer cleanup()
		logger.Error("Error loading fetching ROM directories", "error", err)
	}

	romDirectories := utils.MapTagsToDirectories(fb.Items)

	logger.Debug(fmt.Sprintf("Discovered %d ROM Directories",
		len(romDirectories)))

	for tag, path := range romDirectories {
		logger.Debug(fmt.Sprintf("Mapped System %s", tag), "tag", tag, "path", path)
	}

	for hostIdx, host := range config.Hosts {
		for sectionIdx, section := range host.Platforms {
			if section.SystemTag != "" {
				config.Hosts[hostIdx].Platforms[sectionIdx].LocalDirectory = romDirectories[section.SystemTag]
			}
		}
	}

	missingPlatforms := utils.AllPlatformsHaveLocalFolders(config)

	if len(missingPlatforms) > 0 {
		mps := strings.Join(missingPlatforms, "\n")
		gaba.ConfirmationMessage(fmt.Sprintf("These platforms are missing local folders:\n%s", mps), []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
		}, gaba.MessageOptions{})
		logger.Error("Not all platforms have local folders. Ensure they are correct and try again. "+
			"If you are using the automation folder detection feature, please make sure a matching system tag exits on a directory in your ROM directory.",
			"missing_platforms", missingPlatforms)

		os.Exit(1)
	}

	logger.Info("Populated ROM Directories by System Tag", "config", config)

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
			state.SetLastSelectedPosition(0, 0)
			switch code {
			case 0:
				platform := res.(models.Platform)
				screen = ui.InitGamesList(platform, shared.Items{}, "")
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
					screen = ui.InitGamesList(gl.Platform, state.GetAppState().CurrentFullGamesList, "")
				} else {
					screen = ui.InitPlatformSelection(gl.Platform.Host, quitOnBack)
				}

			case 4:
				screen = ui.InitSearch(gl.Platform, gl.SearchFilter)

			case 404:
				if gl.SearchFilter != "" {
					screen = ui.InitGamesList(gl.Platform, state.GetAppState().CurrentFullGamesList, "")
				} else {
					screen = ui.InitPlatformSelection(gl.Platform.Host, quitOnBack)
				}
			}
		case ui.Screens.SearchBox:
			sb := screen.(ui.Search)
			switch code {
			case 0:
				query := res.(string)
				state.SetLastSelectedPosition(0, 0)
				screen = ui.InitGamesList(sb.Platform, state.GetAppState().CurrentFullGamesList, query)
			default:
				screen = ui.InitGamesList(sb.Platform, state.GetAppState().CurrentFullGamesList, "")
			}
		case ui.Screens.Download:
			ds := screen.(ui.DownloadScreen)
			switch code {
			case 0:
				downloadedGames := res.([]shared.Item)

				for _, game := range downloadedGames {
					isMultiDisc := utils.IsMultiDisc(ds.Platform, game)

					if filepath.Ext(game.Filename) == ".zip" && !screen.(ui.DownloadScreen).Platform.IsArcade {
						isBinCue := utils.HasBinCue(ds.Platform, game)

						if isMultiDisc && appState.Config.GroupMultiDisc {
							utils.GroupMultiDisk(ds.Platform, game)
						} else if appState.Config.GroupBinCue && isBinCue {
							utils.GroupBinCue(ds.Platform, game)
						} else if appState.Config.UnzipDownloads {
							utils.UnzipGame(ds.Platform, game)
						}
					} else if appState.Config.GroupMultiDisc && isMultiDisc {
						utils.GroupMultiDisk(ds.Platform, game)
					}
				}

				if appState.Config.DownloadArt {
					seenBaseNames := make(map[string]bool)

					// Create a pruned list for art downloads that only includes one instance of each multi-disk game
					prunedGamesForArt := make([]shared.Item, 0, len(downloadedGames))

					for _, game := range downloadedGames {
						// Get base name by trimming at "(Disk" or "(Disc"
						baseName := game.DisplayName
						diskIndex := strings.Index(baseName, "(Disk")
						discIndex := strings.Index(baseName, "(Disc")

						trimIndex := -1
						if diskIndex != -1 && discIndex != -1 {
							trimIndex = min(diskIndex, discIndex)
						} else if diskIndex != -1 {
							trimIndex = diskIndex
						} else if discIndex != -1 {
							trimIndex = discIndex
						}

						if trimIndex != -1 {
							baseName = baseName[:trimIndex]
						}
						baseName = strings.TrimSpace(baseName)

						// If we haven't seen this base name before, add it to the pruned list
						if !seenBaseNames[baseName] {
							seenBaseNames[baseName] = true
							game.Filename = baseName
							prunedGamesForArt = append(prunedGamesForArt, game)
						}
					}

					screen = ui.InitDownloadArtScreen(ds.Platform, prunedGamesForArt, appState.Config.ArtDownloadType, ds.SearchFilter)
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
			screen = ui.InitGamesList(da.Platform, state.GetAppState().CurrentFullGamesList, da.SearchFilter)
		}
	}
}
