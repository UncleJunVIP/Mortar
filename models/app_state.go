package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
)

type AppState struct {
	Config      *Config
	HostIndices map[string]int

	CurrentFullGamesList shared.Items
	LastSelectedIndex    int
	LastSelectedPosition int
}
