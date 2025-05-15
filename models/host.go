package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
	"qlova.tech/sum"
	"strconv"
	"strings"
)

type Host struct {
	DisplayName string                   `yaml:"display_name"`
	HostType    sum.Int[shared.HostType] `yaml:"host_type"`
	RootURI     string                   `yaml:"root_uri"`
	Port        int                      `yaml:"port"`

	Username string `yaml:"username"`
	Password string `yaml:"password"`

	ExtensionFilters []string `yaml:"extension_filters"`

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

func (s SourceReplacements) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range s {
		enc.AddString(k, v)
	}

	return nil
}

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

func (h Host) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("display_name", h.DisplayName)
	enc.AddString("host_type", h.HostType.String())
	enc.AddString("root_uri", h.RootURI)
	enc.AddString("port", strconv.Itoa(h.Port))

	if h.HostType == shared.HostTypes.ROMM {
		enc.AddString("username", h.Username)
		enc.AddString("password", h.Password)
	}

	enc.AddString("extension_filters", strings.Join(h.ExtensionFilters, ","))

	_ = enc.AddArray("platforms", h.Platforms)

	_ = enc.AddObject("table_columns", h.TableColumns)

	_ = enc.AddObject("source_replacements", h.SourceReplacements)

	return nil
}

func (h Hosts) Values() []string {
	var list []string
	for _, host := range h {
		list = append(list, host.DisplayName)
	}
	return list
}

func (h Hosts) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, host := range h {
		_ = enc.AppendObject(&host)
	}

	return nil
}
