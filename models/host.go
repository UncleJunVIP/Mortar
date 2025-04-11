package models

import (
	sharedModels "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
	"qlova.tech/sum"
	"strconv"
	"strings"
)

type Host struct {
	DisplayName string                         `yaml:"display_name"`
	HostType    sum.Int[sharedModels.HostType] `yaml:"host_type"`
	RootURI     string                         `yaml:"root_uri"`
	Port        int                            `yaml:"port"`

	Username string `yaml:"username"`
	Password string `yaml:"password"`

	ShareName        string   `yaml:"share_name"`
	ExtensionFilters []string `yaml:"extension_filters"`

	Sections MortarSections `yaml:"sections"`
	Filters  Filters        `yaml:"filters"`

	TableColumns       sharedModels.TableColumns `yaml:"table_columns"`
	SourceReplacements SourceReplacements        `yaml:"source_replacements"`

	SectionIndices SectionIndices `yaml:"-"`
}

type Hosts []Host

type Filters []string

func (f Filters) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, filter := range f {
		enc.AppendString(filter)
	}

	return nil
}

type SourceReplacements map[string]string

func (s SourceReplacements) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range s {
		enc.AddString(k, v)
	}

	return nil
}

type SectionIndices map[string]int

func (s SectionIndices) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range s {
		enc.AddInt(k, v)
	}

	return nil
}

func (h *Host) GetSectionIndices() map[string]int {
	if h.SectionIndices == nil {
		h.SectionIndices = map[string]int{}

		for idx, section := range h.Sections {
			h.SectionIndices[section.Name] = idx
		}
	}

	return h.SectionIndices
}

func (h *Hosts) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, host := range *h {
		_ = enc.AppendObject(&host)
	}

	return nil
}

func (h *Host) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("display_name", h.DisplayName)
	enc.AddString("host_type", h.HostType.String())
	enc.AddString("root_uri", h.RootURI)
	enc.AddString("port", strconv.Itoa(h.Port))

	if h.HostType == sharedModels.HostTypes.ROMM || h.HostType == sharedModels.HostTypes.SMB {
		enc.AddString("username", h.Username)
		enc.AddString("password", h.Password)
	}

	if h.HostType == sharedModels.HostTypes.SMB {
		enc.AddString("share_name", h.ShareName)
	}

	enc.AddString("extension_filters", strings.Join(h.ExtensionFilters, ","))

	_ = enc.AddArray("sections", h.Sections)
	_ = enc.AddArray("filters", h.Filters)

	_ = enc.AddObject("table_columns", h.TableColumns)

	_ = enc.AddObject("source_replacements", h.SourceReplacements)

	_ = enc.AddObject("section_indices", h.SectionIndices)

	return nil
}
