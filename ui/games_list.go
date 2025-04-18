package ui

import (
	"context"
	"encoding/json"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"mortar/models"
	"mortar/utils"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
	"time"
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
		g, _ = loadGamesList(platform)
	}

	return GameList{
		Platform:     platform,
		Games:        g,
		SearchFilter: searchFilter,
	}
}

func (gl GameList) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.GameList
}

func (gl GameList) Draw() (game models.ScreenReturn, exitCode int, e error) {
	host := gl.Platform.Host
	title := host.DisplayName + " | " + gl.Platform.Name

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "DOWNLOAD")

	itemList := gl.Games

	if len(host.Filters.InclusiveFilters) > 0 || len(host.Filters.ExclusiveFilters) > 0 {
		itemList = filterList(gl.Games, host.Filters)
	}

	if gl.SearchFilter != "" {
		title = "[Search: \"" + gl.SearchFilter + "\"]"
		extraArgs = append(extraArgs, "--cancel-text", "CLEAR SEARCH")
		itemList = filterList(itemList, models.Filters{InclusiveFilters: []string{gl.SearchFilter}})
	}

	if len(itemList) == 0 {
		return shared.Item{}, 404, nil
	}

	var itemEntries shared.Items
	itemEntriesMap := make(map[string]shared.Item)

	for _, item := range itemList {
		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))
		itemEntries = append(itemEntries, shared.Item{DisplayName: strings.TrimSpace(itemName)})
		itemEntriesMap[itemName] = item
	}

	selection, err := cui.DisplayList(itemEntries, title, "SEARCH", extraArgs...)
	if err != nil {
		return shared.Item{}, 1, err
	}

	selectedGame := itemEntriesMap[selection.Value]
	return selectedGame, selection.ExitCode, nil
}

func loadGamesList(platform models.Platform) (games shared.Items, e error) {
	logger := common.GetLoggerInstance()

	cacheResults := checkCache(platform)
	if cacheResults != nil {
		return cacheResults, nil
	}

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	exitCodeChan := make(chan int, 1)
	itemsChan := make(chan shared.Items, 1)
	errChan := make(chan error, 1)

	args := []string{"--message", "Loading " + platform.Name + "...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with starting miniui-presenter loading message", zap.Error(err))
	}

	time.Sleep(1000 * time.Millisecond)

	go func() {
		items, err := FetchListStateless(platform, cancel)
		if err != nil {
			logger.Error("Error downloading Item List", zap.Error(err))
			exitCodeChan <- 1
			errChan <- err
		} else {
			itemsChan <- items
			exitCodeChan <- 0
		}
		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error while waiting for miniui-presenter loading message to be killed", zap.Error(err))
	}

	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	select {
	case items := <-itemsChan:
		cache(platform, items)
		return items, nil
	default:
		return nil, nil
	}

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
