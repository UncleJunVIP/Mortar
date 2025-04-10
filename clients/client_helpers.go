package clients

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"mortar/common"
	"mortar/models"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func BuildClient(host models.Host) (models.Client, error) {
	switch host.HostType {
	case models.HostTypes.APACHE,
		models.HostTypes.MEGATHREAD,
		models.HostTypes.CUSTOM:
		return NewHttpTableClient(
			host.RootURI,
			host.HostType,
			host.TableColumns,
			host.SourceReplacements,
			host.Filters,
		), nil
	case models.HostTypes.NGINX:
		return NewNginxJsonClient(host.RootURI, host.Filters), nil
	case models.HostTypes.SMB:
		{
			return NewSMBClient(
				host.RootURI,
				host.Port,
				host.Username,
				host.Password,
				host.ShareName,
				host.ExtensionFilters,
			)
		}
	case models.HostTypes.ROMM:
		{
			return NewRomMClient(
				host.RootURI,
				host.Port,
				host.Username,
				host.Password,
			), nil
		}
	}

	return nil, nil
}

func HttpDownload(rootURL, remotePath, localPath, filename string) error {
	return HttpDownloadRename(rootURL, remotePath, localPath, filename, "")
}

func HttpDownloadRename(rootURL, remotePath, localPath, filename, rename string) error {
	logger := common.GetLoggerInstance()

	logger.Debug("Downloading file...",
		zap.String("remotePath", remotePath),
		zap.String("localPath", localPath),
		zap.String("filename", filename),
		zap.String("rename", rename))

	sourceURL, err := url.JoinPath(rootURL, remotePath, filename)
	if err != nil {
		return fmt.Errorf("unable to build download url: %w", err)
	}

	resp, err := http.Get(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	err = os.MkdirAll(localPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fn := filename

	if rename != "" {
		// Used by the thumbnail downloader when a filename doesn't have the matching tags
		imageExt := filepath.Ext(filename)
		fn = strings.TrimSuffix(rename, filepath.Ext(rename))
		fn = fn + imageExt

		appState := common.GetAppState()

		appState.LastSavedArtPath = filepath.Join(localPath, fn)

		common.UpdateAppState(appState)
	}

	f, err := os.Create(filepath.Join(localPath, fn))
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
