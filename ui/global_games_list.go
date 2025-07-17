package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"mortar/models"
	"mortar/state"
	"mortar/utils"
	"qlova.tech/sum"
	"slices"
	"strings"
	"time"
)

type GlobalGameList struct {
	Host         models.Host
	Games        shared.Items
	SearchFilter string
}

func InitGlobalGamesList(host models.Host, games shared.Items, searchFilter string) GlobalGameList {
	var g shared.Items

	if len(games) > 0 {
		g = games
	} else {
		process, err := gabagool.ProcessMessage(fmt.Sprintf("Loading %s...", host.DisplayName), gabagool.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
			var err error
			g, err = loadHostGamesList(host)
			return g, err
		})
		if err != nil {
			return GlobalGameList{}
		}

		g = process.Result.(shared.Items)
	}

	state.SetCurrentFullGamesList(g)

	return GlobalGameList{
		Host:         host,
		Games:        g,
		SearchFilter: searchFilter,
	}
}

func (ggl GlobalGameList) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.GlobalGamesList
}

func (ggl GlobalGameList) Draw() (game interface{}, exitCode int, e error) {
	host := ggl.Host
	title := fmt.Sprintf("%s All Games", host.DisplayName)

	itemList := ggl.Games

	if len(host.Filters.InclusiveFilters) > 0 || len(host.Filters.ExclusiveFilters) > 0 {
		itemList = utils.FilterList(ggl.Games, host.Filters)
	}

	if ggl.SearchFilter != "" {
		title = "[Search: \"" + ggl.SearchFilter + "\"]"
		itemList = utils.FilterList(itemList, models.Filters{InclusiveFilters: []string{ggl.SearchFilter}})
	}

	if len(itemList) == 0 {
		if ggl.SearchFilter != "" {
			gabagool.ProcessMessage(
				fmt.Sprintf("No results found for \"%s\"", ggl.SearchFilter),
				gabagool.ProcessMessageOptions{ShowThemeBackground: true},
				func() (interface{}, error) {
					time.Sleep(time.Second * 2)
					return nil, nil
				},
			)
		} else {
			gabagool.ProcessMessage(
				fmt.Sprintf("No games found for %s", ggl.Host.DisplayName),
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
			Text:     fmt.Sprintf("[%s] %s", game.Tag, game.DisplayName),
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

func loadHostGamesList(host models.Host) (games shared.Items, e error) {
	var items shared.Items

	for _, platform := range host.Platforms {
		platform.Host = host
		pi, err := loadGamesList(platform)
		if err == nil {
			items = append(items, pi...)
		}
	}

	slices.SortFunc(items, func(a, b shared.Item) int {
		return strings.Compare(strings.ToLower(a.DisplayName), strings.ToLower(b.DisplayName))
	})

	return items, nil

}
