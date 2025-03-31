package models

type Config struct {
	Hosts         []Host `yaml:"hosts"`
	ShowItemCount bool   `yaml:"show_item_count"`
}
