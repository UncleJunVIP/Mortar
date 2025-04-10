package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"mortar/clients"
	"mortar/common"
	"mortar/models"
	"os"
	"os/exec"
	"strings"
)

func fetchList(cancel context.CancelFunc) error {
	defer cancel()

	logger := common.GetLoggerInstance()
	appState := common.GetAppState()

	logger.Debug("Fetching Item List",
		zap.Object("AppState", appState))

	client, err := clients.BuildClient(appState.CurrentHost)
	if err != nil {
		return err
	}

	defer func(client models.Client) {
		err := client.Close()
		if err != nil {
			logger.Error("Unable to close client", zap.Error(err))
		}
	}(client)

	items, err := client.ListDirectory(appState.CurrentSection)
	if err != nil {
		logger.Error("Unable to download listings", zap.Error(err))
		return err
	}

	appState.CurrentItemsList = items

	return nil
}

func filterList(itemList []models.Item, keywords ...string) []models.Item {
	var filteredItemList []models.Item

	for _, item := range itemList {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(item.Filename), strings.ToLower(keyword)) {
				filteredItemList = append(filteredItemList, item)
				break
			}
		}
	}

	return filteredItemList
}

func displayMinUiList(list string, format string, title string, options ...string) models.Selection {
	return displayMinUiListWithAction(list, format, title, "", options...)
}

func displayMinUiListWithAction(list string, format string, title string, actionText string, options ...string) models.Selection {
	args := []string{"--format", format, "--title", title, "--file", "-"}

	if actionText != "" {
		args = append(args, "--action-button", "X", "--action-text", actionText)
	}

	if options != nil {
		args = append(args, options...)
	}

	cmd := exec.Command("minui-list", args...)
	cmd.Env = os.Environ()
	cmd.Env = os.Environ()

	var stdoutbuf, stderrbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
	cmd.Stderr = &stderrbuf

	cmd.Stdin = strings.NewReader(list)

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	if err := cmd.Run(); err != nil {
		return models.Selection{Code: cmd.ProcessState.ExitCode(), Error: fmt.Errorf("failed to run minui-list: %w", err)}
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	return models.Selection{Value: outValue, Code: cmd.ProcessState.ExitCode()}
}
