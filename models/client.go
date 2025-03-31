package models

type Client interface {
	Close() error
	ListDirectory(path string) ([]Item, error)
	DownloadFile(remotePath, localPath, filename string) error
}
