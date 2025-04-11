package models

import (
	sharedModels "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
)

type MortarSection struct {
	sharedModels.Section `yaml:",inline"`

	RomMPlatformID string `yaml:"romm_platform_id"`
}

type MortarSections []MortarSection

func (s MortarSections) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, section := range s {
		_ = enc.AppendObject(section)
	}

	return nil
}

func (s MortarSection) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", s.Name)
	enc.AddString("system_tag", s.SystemTag)
	enc.AddString("local_directory", s.LocalDirectory)
	enc.AddString("host_subdirectory", s.HostSubdirectory)
	enc.AddString("romm_platform_id", s.RomMPlatformID)

	return nil
}
