package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	commonUI "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"mortar/models"
	"mortar/state"
	"mortar/utils"
	"os"
	"os/exec"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
	"time"
)

var Screens = sum.Int[models.Screen]{}.Sum()

var ScreenFuncs = map[sum.Int[models.Screen]]func() (shared.ListSelection, error){
	Screens.MainMenu:         mainMenuScreen,
	Screens.SectionSelection: sectionSelectionScreen,
	Screens.ItemList:         itemListScreen,
	Screens.Loading:          loadingScreen,
	Screens.SearchBox:        searchBox,
	Screens.Download:         downloadScreen,
	Screens.DownloadArt:      downloadArtScreen,
}

func SetScreen(screen sum.Int[models.Screen]) {
	tempAppState := state.GetAppState()
	tempAppState.CurrentScreen = screen
	state.UpdateAppState(tempAppState)
}

func mainMenuScreen() (shared.ListSelection, error) {
	appState := state.GetAppState()

	var hosts shared.Items
	for _, host := range appState.Config.Hosts {
		hosts = append(hosts, shared.Item{
			DisplayName: host.DisplayName,
		})
	}

	var extraArgs []string
	extraArgs = append(extraArgs, "--cancel-text", "QUIT")

	return commonUI.DisplayList(hosts, "Mortar", "", extraArgs...)
}

func sectionSelectionScreen() (shared.ListSelection, error) {
	appState := state.GetAppState()

	var sections shared.Items
	for _, section := range appState.CurrentHost.Sections {
		sections = append(sections, shared.Item{
			DisplayName: section.Name,
		})
	}

	var extraArgs []string

	if len(appState.Config.Hosts) == 1 {
		extraArgs = append(extraArgs, "--cancel-text", "QUIT")
	}

	return ui.DisplayList(sections, appState.CurrentHost.DisplayName, "", extraArgs...)
}

func loadingScreen() (shared.ListSelection, error) {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{"--message", "Loading " + appState.CurrentSection.Name + "...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with starting miniui-presenter loading message", zap.Error(err))
	}

	time.Sleep(1000 * time.Millisecond)

	exitCode := 0

	go func() {
		err := fetchList(cancel)
		if err != nil {
			logger.Error("Error downloading Item List", zap.Error(err))
			exitCode = 1
		}
		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error while waiting for miniui-presenter loading message to be killed", zap.Error(err))
	}

	return shared.ListSelection{ExitCode: exitCode}, nil
}

func searchBox() (shared.ListSelection, error) {
	logger := common.GetLoggerInstance()

	args := []string{"--title", "Mortar Search"}

	cmd := exec.Command("minui-keyboard", args...)
	cmd.Env = os.Environ()
	cmd.Env = os.Environ()

	var stdoutbuf, stderrbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
	cmd.Stderr = &stderrbuf

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	err := cmd.Start()
	if err != nil {
		logger.Fatal("failed to start minui-keyboard", zap.Error(err))
	}

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() == 1 {
		logger.Error("Error with keyboard", zap.String("error", stderrbuf.String()))
		_, _ = commonUI.ShowMessage("Unable to open keyboard!", "3")
		return shared.ListSelection{ExitCode: 1}, nil
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	return shared.ListSelection{Value: strings.TrimSpace(outValue), ExitCode: cmd.ProcessState.ExitCode()}, nil
}

func itemListScreen() (shared.ListSelection, error) {
	appState := state.GetAppState()

	title := appState.CurrentHost.DisplayName + " | " + appState.CurrentSection.Name
	itemList := appState.CurrentItemsList

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "DOWNLOAD")

	if len(appState.CurrentHost.Filters.InclusiveFilters) > 0 || len(appState.CurrentHost.Filters.ExclusiveFilters) > 0 {
		itemList = filterList(itemList, appState.CurrentHost.Filters)
	}

	if appState.SearchFilter != "" {
		title = "[Search: \"" + appState.SearchFilter + "\"]"
		extraArgs = append(extraArgs, "--cancel-text", "CLEAR SEARCH")
		itemList = filterList(itemList, models.Filters{InclusiveFilters: []string{appState.SearchFilter}})
	}

	if len(itemList) == 0 {
		return shared.ListSelection{ExitCode: 404}, nil
	}

	var itemEntries shared.Items
	for _, item := range itemList {
		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))
		itemEntries = append(itemEntries, shared.Item{DisplayName: strings.TrimSpace(itemName)})
	}

	if len(itemEntries) > 500 {
		itemEntries = itemEntries[:500]
	}

	if appState.Config.ShowItemCount {
		p := message.NewPrinter(language.English)
		total := p.Sprintf("%d", len(itemEntries))

		itemCountMessage := fmt.Sprintf("%s MortarItems Returned.", total)

		if len(itemEntries) > 500 {
			itemCountMessage = itemCountMessage + " Showing 500."
		}

		_, _ = commonUI.ShowMessage(itemCountMessage, "3")
	}

	return ui.DisplayList(itemEntries, title, "SEARCH", extraArgs...)
}

func downloadScreen() (shared.ListSelection, error) {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{"--message", "Downloading " + appState.SelectedFile + "...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with starting miniui-presenter download message", zap.Error(err))
	}

	time.Sleep(1000 * time.Millisecond)

	exitCode := 0

	go func() {
		_, err := utils.DownloadFile(cancel)
		if err != nil {
			logger.Error("Error downloading file: %s", zap.Error(err))
			exitCode = 1
		}

		if appState.Config.DownloadArt {
			utils.FindArt()
		}

		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with minui-presenter display of download message: %s", zap.Error(err))
	}

	return shared.ListSelection{ExitCode: exitCode}, nil
}

func downloadArtScreen() (shared.ListSelection, error) {
	logger := common.GetLoggerInstance()

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{"--message", "Attempting to download art...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with starting miniui-presenter download message", zap.Error(err))
	}

	time.Sleep(1000 * time.Millisecond)

	exitCode := 0

	go func() {
		res := utils.FindArt()
		if !res {
			logger.Error("Could not find art!", zap.Error(err))
			exitCode = 1
		}

		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with minui-presenter display of download message", zap.Error(err))
	}

	return shared.ListSelection{ExitCode: exitCode}, nil
}
