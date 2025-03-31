package clients

import (
	"encoding/json"
	"fmt"
	"mortar/models"
	"net/http"
	"strconv"
)

type NginxJsonClient struct {
	RootURL string
	Filters []string
}

func NewNginxJsonClient(rootURL string, filters []string) *NginxJsonClient {
	return &NginxJsonClient{
		RootURL: rootURL,
		Filters: filters,
	}
}

func (c *NginxJsonClient) Close() error {
	return nil
}

func (c *NginxJsonClient) ListDirectory(path string) ([]models.Item, error) {
	resp, err := http.Get(c.RootURL + path)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch json: %v", err)
	}
	defer resp.Body.Close()

	var nginxItems []models.NginxDirectoryListing
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&nginxItems); err != nil {
		return nil, fmt.Errorf("unable to decode nginx json: %v", err)
	}

	var items []models.Item
	for _, nginxItem := range nginxItems {
		items = append(items, models.Item{
			Filename: nginxItem.Filename,
			FileSize: strconv.FormatInt(nginxItem.Size, 10),
			Date:     nginxItem.ModifiedTime,
		})
	}

	return items, nil
}

func (c *NginxJsonClient) DownloadFile(remotePath, localPath, filename string) error {
	return HttpDownload(c.RootURL, remotePath, localPath, filename)
}
