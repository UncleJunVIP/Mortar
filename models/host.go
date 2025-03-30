package models

import "qlova.tech/sum"

type Host struct {
	HostType           sum.Int[HostType] `yaml:"host_type"`
	RootURL            string            `yaml:"root_url"`
	Sections           []Section         `yaml:"sections"`
	Filters            []string          `yaml:"filters"`
	TableColumns       TableColumns      `yaml:"table_columns"`
	SourceReplacements map[string]string `yaml:"source_replacements"`
}

type HostType struct {
	APACHE,
	NGINX,
	CADDY,
	RAPSCALLION,
	CUSTOM sum.Int[HostType]
}

var HostTypes = sum.Int[HostType]{}.Sum()
