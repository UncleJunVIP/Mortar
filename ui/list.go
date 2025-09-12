package ui

import (
	"encoding/json"
	"mortar/clients"
	"mortar/models"
	"mortar/utils"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
)

func FetchListStateless(platform models.Platform) (shared.Items, error) {
	logger := common.GetLoggerInstance()

	logger.Debug("Fetching Item List",
		"host", platform.Host)

	client, err := clients.BuildClient(platform.Host)
	if err != nil {
		return nil, err
	}

	defer func(client shared.Client) {
		err := client.Close()
		if err != nil {
			logger.Error("Unable to close client", "error", err)
		}
	}(client)

	subdirectory := ""

	switch platform.Host.HostType {
	case shared.HostTypes.ROMM:
		subdirectory = platform.RomMPlatformID
	default:
		subdirectory = platform.HostSubdirectory
	}

	items, err := client.ListDirectory(subdirectory)
	if err != nil {
		return nil, err
	}

	for i, item := range items {
		items[i].DisplayName = strings.ReplaceAll(item.Filename, filepath.Ext(item.Filename), "")
	}

	if platform.Host.HostType == shared.HostTypes.MEGATHREAD {
		jsonData, err := json.Marshal(items)
		if err != nil {
			logger.Debug("Unable to get marshal JSON for Megathread", "error", err)

			cwd, err := os.Getwd()
			if err != nil {
				logger.Debug("Unable to get current working directory for caching Megathread", "error", err)
			}

			filePath := path.Join(cwd, ".cache", utils.CachedMegaThreadJsonFilename("", ""))
			err = os.WriteFile(filePath, jsonData, 0644)
			if err != nil {
				logger.Debug("Unable to write JSON to file for Megathread", "error", err)
			}
		}
	}

	return items, nil
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
