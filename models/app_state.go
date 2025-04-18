package models

import (
	"go.uber.org/zap/zapcore"
)

type AppState struct {
	Config      *Config
	HostIndices map[string]int
}

func (a AppState) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddObject("config", a.Config)
	return nil
}
