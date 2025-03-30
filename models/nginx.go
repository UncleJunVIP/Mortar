package models

type NginxDirectoryListing struct {
	Filename     string `json:"name"`
	Type         string `json:"type"`
	ModifiedTime string `json:"mtime"`
	Size         int64  `json:"size"`
}
