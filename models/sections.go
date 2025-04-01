package models

type Section struct {
	Name             string `yaml:"section_name"`
	HostSubdirectory string `yaml:"host_subdirectory"`
	LocalDirectory   string `yaml:"local_directory"`

	RomMPlatformID string `yaml:"romm_platform_id"`
}
type TableColumns struct {
	FilenameHeader string `yaml:"filename_header"`
	FileSizeHeader string `yaml:"file_size_header"`
	DateHeader     string `yaml:"date_header"`
}
