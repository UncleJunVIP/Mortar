package clients

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func HttpDownload(rootURL, remotePath, localPath, filename string) error {
	sourceURL := rootURL + remotePath + filename
	resp, err := http.Get(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	err = os.MkdirAll(localPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(filepath.Join(localPath, filename))
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
