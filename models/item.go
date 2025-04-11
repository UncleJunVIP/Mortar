package models

import (
	sharedModels "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
)

type MortarItem struct {
	sharedModels.Item `yaml:",inline"`
	FileSize          string `json:"file_size"`
	Date              string `json:"date"`

	RomID  string `json:"-"` // For RomM Support
	ArtURL string `json:"-"` // For RomM Support
}

func (i MortarItem) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("filename", i.Filename)
	enc.AddString("file_size", i.FileSize)
	enc.AddString("date", i.Date)
	enc.AddString("rom_id", i.RomID)
	enc.AddString("art_url", i.ArtURL)

	return nil
}

type MortarItems []MortarItem

func (i MortarItems) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, item := range i {
		_ = enc.AppendObject(item)
	}

	return nil
}
