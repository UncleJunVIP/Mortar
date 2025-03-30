package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	converthtmltabletodata "github.com/activcoding/HTML-Table-to-JSON"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"mortar/models"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"qlova.tech/sum"
	"strconv"
	"strings"
	"time"
)

type AppState struct {
	Config           *models.Config
	CurrentScreen    sum.Int[Screen]
	CurrentSection   string
	SearchFilter     string
	CurrentItemsList []models.Item
	SectionIndices   map[string]int
	SelectedFile     string
	LogFile          *os.File
}

type Screen struct {
	MainMenu,
	ItemListScreen,
	LoadingScreen,
	SearchBox,
	DownloadScreen sum.Int[Screen]
}

var Screens = sum.Int[Screen]{}.Sum()

var screenFuncs = map[sum.Int[Screen]]func() models.Selection{
	Screens.MainMenu:       mainMenuScreen,
	Screens.ItemListScreen: itemListScreen,
	Screens.LoadingScreen:  loadingScreen,
	Screens.SearchBox:      searchBox,
	Screens.DownloadScreen: downloadScreen,
}

var sugar *zap.SugaredLogger
var appState AppState

func init() {
	cwd, _ := os.Getwd()

	_ = os.Setenv("DEVICE", "brick")
	_ = os.Setenv("PLATFORM", "tg5040")
	_ = os.Setenv("PATH", cwd+"/bin/tg5040")
	_ = os.Setenv("LD_LIBRARY_PATH", "/mnt/SDCARD/.system/tg5040/lib:/usr/trimui/lib")

	// So users don't have to install TrimUI_Ex
	_ = os.Setenv("SSL_CERT_DIR", cwd+"/certs")

	logFile, err := os.OpenFile(cwd+"/mortar.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
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

	logger := zap.New(core)
	defer logger.Sync()

	sugar = logger.Sugar()

	appState.Config = loadConfig()

	appState.SectionIndices = make(map[string]int)

	for idx, section := range appState.Config.Host.Sections {
		appState.SectionIndices[section.Name] = idx
	}

	sugar.Info(appState.Config)
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

func currentSection() models.Section {
	idx := appState.SectionIndices[appState.CurrentSection]
	return appState.Config.Host.Sections[idx]
}

func parseTableForUrl(tableURL string) ([]models.Item, error) {
	params := url.Values{}

	switch appState.Config.Host.HostType {
	case models.HostTypes.APACHE:
		params.Add("F", "2") // To enable table mode for mod_autoindex
	}

	u, err := url.Parse(tableURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse table URL: %v", err)
	}
	u.RawQuery = params.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch table: %v", err)
	}
	defer resp.Body.Close()

	jsonBytes, err := converthtmltabletodata.ConvertReaderToJSON(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse table into json: %v", err)
	}

	rawJson := string(jsonBytes)

	cleaned := rawJson

	switch appState.Config.Host.HostType {
	case models.HostTypes.APACHE:
		cleaned = strings.ReplaceAll(cleaned, "[[", "[")
		cleaned = strings.ReplaceAll(cleaned, "]]", "]")
		cleaned = strings.ReplaceAll(cleaned, "Name", "filename")
		cleaned = strings.ReplaceAll(cleaned, "Size", "file_size")
		cleaned = strings.ReplaceAll(cleaned, "Last modified", "date")
	case models.HostTypes.RAPSCALLION:
		{
			cleaned = strings.ReplaceAll(cleaned, "  ↓", "")
			cleaned = strings.ReplaceAll(cleaned, "[[", "[")
			cleaned = strings.ReplaceAll(cleaned, "]]", "]")
			cleaned = strings.ReplaceAll(cleaned, "File Name", "filename")
			cleaned = strings.ReplaceAll(cleaned, "File Size", "file_size")
			cleaned = strings.ReplaceAll(cleaned, "Date", "date")
		}
	case models.HostTypes.CUSTOM:
		{
			for oldValue, newValue := range appState.Config.Host.SourceReplacements {
				cleaned = strings.ReplaceAll(cleaned, oldValue, newValue)
			}

			cleaned = strings.ReplaceAll(cleaned, appState.Config.Host.TableColumns.FilenameHeader, "filename")
			cleaned = strings.ReplaceAll(cleaned, appState.Config.Host.TableColumns.FileSizeHeader, "file_size")
			cleaned = strings.ReplaceAll(cleaned, appState.Config.Host.TableColumns.DateHeader, "date")
		}

	}

	var items []models.Item
	err = json.Unmarshal([]byte(cleaned), &items)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %v", err)
	}

	// Skip the header row(s)
	switch appState.Config.Host.HostType {
	case models.HostTypes.APACHE,
		models.HostTypes.RAPSCALLION:
		{
			if len(items) > 1 {
				return items[1:], nil
			}
		}
	}

	return nil, errors.New("wtf")
}

func parseJSONForUrl(jsonURL string) ([]models.Item, error) {
	resp, err := http.Get(jsonURL)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch json: %v", err)
	}
	defer resp.Body.Close()

	switch appState.Config.Host.HostType {
	case models.HostTypes.NGINX:
		{
			var nginxItems []models.NginxDirectoryListing
			decoder := json.NewDecoder(resp.Body)
			if err := decoder.Decode(&nginxItems); err != nil {
				return nil, fmt.Errorf("unable to decode nginx json: %v", err)
			}

			var items []models.Item
			for _, nginxItem := range nginxItems {
				items = append(items, models.Item{
					Filename: nginxItem.Filename,
					FileSize: strconv.FormatInt(nginxItem.Size, 10),
					Date:     nginxItem.ModifiedTime,
				})
			}

			return items, nil
		}
	default:
		showMessage("Invalid host type!", "3")
		sugar.Fatal("Invalid host type!")
	}

	return nil, nil
}

func buildSourceURL() string {
	rootURL := appState.Config.Host.RootURL
	subdirectory := currentSection().HostSubdirectory

	return rootURL + subdirectory
}

func downloadList(cancel context.CancelFunc) error {
	defer cancel()

	sourceURL := buildSourceURL()

	var items []models.Item
	var err error

	switch appState.Config.Host.HostType {
	case models.HostTypes.NGINX:
		items, _ = parseJSONForUrl(sourceURL)
	default:
		items, err = parseTableForUrl(sourceURL)
	}

	if err != nil {
		sugar.Fatalf("Error while parsing list: %v", err)
	}

	appState.CurrentItemsList = items

	return nil
}

func downloadFile(cancel context.CancelFunc) error {
	defer cancel()

	sourceURL := buildSourceURL() + appState.SelectedFile
	resp, err := http.Get(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	localDirectory := currentSection().LocalDirectory

	outFile, err := os.Create(localDirectory + appState.SelectedFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

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
	args := []string{"--header", "Mortar Search"}

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

	switch appState.Config.Host.HostType {
	case models.HostTypes.APACHE, models.HostTypes.CADDY,
		models.HostTypes.NGINX, models.HostTypes.RAPSCALLION, models.HostTypes.CUSTOM:
		var sections []string
		for _, section := range appState.Config.Host.Sections {
			sections = append(sections, section.Name)
		}
		menu = strings.Join(sections, "\n")
	}

	return displayMinUiList(menu, "text", "Brick & Mortar")
}

func itemListScreen() models.Selection {
	title := appState.CurrentSection
	itemList := appState.CurrentItemsList

	var extraArgs []string

	if len(appState.Config.Host.Filters) > 0 {
		itemList = filterList(itemList, appState.Config.Host.Filters...)
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

	args := []string{"--message", "Loading " + appState.CurrentSection + "...", "--timeout", "-1"}
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

func drawScreen() models.Selection {
	return screenFuncs[appState.CurrentScreen]()
}

func main() {
	defer appState.LogFile.Close()

	for {
		selection := drawScreen()

		switch appState.CurrentScreen {
		case Screens.MainMenu:
			{
				switch selection.Code {
				case 0:
					{
						appState.CurrentScreen = Screens.LoadingScreen
						appState.CurrentSection = strings.TrimSpace(selection.Value)
					}
				case 1, 2:
					{
						os.Exit(0)
					}
				}
			}
		case Screens.ItemListScreen:
			{
				switch selection.Code {
				case 0:
					appState.SelectedFile = strings.TrimSpace(selection.Value)
					appState.CurrentScreen = Screens.DownloadScreen
				case 2:
					if appState.SearchFilter != "" {
						appState.SearchFilter = ""
					} else {
						appState.CurrentScreen = Screens.MainMenu
					}
				case 4:
					appState.CurrentScreen = Screens.SearchBox
				case 404:
					{
						showMessage("No results found for \""+appState.SearchFilter+"\"", "3")
						appState.SearchFilter = ""
						appState.CurrentScreen = Screens.SearchBox
					}
				}
			}

		case Screens.LoadingScreen:
			{
				switch selection.Code {
				case 0:
					appState.CurrentScreen = Screens.ItemListScreen
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

				appState.CurrentScreen = Screens.ItemListScreen
			}
		case Screens.DownloadScreen:
			{
				switch selection.Code {
				case 1:
					showMessage("Unable to download "+appState.SelectedFile, "3")
				}

				appState.CurrentScreen = Screens.ItemListScreen
			}
		}
	}
}
