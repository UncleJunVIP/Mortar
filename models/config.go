package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
	"qlova.tech/sum"
)

type Config struct {
	Hosts              Hosts  `yaml:"hosts" json:"hosts"`
	UnzipDownloads     bool   `yaml:"unzip_downloads" json:"unzip_downloads"`
	GroupBinCue        bool   `yaml:"group_bin_cue" json:"group_bin_cue"`
	GroupMultiDisc     bool   `yaml:"group_multi_disc" json:"group_multi_disc"`
	DownloadArt        bool   `yaml:"download_art" json:"download_art"`
	RawArtDownloadType string `yaml:"art_download_type" json:"art_download_type"`
	LogLevel           string `yaml:"log_level" json:"log_level"`

	ArtDownloadType sum.Int[shared.ArtDownloadType] `yaml:"-" json:"-"`
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddArray("hosts", &c.Hosts)
	enc.AddBool("download_art", c.DownloadArt)
	enc.AddBool("unzip_downloads", c.UnzipDownloads)
	enc.AddBool("group_bin_cue", c.GroupBinCue)
	enc.AddBool("group_multi_disc", c.GroupMultiDisc)
	enc.AddString("art_download_type", c.ArtDownloadType.String())
	enc.AddString("log_level", c.LogLevel)

	return nil
}
