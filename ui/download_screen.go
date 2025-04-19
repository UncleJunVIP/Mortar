package ui

import (
	"context"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"mortar/models"
	"mortar/utils"
	"os/exec"
	"qlova.tech/sum"
)

type DownloadScreen struct {
	Platform     models.Platform
	Games        shared.Items
	Game         shared.Item
	SearchFilter string
}

func InitDownloadScreen(platform models.Platform, games shared.Items,
	game shared.Item, searchFilter string) DownloadScreen {
	return DownloadScreen{
		Platform:     platform,
		Games:        games,
		Game:         game,
		SearchFilter: searchFilter,
	}
}

func (d DownloadScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Download
}

func (d DownloadScreen) Draw() (value models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{"--message", "Downloading " + d.Game.Filename + "...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with starting miniui-presenter download message", zap.Error(err))
	}

	filepath, err := utils.DownloadFile(d.Platform, d.Games, d.Game)

	if err != nil {
		logger.Error("Error downloading file: %s", zap.Error(err))
	}

	cancel()

	if filepath != "" {
		return shared.Item{
			DisplayName: d.Game.DisplayName,
			Filename:    d.Game.Filename,
			Path:        filepath,
			IsDirectory: false,
		}, 0, nil
	}

	return nil, 1, err
}
