package clients

import (
	"encoding/json"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	sharedModels "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
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

func (c *NginxJsonClient) ListDirectory(section models.MortarSection) ([]models.MortarItem, error) {
	resp, err := http.Get(c.RootURL + section.HostSubdirectory)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch json: %v", err)
	}
	defer resp.Body.Close()

	var nginxItems []models.NginxDirectoryListing
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&nginxItems); err != nil {
		return nil, fmt.Errorf("unable to decode nginx json: %v", err)
	}

	var items []models.MortarItem
	for _, nginxItem := range nginxItems {
		items = append(items, models.MortarItem{
			Item: sharedModels.Item{
				Filename: nginxItem.Filename,
			},
			FileSize: strconv.FormatInt(nginxItem.Size, 10),
			Date:     nginxItem.ModifiedTime,
		})
	}

	return items, nil
}

func (c *NginxJsonClient) DownloadFile(remotePath, localPath, filename string) error {
	return common.HttpDownload(c.RootURL, remotePath, localPath, filename)
}

func (c *NginxJsonClient) DownloadFileRename(remotePath, localPath, filename, rename string) (string, error) {
	panic("not implemented")
}
