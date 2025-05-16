package state

import (
	"fmt"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
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

func SetConfig(config *models.Config) {
	temp := GetAppState()
	temp.Config = config

	temp.HostIndices = make(map[string]int)
	for idx, host := range temp.Config.Hosts {
		temp.HostIndices[host.DisplayName] = idx
	}

	UpdateAppState(temp)
}

func SetCurrentFullGamesList(games shared.Items) {
	temp := GetAppState()
	temp.CurrentFullGamesList = games
	UpdateAppState(temp)
}

func SetLastSelectedPosition(index, position int) {
	temp := GetAppState()
	temp.LastSelectedIndex = index
	temp.LastSelectedPosition = position
	UpdateAppState(temp)
}
