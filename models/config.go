package models

type Config struct {
	Host          Host `yaml:"host"`
	ShowItemCount bool `yaml:"show_item_count"`
}
