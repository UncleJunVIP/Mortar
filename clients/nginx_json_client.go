package clients

import (
	"encoding/json"
	"fmt"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"mortar/models"
	"net/http"
	"net/url"
	"strconv"
)

type NginxJsonClient struct {
	RootURL string
}

func NewNginxJsonClient(rootURL string) *NginxJsonClient {
	return &NginxJsonClient{
		RootURL: rootURL,
	}
}

func (c *NginxJsonClient) Close() error {
	return nil
}

func (c *NginxJsonClient) ListDirectory(subdirectory string) (shared.Items, error) {
	u, err := url.Parse(c.RootURL)
	if err != nil {
		return shared.Items{}, fmt.Errorf("error parsing root url: %w", err)
	}

	u = u.JoinPath(subdirectory)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch json: %v", err)
	}
	defer resp.Body.Close()

	var nginxItems []models.NginxDirectoryListing
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&nginxItems); err != nil {
		return nil, fmt.Errorf("unable to decode nginx json: %v", err)
	}

	var items []shared.Item
	for _, nginxItem := range nginxItems {
		items = append(items, shared.Item{
			Filename:     nginxItem.Filename,
			FileSize:     strconv.FormatInt(nginxItem.Size, 10),
			LastModified: nginxItem.ModifiedTime,
		})
	}

	return items, nil
}

func (c *NginxJsonClient) BuildDownloadHeaders() map[string]string {
	headers := make(map[string]string)
	return headers
}
