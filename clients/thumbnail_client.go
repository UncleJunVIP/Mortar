package clients

import (
	"encoding/json"
	"log"
	"mortar/models"
	"os"
	"path/filepath"
)

const ThumbnailServerRoot = "https://thumbnails.libretro.com"
const NamedBoxArtStub = "Named_Boxarts/"

type ThumbnailClient struct {
	HttpTableClient
	SystemMapping map[string]string
}

func NewThumbnailClient() *ThumbnailClient {
	client := &ThumbnailClient{
		HttpTableClient: HttpTableClient{
			RootURL:  ThumbnailServerRoot,
			HostType: models.HostTypes.APACHE,
			TableColumns: models.TableColumns{
				FilenameHeader: "Name",
				FileSizeHeader: "",
				DateHeader:     "",
			},
		},
	}

	cwd, _ := os.Getwd()

	jsonPath := filepath.Join(cwd, "data", "systems-mapping.json")
	file, err := os.Open(jsonPath)
	if err == nil {
		defer file.Close()
		err := json.NewDecoder(file).Decode(&client.SystemMapping)

		if err != nil {
			log.Fatal(err) // TODO Remove!
		}
	}

	return client
}

func (c *ThumbnailClient) BuildThumbnailSection(tag string) models.Section {
	systemName := c.SystemMapping[tag]
	subdirectory := "/" + systemName + "/" + NamedBoxArtStub

	return models.Section{
		Name:             tag,
		HostSubdirectory: subdirectory,
		LocalDirectory:   "",
	}
}

func (c *ThumbnailClient) Close() error {
	return nil
}

func (c *ThumbnailClient) ListDirectory(section models.Section) ([]models.Item, error) {
	artList, err := c.HttpTableClient.ListDirectory(section)

	if err != nil {
		// return nil, errors.New("unable to fetch art for " + section.Name)
		log.Fatal(err) // TODO remove!
	}

	return artList, nil
}

func (c *ThumbnailClient) DownloadFile(remotePath, localPath, filename string) error {
	return HttpDownload(c.RootURL, remotePath, localPath, filename)
}
