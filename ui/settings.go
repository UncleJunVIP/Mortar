package ui

import (
	"fmt"
	gabamod "github.com/UncleJunVIP/gabagool/models"
	"github.com/UncleJunVIP/gabagool/ui"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"mortar/models"
	"mortar/state"
	"mortar/utils"
	"qlova.tech/sum"
)

type SettingsScreen struct {
}

func InitSettingsScreen() SettingsScreen {
	return SettingsScreen{}
}

func (s SettingsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Settings
}

func (s SettingsScreen) Draw() (settings interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	appState := state.GetAppState()

	items := []ui.ItemWithOptions{
		{
			Item: gabamod.MenuItem{
				Text: "Download Art",
			},
			Options: []ui.Option{
				{DisplayName: "True", Value: true},
				{DisplayName: "False", Value: false},
			},
			SelectedOption: func() int {
				if appState.Config.DownloadArt {
					return 0
				}
				return 1
			}(),
		},
		{
			Item: gabamod.MenuItem{
				Text: "Art Type",
			},
			Options: []ui.Option{
				{DisplayName: "Box Art", Value: "BOX_ART"},
				{DisplayName: "Title Screen", Value: "TITLE_SCREEN"},
				{DisplayName: "Logos", Value: "LOGOS"},
				{DisplayName: "Screenshots", Value: "SCREENSHOTS"},
			},
			SelectedOption: func() int {
				switch appState.Config.ArtDownloadType {
				case shared.ArtDownloadTypes.BOX_ART:
					return 0
				case shared.ArtDownloadTypes.TITLE_SCREEN:
					return 1
				case shared.ArtDownloadTypes.LOGOS:
					return 2
				case shared.ArtDownloadTypes.SCREENSHOTS:
					return 3
				default:
					return 0
				}
			}(),
		},
	}

	if utils.CacheFolderExists() {
		items = append(items, ui.ItemWithOptions{
			Item: gabamod.MenuItem{
				Text: "Empty Cache",
			},
			Options: []ui.Option{
				{
					DisplayName: "",
					Value:       "empty",
					Type:        ui.OptionTypeClickable,
				},
			},
		})
	}

	footerHelpItems := []ui.FooterHelpItem{
		{ButtonName: "B", HelpText: "Cancel"},
		{ButtonName: "←→", HelpText: "Change option"},
		{ButtonName: "Start", HelpText: "Save"},
	}

	result, err := ui.OptionsList(
		"Mortar Settings",
		items,
		footerHelpItems,
	)

	if err != nil {
		fmt.Println("Error showing options list:", err)
		return
	}

	if result.IsSome() {
		if result.Unwrap().SelectedItem.Item.Text == "Empty Cache" {
			_ = utils.DeleteCache()

			_, _ = ui.BlockingProcess(fmt.Sprintf("Cache Emptied!"), func() (interface{}, error) {
				return nil, nil
			})
			return result, 404, nil
		}

		newSettingOptions := result.Unwrap().Items

		for _, option := range newSettingOptions {
			if option.Item.Text == "Download Art" {
				if option.SelectedOption == 0 {
					appState.Config.DownloadArt = true
				} else {
					appState.Config.DownloadArt = false
				}
			} else if option.Item.Text == "Art Type" {
				artTypeValue := option.Options[option.SelectedOption].Value.(string)
				switch artTypeValue {
				case "BOX_ART":
					appState.Config.ArtDownloadType = shared.ArtDownloadTypes.BOX_ART
				case "TITLE_SCREEN":
					appState.Config.ArtDownloadType = shared.ArtDownloadTypes.TITLE_SCREEN
				case "LOGOS":
					appState.Config.ArtDownloadType = shared.ArtDownloadTypes.LOGOS
				case "SCREENSHOTS":
					appState.Config.ArtDownloadType = shared.ArtDownloadTypes.SCREENSHOTS
				}
			}
		}

		err := SaveConfig(appState.Config)
		if err != nil {
			logger.Error("Error saving config", zap.Error(err))
			return nil, 0, err
		}

		state.UpdateAppState(appState)

		return result, 0, nil
	}

	return nil, 2, nil
}

func SaveConfig(config *models.Config) error {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	viper.Set("download_art", config.DownloadArt)
	viper.Set("art_download_type", config.ArtDownloadType)

	return viper.WriteConfigAs("config.yml")
}
