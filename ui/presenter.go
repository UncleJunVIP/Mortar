package ui

import (
	"bytes"
	"go.uber.org/zap"
	"mortar/common"
	"os/exec"
	"strings"
)

func ShowMessage(message string, timeout string) {
	ShowMessageWithOptions(message, timeout)
}

func ShowMessageWithOptions(message string, timeout string, options ...string) int {
	args := []string{"--message", message, "--timeout", timeout}

	if options != nil {
		args = append(args, options...)
	}

	cmd := exec.Command("minui-presenter", args...)

	var stdoutbuf, stderrbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
	cmd.Stderr = &stderrbuf

	err := cmd.Run()

	if err != nil && cmd.ProcessState.ExitCode() == 1 {
		if !common.LoggerInitialized.Load() {
			common.LogStandardFatal("Failed to run minui-presenter", err)
		}

		logger := common.GetLoggerInstance()

		logger.Error("Failed to run minui-presenter",
			zap.String("command", strings.Join(cmd.Args, " ")),
			zap.Error(err),
			zap.String("output_out", stdoutbuf.String()),
			zap.String("output_err", stderrbuf.String()))
	}

	return cmd.ProcessState.ExitCode()
}
