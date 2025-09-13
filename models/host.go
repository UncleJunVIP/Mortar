package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"qlova.tech/sum"
)

type Host struct {
	DisplayName string                   `yaml:"display_name"`
	HostType    sum.Int[shared.HostType] `yaml:"host_type"`
	RootURI     string                   `yaml:"root_uri"`
	Port        int                      `yaml:"port"`

	Username string `yaml:"username"`
	Password string `yaml:"password"`

	Platforms Platforms `yaml:"platforms"`
	Filters   Filters   `yaml:"filters"`

	TableColumns       shared.TableColumns `yaml:"table_columns"`
	SourceReplacements SourceReplacements  `yaml:"source_replacements"`

	PlatformIndices PlatformIndices `yaml:"-"`
}

func (h Host) Value() interface{} {
	return h
}

type Hosts []Host

type Filters struct {
	InclusiveFilters []string `yaml:"inclusive_filters"`
	ExclusiveFilters []string `yaml:"exclusive_filters"`
}

type SourceReplacements map[string]string

type PlatformIndices map[string]int

func (h Host) GetPlatformIndices() PlatformIndices {
	if h.PlatformIndices == nil {
		h.PlatformIndices = map[string]int{}

		for idx, section := range h.Platforms {
			h.PlatformIndices[section.Name] = idx
		}
	}

	return h.PlatformIndices
}

func (h Hosts) Values() []string {
	var list []string
	for _, host := range h {
		list = append(list, host.DisplayName)
	}
	return list
}
