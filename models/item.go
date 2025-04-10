package models

import "go.uber.org/zap/zapcore"

type Item struct {
	Filename string `json:"filename"`
	FileSize string `json:"file_size"`
	Date     string `json:"date"`

	RomID  string `json:"-"` // For RomM Support
	ArtURL string `json:"-"` // For RomM Support
}

func (i Item) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("filename", i.Filename)
	enc.AddString("file_size", i.FileSize)
	enc.AddString("date", i.Date)
	enc.AddString("rom_id", i.RomID)
	enc.AddString("art_url", i.ArtURL)

	return nil
}

type Items []Item

func (i Items) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, item := range i {
		_ = enc.AppendObject(item)
	}

	return nil
}
