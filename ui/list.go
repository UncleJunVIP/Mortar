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
	var filteredItemListInclusive []shared.Item

	for _, item := range itemList {
		for _, filter := range filters.InclusiveFilters {
			if strings.Contains(strings.ToLower(item.Filename), strings.ToLower(filter)) {
				filteredItemListInclusive = append(filteredItemListInclusive, item)
				break
			}
		}
	}

	var filteredItemListExclusive []shared.Item

	for _, item := range filteredItemListInclusive {
		contains := false
		for _, filter := range filters.ExclusiveFilters {
			if strings.Contains(strings.ToLower(item.Filename), strings.ToLower(filter)) {
				contains = true
				break
			}
		}
		if !contains {
			filteredItemListExclusive = append(filteredItemListExclusive, item)
		}
	}

	return filteredItemListExclusive
}
