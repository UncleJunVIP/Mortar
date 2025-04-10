package models

import (
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Hosts         Hosts  `yaml:"hosts"`
	ShowItemCount bool   `yaml:"show_item_count"`
	DownloadArt   bool   `yaml:"download_art"`
	LogLevel      string `yaml:"log_level"`
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddArray("hosts", &c.Hosts)
	enc.AddBool("show_item_count", c.ShowItemCount)
	enc.AddBool("download_art", c.DownloadArt)
	enc.AddString("log_level", c.LogLevel)

	return nil
}
