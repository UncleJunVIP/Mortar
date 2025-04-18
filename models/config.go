package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
	"qlova.tech/sum"
)

type Config struct {
	Hosts              Hosts                           `yaml:"hosts"`
	DownloadArt        bool                            `yaml:"download_art"`
	RawArtDownloadType string                          `yaml:"art_download_type"`
	ArtDownloadType    sum.Int[shared.ArtDownloadType] `yaml:"-"`
	LogLevel           string                          `yaml:"log_level"`
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddArray("hosts", &c.Hosts)
	enc.AddBool("download_art", c.DownloadArt)
	enc.AddString("art_download_type", c.ArtDownloadType.String())
	enc.AddString("log_level", c.LogLevel)

	return nil
}
