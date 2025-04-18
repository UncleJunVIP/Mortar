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
	"time"
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

func (d DownloadScreen) Draw() (xxx models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	exitCodeChan := make(chan int, 1)
	filepathChan := make(chan string, 1)
	errChan := make(chan error, 1)

	args := []string{"--message", "Downloading " + d.Game.Filename + "...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with starting miniui-presenter download message", zap.Error(err))
	}

	time.Sleep(1000 * time.Millisecond)

	go func() {
		filepath, err := utils.DownloadFile(d.Platform, d.Games, d.Game, cancel)
		if err != nil {
			logger.Error("Error downloading file: %s", zap.Error(err))
			errChan <- err
			exitCodeChan <- 1
		} else {
			filepathChan <- filepath
			exitCodeChan <- 0
		}
		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() > 6 {
		logger.Fatal("Error with minui-presenter display of download message: %s", zap.Error(err))
	}

	filepath := <-filepathChan
	exitCode = <-exitCodeChan

	select {
	case err := <-errChan:
		return nil, exitCode, err
	default:
		return shared.Item{
			DisplayName: d.Game.DisplayName,
			Filename:    d.Game.Filename,
			Path:        filepath,
			IsDirectory: false,
		}, 0, nil
	}
}
