package models

type Client interface {
	Close() error
	ListDirectory(section MortarSection) ([]MortarItem, error)
	DownloadFile(remotePath, localPath, filename string) error
	DownloadFileRename(remotePath, localPath, filename, rename string) (string, error)
}
