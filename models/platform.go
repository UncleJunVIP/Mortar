package models

type Platform struct {
	Name             string `yaml:"platform_name"`
	SystemTag        string `yaml:"system_tag"`
	LocalDirectory   string `yaml:"local_directory"`
	HostSubdirectory string `yaml:"host_subdirectory"`
	RomMPlatformID   string `yaml:"romm_platform_id"`

	SkipExclusiveFilters bool `yaml:"skip_exclusive_filters"`
	SkipInclusiveFilters bool `yaml:"skip_inclusive_filters"`
	IsArcade             bool `yaml:"is_arcade"`

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
