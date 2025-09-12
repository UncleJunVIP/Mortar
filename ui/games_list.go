package ui

import (
	"encoding/json"
	"fmt"
	"mortar/models"
	"mortar/state"
	"mortar/utils"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"qlova.tech/sum"
)

type GameList struct {
	Platform     models.Platform
	Games        shared.Items
	SearchFilter string
}

func InitGamesList(platform models.Platform, games shared.Items, searchFilter string) GameList {
	var g shared.Items

	if len(games) > 0 {
		g = games
	} else {
		process, err := gabagool.ProcessMessage(fmt.Sprintf("Loading %s...", platform.Name), gabagool.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
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
		Platform:     platform,
		Games:        g,
		SearchFilter: searchFilter,
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
		if gl.SearchFilter != "" {
			gabagool.ProcessMessage(
				fmt.Sprintf("No results found for \"%s\"", gl.SearchFilter),
				gabagool.ProcessMessageOptions{ShowThemeBackground: true},
				func() (interface{}, error) {
					time.Sleep(time.Second * 2)
					return nil, nil
				},
			)
		} else {
			gabagool.ProcessMessage(
				fmt.Sprintf("No games found for %s", gl.Platform.Name),
				gabagool.ProcessMessageOptions{ShowThemeBackground: true},
				func() (interface{}, error) {
					time.Sleep(time.Second * 2)
					return nil, nil
				},
			)
		}
		return nil, 404, nil
	}

	var itemEntries []gabagool.MenuItem
	for _, game := range itemList {
		itemEntries = append(itemEntries, gabagool.MenuItem{
			Text:     strings.ReplaceAll(game.Filename, filepath.Ext(game.Filename), ""),
			Selected: false,
			Focused:  false,
			Metadata: game,
		})
	}

	options := gabagool.DefaultListOptions(title, itemEntries)
	options.EnableAction = true
	options.EnableMultiSelect = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Search"},
		{ButtonName: "Select", HelpText: "Multi"},
		{ButtonName: "A", HelpText: "Select"},
	}
	options.SelectedIndex = state.GetAppState().LastSelectedIndex
	options.VisibleStartIndex = max(0, state.GetAppState().LastSelectedIndex-state.GetAppState().LastSelectedPosition)

	selection, err := gabagool.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {

		var selections shared.Items
		for _, item := range selection.Unwrap().SelectedItems {
			selections = append(selections, item.Metadata.(shared.Item))
		}

		state.SetLastSelectedPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)

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
		logger.Error("Error downloading Item List", "error", err)
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
		logger.Debug("Unable to get current working directory for loading cached Megathread", "error", err)
		return nil
	}

	if platform.Host.HostType == shared.HostTypes.MEGATHREAD {
		cachePath := filepath.Join(cwd, ".cache", utils.CachedMegaThreadJsonFilename(platform.Host.DisplayName, platform.Name))
		if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			return nil
		}

		data, err := os.ReadFile(cachePath)
		if err != nil {
			logger.Debug("Unable to read cached Megathread JSON file", "error", err)
			return nil
		}

		var items shared.Items
		err = json.Unmarshal(data, &items)
		if err != nil {
			logger.Debug("Unable to unmarshal cached Megathread JSON data", "error", err)
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
			logger.Debug("Unable to get marshal JSON for Megathread", "error", err)
			return
		}

		cwd, err := os.Getwd()
		if err != nil {
			logger.Debug("Unable to get current working directory for caching Megathread", "error", err)
			return
		}

		dir := filepath.Join(cwd, ".cache")
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Debug("Unable to make cache directory", "error", err)
			return
		}

		filePath := path.Join(cwd, ".cache", utils.CachedMegaThreadJsonFilename(platform.Host.DisplayName, platform.Name))
		err = os.WriteFile(filePath, jsonData, 0644)
		if err != nil {
			logger.Debug("Unable to write JSON to file for Megathread", "error", err)
			return
		}

		logger.Info("Cached Megathread Platform", "platform_name", platform.Name)
	}
}
