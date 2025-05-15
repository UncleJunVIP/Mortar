package ui

import (
	"encoding/json"
	"fmt"
	gabamod "github.com/UncleJunVIP/gabagool/models"
	gaba "github.com/UncleJunVIP/gabagool/ui"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"mortar/models"
	"mortar/state"
	"mortar/utils"
	"os"
	"path"
	"path/filepath"
	"qlova.tech/sum"
	"slices"
	"strings"
)

type GameList struct {
	Platform      models.Platform
	Games         shared.Items
	SearchFilter  string
	SelectedIndex int
}

func InitGamesList(platform models.Platform, games shared.Items, searchFilter string, selectedIndex int) GameList {
	var g shared.Items

	if len(games) > 0 {
		g = games
	} else {
		process, err := gaba.BlockingProcess(fmt.Sprintf("Loading %s...", platform.Name), func() (interface{}, error) {
			var err error
			g, err = loadGamesList(platform)
			return g, err
		})
		if err != nil {
			return GameList{}
		}

		g = process.Result.(shared.Items)
	}

	state.SetCurrentFullGamesList(g)

	return GameList{
		Platform:      platform,
		Games:         g,
		SearchFilter:  searchFilter,
		SelectedIndex: selectedIndex,
	}
}

func (gl GameList) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.GameList
}

func (gl GameList) Draw() (game interface{}, exitCode int, e error) {
	host := gl.Platform.Host
	title := gl.Platform.Name

	itemList := gl.Games

	if len(host.Filters.InclusiveFilters) > 0 || len(host.Filters.ExclusiveFilters) > 0 {
		itemList = filterList(gl.Games, host.Filters)
	}

	if gl.SearchFilter != "" {
		title = "[Search: \"" + gl.SearchFilter + "\"]"
		itemList = filterList(itemList, models.Filters{InclusiveFilters: []string{gl.SearchFilter}})
	}

	if len(itemList) == 0 {
		return nil, 404, nil
	}

	var itemEntries []gabamod.MenuItem
	for _, game := range itemList {
		itemEntries = append(itemEntries, gabamod.MenuItem{
			Text:     strings.ReplaceAll(game.Filename, filepath.Ext(game.Filename), ""),
			Selected: false,
			Focused:  false,
			Metadata: game,
		})
	}

	fhi := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Search"},
		{ButtonName: "SELECT", HelpText: "Multi-Select"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selectedIndex := gl.SelectedIndex

	if selectedIndex < 9 {
		selectedIndex = 0
	}

	selection, err := gaba.List(title, itemEntries,
		gaba.ListOptions{
			FooterHelpItems:   fhi,
			EnableAction:      true,
			EnableMultiSelect: true,
			EnableReordering:  false,
			SelectedIndex:     selectedIndex,
		})
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && !selection.Unwrap().Cancelled && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {

		var selections shared.Items
		for _, item := range selection.Unwrap().SelectedItems {
			selections = append(selections, item.Metadata.(shared.Item))
		}

		state.SetLastSelectedIndex(selection.Unwrap().SelectedIndex)

		return selections, 0, nil
	} else if selection.IsSome() && selection.Unwrap().ActionTriggered {
		return nil, 4, nil
	}

	return nil, 2, err
}

func loadGamesList(platform models.Platform) (games shared.Items, e error) {
	logger := common.GetLoggerInstance()

	cacheResults := checkCache(platform)
	if cacheResults != nil {
		return cacheResults, nil
	}

	items, err := FetchListStateless(platform)
	if err != nil {
		logger.Error("Error downloading Item List", zap.Error(err))
	}

	if len(items) == 0 {
		return nil, nil
	}

	slices.SortFunc(items, func(a, b shared.Item) int {
		return strings.Compare(strings.ToLower(a.Filename), strings.ToLower(b.Filename))
	})

	cache(platform, items)
	return items, nil

}

func checkCache(platform models.Platform) shared.Items {
	logger := common.GetLoggerInstance()

	cwd, err := os.Getwd()
	if err != nil {
		logger.Debug("Unable to get current working directory for loading cached Megathread", zap.Error(err))
		return nil
	}

	if platform.Host.HostType == shared.HostTypes.MEGATHREAD {
		cachePath := filepath.Join(cwd, ".cache", utils.CachedMegaThreadJsonFilename(platform.Host.DisplayName, platform.Name))
		if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			return nil
		}

		data, err := os.ReadFile(cachePath)
		if err != nil {
			logger.Debug("Unable to read cached Megathread JSON file", zap.Error(err))
			return nil
		}

		var items shared.Items
		err = json.Unmarshal(data, &items)
		if err != nil {
			logger.Debug("Unable to unmarshal cached Megathread JSON data", zap.Error(err))
			return nil
		}

		return items
	}

	return nil
}

func cache(platform models.Platform, gamesList shared.Items) {
	if platform.Host.HostType == shared.HostTypes.MEGATHREAD {
		logger := common.GetLoggerInstance()

		jsonData, err := json.Marshal(gamesList)
		if err != nil {
			logger.Debug("Unable to get marshal JSON for Megathread", zap.Error(err))
			return
		}

		cwd, err := os.Getwd()
		if err != nil {
			logger.Debug("Unable to get current working directory for caching Megathread", zap.Error(err))
			return
		}

		dir := filepath.Join(cwd, ".cache")
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Debug("Unable to make cache directory", zap.Error(err))
			return
		}

		filePath := path.Join(cwd, ".cache", utils.CachedMegaThreadJsonFilename(platform.Host.DisplayName, platform.Name))
		err = os.WriteFile(filePath, jsonData, 0644)
		if err != nil {
			logger.Debug("Unable to write JSON to file for Megathread", zap.Error(err))
			return
		}

		logger.Info("Cached Megathread Platform", zap.String("platform_name", platform.Name))
	}
}
