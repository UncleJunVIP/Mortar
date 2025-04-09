package utils

import (
	"fmt"
	"go.uber.org/atomic"
	"gopkg.in/yaml.v3"
	"mortar/models"
	"os"
	"sync"
)

var appState atomic.Pointer[models.AppState]
var onceAppState sync.Once

func LoadConfig() (*models.Config, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		return nil, fmt.Errorf("reading config.yml: %w", err)
	}

	var config models.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("parsing config.yml: %w", err)
	}

	return &config, nil
}

func GetAppState() *models.AppState {
	onceAppState.Do(func() {
		appState.Store(&models.AppState{})
	})
	return appState.Load()
}

func UpdateAppState(newAppState *models.AppState) {
	appState.Store(newAppState)
}

func SetHost(host models.Host) {
	temp := GetAppState()
	temp.CurrentHost = host
	UpdateAppState(temp)
}

func SetConfig(config *models.Config) {
	temp := GetAppState()
	temp.Config = config
	UpdateAppState(temp)
}

func SetSection(section models.Section) {
	temp := GetAppState()
	temp.CurrentSection = section
	UpdateAppState(temp)
}

func SetSearchFilter(filter string) {
	temp := GetAppState()
	temp.SearchFilter = filter
	UpdateAppState(temp)
}

func SetSelectedFile(file string) {
	temp := GetAppState()
	temp.SelectedFile = file
	UpdateAppState(temp)
}
