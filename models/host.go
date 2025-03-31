package models

import "qlova.tech/sum"

type Host struct {
	DisplayName string            `yaml:"display_name"`
	HostType    sum.Int[HostType] `yaml:"host_type"`
	RootURI     string            `yaml:"root_uri"`
	Port        int               `yaml:"port"`

	Username         string   `yaml:"username"`
	Password         string   `yaml:"password"`
	ShareName        string   `yaml:"share_name"`
	ExtensionFilters []string `yaml:"extension_filters"`

	Sections []Section `yaml:"sections"`
	Filters  []string  `yaml:"filters"`

	TableColumns       TableColumns      `yaml:"table_columns"`
	SourceReplacements map[string]string `yaml:"source_replacements"`

	SectionIndices map[string]int `yaml:"-"`
}

type HostType struct {
	APACHE,
	NGINX,
	SMB,
	RAPSCALLION,
	CUSTOM sum.Int[HostType]
}

var HostTypes = sum.Int[HostType]{}.Sum()

func (h Host) GetSectionIndices() map[string]int {
	if h.SectionIndices == nil {
		h.SectionIndices = map[string]int{}

		for idx, section := range h.Sections {
			h.SectionIndices[section.Name] = idx
		}
	}

	return h.SectionIndices
}
