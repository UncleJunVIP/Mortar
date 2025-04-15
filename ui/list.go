package ui

import (
	"context"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
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

	logger.Debug("Fetching Item List",
		zap.Object("AppState", appState))

	client, err := clients.BuildClient(appState.CurrentHost)
	if err != nil {
		return err
	}

	defer func(client shared.Client) {
		err := client.Close()
		if err != nil {
			logger.Error("Unable to close client", zap.Error(err))
		}
	}(client)

	section := ""

	switch appState.CurrentHost.HostType {
	case shared.HostTypes.ROMM:
		section = appState.CurrentSection.RomMPlatformID
	default:
		section = appState.CurrentSection.HostSubdirectory
	}

	items, err := client.ListDirectory(section)
	if err != nil {
		return err
	}

	appState.CurrentItemsList = items

	return nil
}

func filterList(itemList []shared.Item, filters models.Filters) []shared.Item {
	var filteredItemList []shared.Item

	for _, item := range itemList {
		contains := false
		for _, keyword := range filters.ExclusiveFilters {
			if strings.Contains(strings.ToLower(item.Filename), strings.ToLower(keyword)) {
				contains = true
				break
			}
		}
		if !contains {
			filteredItemList = append(filteredItemList, item)
		}
	}

	for _, item := range itemList {
		for _, keyword := range filters.InclusiveFilters {
			if strings.Contains(strings.ToLower(item.Filename), strings.ToLower(keyword)) {
				filteredItemList = append(filteredItemList, item)
				break
			}
		}
	}

	return filteredItemList
}
