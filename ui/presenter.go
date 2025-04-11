package ui

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
)

func ShowMessage(message string, timeout string) {
	ShowMessageWithOptions(message, timeout)
}

func ShowMessageWithOptions(message string, timeout string, options ...string) int {
	exitCode, err := ui.ShowMessageWithOptions(message, timeout, options...)

	if err != nil && exitCode == 1 {
		if !common.LoggerInitialized.Load() {
			common.LogStandardFatal("Failed to run minui-presenter", err)
		}

		logger := common.GetLoggerInstance()

		logger.Error("Failed to run minui-presenter",
			zap.Error(err),
		)
	}

	return exitCode
}
