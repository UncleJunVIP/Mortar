package models

type Client interface {
	Close() error
	ListDirectory(section Section) ([]Item, error)
	DownloadFile(remotePath, localPath, filename string) error
	DownloadFileRename(remotePath, localPath, filename, rename string) error
}
