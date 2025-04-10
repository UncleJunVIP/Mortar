package models

import (
	"go.uber.org/zap/zapcore"
	"qlova.tech/sum"
)

type AppState struct {
	Config      *Config
	HostIndices map[string]int

	CurrentHost      Host
	CurrentScreen    sum.Int[Screen]
	CurrentSection   Section
	CurrentItemsList Items
	SearchFilter     string
	SelectedFile     string

	LastSavedArtPath string
}

func (a AppState) MarshalLogObject(enc zapcore.ObjectEncoder) error {

	_ = enc.AddObject("config", a.Config)
	_ = enc.AddObject("host", &a.CurrentHost)
	enc.AddString("current_screen", a.CurrentScreen.String())
	_ = enc.AddObject("current_section", a.CurrentSection)
	_ = enc.AddArray("current_items_list", a.CurrentItemsList)
	enc.AddString("search_filter", a.SearchFilter)
	enc.AddString("selected_file", a.SelectedFile)
	enc.AddString("last_saved_art_path", a.LastSavedArtPath)

	return nil
}
