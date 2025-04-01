package models

type Item struct {
	Filename string `json:"filename"`
	FileSize string `json:"file_size"`
	Date     string `json:"date"`

	RomID string `json:"-"` // For RomM Support
}
