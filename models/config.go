package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"qlova.tech/sum"
)

type Config struct {
	Hosts              Hosts                           `yaml:"hosts"`
	RawArtDownloadType string                          `yaml:"art_download_type"`
	ArtDownloadType    sum.Int[shared.ArtDownloadType] `yaml:"-"`
	UnzipDownloads     bool                            `yaml:"unzip_downloads"`
	DownloadArt        bool                            `yaml:"download_art"`
	GroupBinCue        bool                            `yaml:"group_bin_cue"`
	GroupMultiDisc     bool                            `yaml:"group_multi_disc"`
	LogLevel           string                          `yaml:"log_level"`
}
