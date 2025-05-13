package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
)

type AppState struct {
	Config      *Config
	HostIndices map[string]int

	CurrentFullGamesList shared.Items
}

func (a AppState) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddObject("config", a.Config)
	return nil
}
