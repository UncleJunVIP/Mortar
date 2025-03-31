package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"
	"log"
	"mortar/clients"
	"mortar/models"
	"os"
	"os/exec"
	"qlova.tech/sum"
	"strings"
	"time"
)

type AppState struct {
	Config      *models.Config
	HostIndices map[string]int

	CurrentHost      models.Host
	CurrentScreen    sum.Int[Screen]
	CurrentSection   models.Section
	CurrentItemsList []models.Item
	SearchFilter     string
	SelectedFile     string

	LogFile *os.File
	Logger  *zap.Logger
}

type Screen struct {
	MainMenu,
	SectionSelection,
	ItemList,
	Loading,
	SearchBox,
	Download sum.Int[Screen]
}

var Screens = sum.Int[Screen]{}.Sum()

var screenFuncs = map[sum.Int[Screen]]func() models.Selection{
	Screens.MainMenu:         mainMenuScreen,
	Screens.SectionSelection: sectionSelectionScreen,
	Screens.ItemList:         itemListScreen,
	Screens.Loading:          loadingScreen,
	Screens.SearchBox:        searchBox,
	Screens.Download:         downloadScreen,
}

var sugar *zap.SugaredLogger
var appState AppState

func init() {
	cwd, _ := os.Getwd()

	_ = os.Setenv("DEVICE", "brick")
	_ = os.Setenv("PLATFORM", "tg5040")
	_ = os.Setenv("PATH", cwd+"/bin/tg5040")
	_ = os.Setenv("LD_LIBRARY_PATH", "/mnt/SDCARD/.system/tg5040/lib:/usr/trimui/lib")

	// So users don't have to install TrimUI_EX
	_ = os.Setenv("SSL_CERT_DIR", cwd+"/certs")

	logFile, err := os.OpenFile(cwd+"/mortar.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		showMessage("Unable to open log file!", "3")
		log.Fatalf("Unable to open log file: %v", err)
	}

	writeSyncer := zapcore.AddSync(logFile)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		writeSyncer,
		zap.InfoLevel,
	)

	appState.Logger = zap.New(core)
	sugar = appState.Logger.Sugar()

	appState.Config = loadConfig()

	appState.HostIndices = make(map[string]int)
	for idx, host := range appState.Config.Hosts {
		appState.HostIndices[host.DisplayName] = idx
	}
}

func loadConfig() *models.Config {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		showMessage("Unable to load config.yml! Quitting!", "3")
		sugar.Fatal("Unable to load config.yml!", err.Error())
	}

	var config models.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		showMessage("Unable to parse config.yml! Quitting!", "3")
		sugar.Fatal("Unable to parse config.yml!", err.Error())
	}

	return &config
}

func buildClient(host models.Host) (models.Client, error) {
	switch host.HostType {
	case models.HostTypes.APACHE,
		models.HostTypes.RAPSCALLION,
		models.HostTypes.CUSTOM:
		return clients.NewHttpTableClient(
			host.RootURI,
			host.HostType,
			host.TableColumns,
			host.SourceReplacements,
			host.Filters,
		), nil
	case models.HostTypes.NGINX:
		return clients.NewNginxJsonClient(host.RootURI, host.Filters), nil
	case models.HostTypes.SMB:
		{
			return clients.NewSMBClient(
				host.RootURI,
				host.Port,
				host.Username,
				host.Password,
				host.ShareName,
				host.ExtensionFilters,
			)
		}
	}

	return nil, nil
}

func downloadList(cancel context.CancelFunc) error {
	defer cancel()

	client, err := buildClient(appState.CurrentHost)
	if err != nil {
		return err
	}

	defer func(client models.Client) {
		err := client.Close()
		if err != nil {
			sugar.Errorf("Unable to close client: %v", err)
		}
	}(client)

	items, err := client.ListDirectory(appState.CurrentSection.HostSubdirectory)
	if err != nil {
		sugar.Errorf("Unable to download listings: %v", err)
		return err
	}

	appState.CurrentItemsList = items

	return nil
}

func downloadFile(cancel context.CancelFunc) error {
	defer cancel()

	client, err := buildClient(appState.CurrentHost)
	if err != nil {
		return err
	}

	defer func(client models.Client) {
		err := client.Close()
		if err != nil {
			sugar.Errorf("Unable to close client: %v", err)
		}
	}(client)

	return client.DownloadFile(appState.CurrentSection.HostSubdirectory,
		appState.CurrentSection.LocalDirectory, appState.SelectedFile)
}

func filterList(itemList []models.Item, keywords ...string) []models.Item {
	var filteredItemList []models.Item

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

func displayMinUiList(list string, format string, title string) models.Selection {
	return displayMinUiListWithAction(list, format, title, "")
}

func displayMinUiListWithAction(list string, format string, title string, actionText string, extraArgs ...string) models.Selection {
	args := []string{"--format", format, "--title", title, "--file", "-"}

	if actionText != "" {
		args = append(args, "--action-button", "X", "--action-text", actionText)
	}

	if extraArgs != nil {
		args = append(args, extraArgs...)
	}

	cmd := exec.Command("minui-list", args...)
	cmd.Env = os.Environ()
	cmd.Env = os.Environ()

	var stdoutbuf, stderrbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
	cmd.Stderr = &stderrbuf

	cmd.Stdin = strings.NewReader(list)

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	if err := cmd.Run(); err != nil {
		return models.Selection{Code: cmd.ProcessState.ExitCode(), Error: fmt.Errorf("failed to run minui-list: %w", err)}
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	return models.Selection{Value: outValue, Code: cmd.ProcessState.ExitCode()}
}

func showMessage(message string, timeout string) {
	args := []string{"--message", message, "--timeout", timeout}
	cmd := exec.Command("minui-presenter", args...)
	err := cmd.Run()

	if err != nil && cmd.ProcessState.ExitCode() != 124 {
		sugar.Fatalf("failed to run minui-presenter: %v", err)
	}
}

func searchBox() models.Selection {
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
		sugar.Fatalf("failed to start minui-keyboard: %v", err)
	}

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() == 1 {
		sugar.Errorf("Error with keyboard: %s", stderrbuf.String())
		showMessage("Unable to open keyboard!", "3")
		return models.Selection{Code: 1}
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	return models.Selection{Value: strings.TrimSpace(outValue), Code: cmd.ProcessState.ExitCode()}
}

func mainMenuScreen() models.Selection {
	menu := ""

	var hosts []string
	for _, host := range appState.Config.Hosts {
		hosts = append(hosts, host.DisplayName)
	}

	menu = strings.Join(hosts, "\n")

	return displayMinUiList(menu, "text", "Brick & Mortar")
}

func sectionSelectionScreen() models.Selection {
	menu := ""

	var sections []string
	for _, section := range appState.CurrentHost.Sections {
		sections = append(sections, section.Name)
	}

	menu = strings.Join(sections, "\n")

	return displayMinUiList(menu, "text", appState.CurrentHost.DisplayName)
}

func itemListScreen() models.Selection {
	title := appState.CurrentSection.Name
	itemList := appState.CurrentItemsList

	var extraArgs []string

	if len(appState.CurrentHost.Filters) > 0 {
		itemList = filterList(itemList, appState.CurrentHost.Filters...)
	}

	if appState.SearchFilter != "" {
		title = title + "   [Search: \"" + appState.SearchFilter + "\"]"
		extraArgs = append(extraArgs, "--cancel-text", "CLEAR SEARCH")
		itemList = filterList(itemList, appState.SearchFilter)
	}

	if len(itemList) == 0 {
		return models.Selection{Code: 404}
	}

	var itemEntries []string
	for _, item := range itemList {
		itemEntries = append(itemEntries, item.Filename)
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

		showMessage(itemCountMessage, "3")
	}

	return displayMinUiListWithAction(strings.Join(itemEntries, "\n"), "text", title, "SEARCH", extraArgs...)
}

func downloadScreen() models.Selection {
	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{"--message", "Downloading " + appState.SelectedFile + "...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		sugar.Fatalf("Error with starting miniui-presenter download message: %s", err)
	}

	time.Sleep(1000 * time.Millisecond)

	exitCode := 0

	go func() {
		err := downloadFile(cancel)
		if err != nil {
			sugar.Errorf("Error downloading file: %s", err)
			exitCode = 1
		}
		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		sugar.Fatalf("Error with minui-presenter display of download message: %s", err)
	}

	return models.Selection{Code: exitCode}
}

func loadingScreen() models.Selection {
	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{"--message", "Loading " + appState.CurrentSection.Name + "...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		sugar.Fatalf("Error with starting miniui-presenter loading message: %s", err)
	}

	time.Sleep(1000 * time.Millisecond)

	exitCode := 0

	go func() {
		err := downloadList(cancel)
		if err != nil {
			sugar.Errorf("Error downloading Item List: %s", err)
			exitCode = 1
		}
		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		sugar.Fatalf("Error while waiting for miniui-presenter loading message to be killed: %s", err)
	}

	return models.Selection{Code: exitCode}
}

func cleanup() {
	appState.Logger.Sync()
	appState.LogFile.Close()
}

func main() {
	defer cleanup()

	for {
		selection := screenFuncs[appState.CurrentScreen]()

		switch appState.CurrentScreen {
		case Screens.MainMenu:
			{
				switch selection.Code {
				case 0:
					{
						appState.CurrentScreen = Screens.SectionSelection
						idx := appState.HostIndices[strings.TrimSpace(selection.Value)]
						appState.CurrentHost = appState.Config.Hosts[idx]
					}
				case 1, 2:
					{
						os.Exit(0)
					}
				}
			}

		case Screens.SectionSelection:
			{
				switch selection.Code {
				case 0:
					{
						appState.CurrentScreen = Screens.Loading
						idx := appState.CurrentHost.GetSectionIndices()[strings.TrimSpace(selection.Value)]
						appState.CurrentSection = appState.CurrentHost.Sections[idx]
					}
				case 1, 2:
					{
						appState.CurrentScreen = Screens.MainMenu
					}
				}
			}

		case Screens.ItemList:
			{
				switch selection.Code {
				case 0:
					appState.SelectedFile = strings.TrimSpace(selection.Value)
					appState.CurrentScreen = Screens.Download
				case 2:
					if appState.SearchFilter != "" {
						appState.SearchFilter = ""
					} else {
						appState.CurrentScreen = Screens.SectionSelection
					}
				case 4:
					appState.CurrentScreen = Screens.SearchBox
				case 404:
					{
						if appState.SearchFilter != "" {
							showMessage("No results found for \""+appState.SearchFilter+"\"", "3")
							appState.SearchFilter = ""
							appState.CurrentScreen = Screens.SearchBox
						} else {
							showMessage("This section contains no items", "3")
							appState.CurrentScreen = Screens.SectionSelection
						}
					}
				}
			}

		case Screens.Loading:
			{
				switch selection.Code {
				case 0:
					appState.CurrentScreen = Screens.ItemList
				case 1:
					showMessage("Unable to download item listing from source", "3")
					appState.CurrentScreen = Screens.MainMenu
				}
			}

		case Screens.SearchBox:
			{
				switch selection.Code {
				case 0:
					appState.SearchFilter = selection.Value
				case 1, 2, 3:
					appState.SearchFilter = ""
				}

				appState.CurrentScreen = Screens.ItemList
			}

		case Screens.Download:
			{
				switch selection.Code {
				case 1:
					showMessage("Unable to download "+appState.SelectedFile, "3")
				}

				appState.CurrentScreen = Screens.ItemList
			}
		}
	}
}
