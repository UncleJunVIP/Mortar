package models

import (
	"go.uber.org/zap"
	"os"
	"qlova.tech/sum"
)

type AppState struct {
	Config      *Config
	HostIndices map[string]int

	CurrentHost      Host
	CurrentScreen    sum.Int[Screen]
	CurrentSection   Section
	CurrentItemsList []Item
	SearchFilter     string
	SelectedFile     string

	LogFile *os.File
	Logger  *zap.Logger
}
