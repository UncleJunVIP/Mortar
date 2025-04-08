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
	"path/filepath"
	"qlova.tech/sum"
	"regexp"
	"strings"
	"time"
)

var Screens = sum.Int[models.Screen]{}.Sum()

var screenFuncs = map[sum.Int[models.Screen]]func() models.Selection{
	Screens.MainMenu:         mainMenuScreen,
	Screens.SectionSelection: sectionSelectionScreen,
	Screens.ItemList:         itemListScreen,
	Screens.Loading:          loadingScreen,
	Screens.SearchBox:        searchBox,
	Screens.Download:         downloadScreen,
	Screens.DownloadArt:      downloadArtScreen,
}

var sugar *zap.SugaredLogger
var appState models.AppState

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
		models.HostTypes.MEGATHREAD,
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
	case models.HostTypes.ROMM:
		{
			return clients.NewRomMClient(
				host.RootURI,
				host.Port,
				host.Username,
				host.Password,
			), nil
		}
	}

	return nil, nil
}

func fetchList(cancel context.CancelFunc) error {
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

	items, err := client.ListDirectory(appState.CurrentSection)
	if err != nil {
		sugar.Errorf("Unable to download listings: %v", err)
		return err
	}

	appState.CurrentItemsList = items

	return nil
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

func findArt() bool {
	re := regexp.MustCompile(`\((.*?)\)`)
	tag := re.FindStringSubmatch(appState.CurrentSection.LocalDirectory)

	if len(tag) < 2 {
		return false
	}

	client := clients.NewThumbnailClient()
	section := client.BuildThumbnailSection(tag[1])

	artList, err := client.ListDirectory(section)

	if err != nil {
		sugar.Infof("Unable to fetch artlist: %v", err)
		return false
	}

	noExtension := strings.TrimSuffix(appState.SelectedFile, filepath.Ext(appState.SelectedFile))

	var matched models.Item

	// naive search first
	for _, art := range artList {
		if strings.Contains(strings.ToLower(art.Filename), strings.ToLower(noExtension)) {
			matched = art
			break
		}
	}

	if matched.Filename == "" {
		// TODO Levenshtein Distance support at some point
	}

	if matched.Filename != "" {
		err = client.DownloadFileRename(section.HostSubdirectory,
			filepath.Join(appState.CurrentSection.LocalDirectory, ".media"), matched.Filename, appState.SelectedFile)

		if err != nil {
			return false
		}

		return true
	}

	return false
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

	var hostSubdirectory string

	if appState.CurrentHost.HostType == models.HostTypes.ROMM {
		var selectedItem models.Item
		for _, item := range appState.CurrentItemsList {
			if item.Filename == appState.SelectedFile {
				selectedItem = item
				break
			}
		}
		hostSubdirectory = selectedItem.RomID
	} else {
		hostSubdirectory = appState.CurrentSection.HostSubdirectory
	}

	return client.DownloadFile(hostSubdirectory,
		appState.CurrentSection.LocalDirectory, appState.SelectedFile)
}

func displayMinUiList(list string, format string, title string, extraArgs ...string) models.Selection {
	return displayMinUiListWithAction(list, format, title, "", extraArgs...)
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

func mainMenuScreen() models.Selection {
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
		err := fetchList(cancel)
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

func itemListScreen() models.Selection {
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

		if appState.Config.DownloadArt {
			findArt()
		}

		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		sugar.Fatalf("Error with minui-presenter display of download message: %s", err)
	}

	return models.Selection{Code: exitCode}
}

func downloadArtScreen() models.Selection {
	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{"--message", "Attempting to download art...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		sugar.Fatalf("Error with starting miniui-presenter download message: %s", err)
	}

	time.Sleep(1000 * time.Millisecond)

	exitCode := 0

	go func() {
		res := findArt()
		if !res {
			sugar.Errorf("Could not find art!: %s", err)
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

func cleanup() {
	appState.Logger.Sync()
	appState.LogFile.Close()
}

func main() {
	defer cleanup()

	if len(appState.Config.Hosts) == 1 {
		appState.CurrentScreen = Screens.SectionSelection
		appState.CurrentHost = appState.Config.Hosts[0]
	} else {
		appState.CurrentScreen = Screens.MainMenu
	}

	for {
		selection := screenFuncs[appState.CurrentScreen]()

		// Hacky way to handle bad input on deep sleep
		if strings.Contains(selection.Value, "SetRawBrightness") || strings.Contains(selection.Value, "nSetRawVolume") {
			continue
		}

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
						if len(appState.Config.Hosts) == 1 {
							os.Exit(0)
						}
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
				case 0:
					{
						if appState.Config.DownloadArt {
							appState.CurrentScreen = Screens.DownloadArt
						} else {
							appState.CurrentScreen = Screens.ItemList
						}
					}

				case 1:
					showMessage("Unable to download "+appState.SelectedFile, "3")
					appState.CurrentScreen = Screens.ItemList

				default:
					appState.CurrentScreen = Screens.ItemList
				}
			}

		case Screens.DownloadArt:
			{
				switch selection.Code {
				case 0:
					showMessage("Found art! :)", "3")
				case 1:
					showMessage("Could not find art :(", "3")
				}
				appState.CurrentScreen = Screens.ItemList
			}
		}
	}
}
