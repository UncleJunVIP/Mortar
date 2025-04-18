package models

import "go.uber.org/zap/zapcore"

type Platform struct {
	Name             string `yaml:"platform_name"`
	SystemTag        string `yaml:"system_tag"`
	LocalDirectory   string `yaml:"local_directory"`
	HostSubdirectory string `yaml:"host_subdirectory"`
	RomMPlatformID   string `yaml:"romm_platform_id"`

	Host Host `yaml:"-"`
}

type Platforms []Platform

func (p Platform) Value() interface{} {
	return p
}

func (p Platforms) Values() []string {
	var list []string
	for _, platform := range p {
		list = append(list, platform.Name)
	}
	return list
}

func (p Platforms) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, section := range p {
		_ = encoder.AppendObject(section)
	}

	return nil
}

func (p Platform) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("name", p.Name)
	encoder.AddString("system_tag", p.SystemTag)
	encoder.AddString("local_directory", p.LocalDirectory)
	encoder.AddString("host_subdirectory", p.HostSubdirectory)
	encoder.AddString("romm_platform_id", p.RomMPlatformID)

	return nil
}
