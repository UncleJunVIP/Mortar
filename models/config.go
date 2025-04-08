package models

type Config struct {
	Hosts         []Host `yaml:"hosts"`
	ShowItemCount bool   `yaml:"show_item_count"`
	DownloadArt   bool   `yaml:"download_art"`
}
