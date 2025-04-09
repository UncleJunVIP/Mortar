package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"mortar/models"
	"mortar/utils"
	"os"
	"os/exec"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
	"time"
)

var Screens = sum.Int[models.Screen]{}.Sum()

var ScreenFuncs = map[sum.Int[models.Screen]]func() models.Selection{
	Screens.MainMenu:         mainMenuScreen,
	Screens.SectionSelection: sectionSelectionScreen,
	Screens.ItemList:         itemListScreen,
	Screens.Loading:          loadingScreen,
	Screens.SearchBox:        searchBox,
	Screens.Download:         downloadScreen,
	Screens.DownloadArt:      downloadArtScreen,
}

func SetScreen(screen sum.Int[models.Screen]) {
	tempAppState := utils.GetAppState()
	tempAppState.CurrentScreen = screen
	utils.UpdateAppState(tempAppState)
}

func mainMenuScreen() models.Selection {
	appState := utils.GetAppState()

	menu := ""

	var hosts []string
	for _, host := range appState.Config.Hosts {
		hosts = append(hosts, host.DisplayName)
	}

	menu = strings.Join(hosts, "\n")

	var extraArgs []string
	extraArgs = append(extraArgs, "--cancel-text", "QUIT")

	return displayMinUiList(menu, "text", "Mortar", extraArgs...)
}

func sectionSelectionScreen() models.Selection {
	appState := utils.GetAppState()

	menu := ""

	var sections []string
	for _, section := range appState.CurrentHost.Sections {
		sections = append(sections, section.Name)
	}

	menu = strings.Join(sections, "\n")

	var extraArgs []string

	if len(appState.Config.Hosts) == 1 {
		extraArgs = append(extraArgs, "--cancel-text", "QUIT")
	}

	return displayMinUiList(menu, "text", appState.CurrentHost.DisplayName, extraArgs...)
}

func loadingScreen() models.Selection {
	logger := utils.GetLoggerInstance()
	appState := utils.GetAppState()

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

	return models.Selection{Code: exitCode}
}

func searchBox() models.Selection {
	logger := utils.GetLoggerInstance()

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
		ShowMessage("Unable to open keyboard!", "3")
		return models.Selection{Code: 1}
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	return models.Selection{Value: strings.TrimSpace(outValue), Code: cmd.ProcessState.ExitCode()}
}

func itemListScreen() models.Selection {
	appState := utils.GetAppState()

	title := appState.CurrentHost.DisplayName + " | " + appState.CurrentSection.Name
	itemList := appState.CurrentItemsList

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "DOWNLOAD")

	if len(appState.CurrentHost.Filters) > 0 {
		itemList = filterList(itemList, appState.CurrentHost.Filters...)
	}

	if appState.SearchFilter != "" {
		title = "[Search: \"" + appState.SearchFilter + "\"]"
		extraArgs = append(extraArgs, "--cancel-text", "CLEAR SEARCH")
		itemList = filterList(itemList, appState.SearchFilter)
	}

	if len(itemList) == 0 {
		return models.Selection{Code: 404}
	}

	var itemEntries []string
	for _, item := range itemList {
		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))
		itemEntries = append(itemEntries, itemName)
	}

	if len(itemEntries) > 500 {
		itemEntries = itemEntries[:500]
	}

	if appState.Config.ShowItemCount {
		p := message.NewPrinter(language.English)
		total := p.Sprintf("%d", len(itemEntries))

		itemCountMessage := fmt.Sprintf("%s Items Returned.", total)

		if len(itemEntries) > 500 {
			itemCountMessage = itemCountMessage + " Showing 500."
		}

		ShowMessage(itemCountMessage, "3")
	}

	return displayMinUiListWithAction(strings.Join(itemEntries, "\n"), "text", title, "SEARCH", extraArgs...)
}

func downloadScreen() models.Selection {
	logger := utils.GetLoggerInstance()
	appState := utils.GetAppState()

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
		err := downloadFile(cancel)
		if err != nil {
			logger.Error("Error downloading file: %s", zap.Error(err))
			exitCode = 1
		}

		if appState.Config.DownloadArt {
			findArt()
		}

		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with minui-presenter display of download message: %s", zap.Error(err))
	}

	return models.Selection{Code: exitCode}
}

func downloadArtScreen() models.Selection {
	logger := utils.GetLoggerInstance()

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
		res := findArt()
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

	return models.Selection{Code: exitCode}
}
