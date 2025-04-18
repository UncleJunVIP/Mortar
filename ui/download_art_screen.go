package ui

import (
	"context"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"mortar/models"
	"mortar/utils"
	"os/exec"
	"qlova.tech/sum"
	"time"
)

type DownloadArtScreen struct {
	Platform     models.Platform
	Game         shared.Item
	DownloadType sum.Int[shared.ArtDownloadType]
	SearchFilter string
}

func InitDownloadArtScreen(platform models.Platform, game shared.Item,
	downloadType sum.Int[shared.ArtDownloadType], searchFilter string) DownloadArtScreen {
	return DownloadArtScreen{
		Platform:     platform,
		Game:         game,
		DownloadType: downloadType,
		SearchFilter: searchFilter,
	}
}

func (a DownloadArtScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.DownloadArt
}

func (a DownloadArtScreen) Draw() (value models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	exitCodeChan := make(chan int, 1)
	artChan := make(chan string, 1)

	args := []string{"--message", "Attempting to download art...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() > 6 {
		logger.Fatal("Error with starting miniui-presenter download art message", zap.Error(err))
	}

	time.Sleep(1000 * time.Millisecond)

	go func() {
		art := utils.FindArt(a.Platform, a.Game, a.DownloadType)
		if art == "" {
			logger.Error("Could not find art!")
			exitCodeChan <- 404
		} else {
			artChan <- art
			exitCodeChan <- 0
		}

		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with minui-presenter display of download message", zap.Error(err))
	}

	artPath := <-artChan
	exitCode = <-exitCodeChan

	switch exitCode {
	case 404:
		return shared.Item{}, 404, nil
	default:
		code, _ := cui.ShowMessageWithOptions("　　　　　　　　　　　　　　　　　　　　　　　　　", "0",
			"--background-image", artPath,
			"--confirm-text", "Use",
			"--confirm-show", "true",
			"--action-button", "X",
			"--action-text", "I'll Find My Own",
			"--action-show", "true",
			"--message-alignment", "bottom")

		if code == 2 || code == 4 {
			common.DeleteFile(artPath)
		}
		return shared.Item{
			Path: artPath,
		}, 0, nil
	}

}
