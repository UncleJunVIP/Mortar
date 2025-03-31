package clients

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func HttpDownload(rootURL, remotePath, localPath, filename string) error {
	sourceURL := rootURL + remotePath
	resp, err := http.Get(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(filepath.Join(localPath, filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
