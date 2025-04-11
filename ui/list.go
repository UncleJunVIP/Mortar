package ui

import (
	"context"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"go.uber.org/zap"
	"mortar/clients"
	"mortar/models"
	"mortar/state"
	"strings"
)

func fetchList(cancel context.CancelFunc) error {
	defer cancel()

	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	logger.Debug("Fetching MortarItem List",
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
		return err
	}

	appState.CurrentItemsList = items

	return nil
}

func filterList(itemList []models.MortarItem, keywords ...string) []models.MortarItem {
	var filteredItemList []models.MortarItem

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
