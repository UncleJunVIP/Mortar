package clients

import (
	"encoding/json"
	"fmt"
	converthtmltabletodata "github.com/activcoding/HTML-Table-to-JSON"
	"mortar/models"
	"net/http"
	"net/url"
	"qlova.tech/sum"
	"strings"
)

type HostType = sum.Int[models.HostType]

type HttpTableClient struct {
	RootURL            string
	HostType           HostType
	TableColumns       models.TableColumns
	SourceReplacements map[string]string
	Filters            []string
}

func NewHttpTableClient(rootURL string, hostType HostType, tableColumns models.TableColumns,
	sourceReplacements map[string]string, filters []string) *HttpTableClient {
	return &HttpTableClient{
		RootURL:            rootURL,
		HostType:           hostType,
		TableColumns:       tableColumns,
		SourceReplacements: sourceReplacements,
		Filters:            filters,
	}
}

func (c *HttpTableClient) Close() error {
	return nil
}

func (c *HttpTableClient) ListDirectory(section models.Section) ([]models.Item, error) {
	params := url.Values{}

	switch c.HostType {
	case models.HostTypes.APACHE:
		params.Add("F", "2") // To enable table mode for mod_autoindex
	}

	combinedUrl := c.RootURL + section.HostSubdirectory
	u, err := url.Parse(combinedUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse table URL: %v", err)
	}
	u.RawQuery = params.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch table:, %v", err)
	}
	defer resp.Body.Close()

	jsonBytes, err := converthtmltabletodata.ConvertReaderToJSON(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse table into json: %v", err)
	}

	rawJson := string(jsonBytes)

	cleaned := rawJson

	switch c.HostType {
	case models.HostTypes.APACHE:
		cleaned = strings.ReplaceAll(cleaned, "[[", "[")
		cleaned = strings.ReplaceAll(cleaned, "]]", "]")
		cleaned = strings.ReplaceAll(cleaned, "Name", "filename")
		cleaned = strings.ReplaceAll(cleaned, "Size", "file_size")
		cleaned = strings.ReplaceAll(cleaned, "Last modified", "date")
	case models.HostTypes.MEGATHREAD:
		{
			cleaned = strings.ReplaceAll(cleaned, "  ↓", "")
			cleaned = strings.ReplaceAll(cleaned, "[[", "[")
			cleaned = strings.ReplaceAll(cleaned, "]]", "]")
			cleaned = strings.ReplaceAll(cleaned, "File Name", "filename")
			cleaned = strings.ReplaceAll(cleaned, "File Size", "file_size")
			cleaned = strings.ReplaceAll(cleaned, "Date", "date")
		}
	case models.HostTypes.CUSTOM:
		{
			for oldValue, newValue := range c.SourceReplacements {
				cleaned = strings.ReplaceAll(cleaned, oldValue, newValue)
			}

			cleaned = strings.ReplaceAll(cleaned, c.TableColumns.FilenameHeader, "filename")
			cleaned = strings.ReplaceAll(cleaned, c.TableColumns.FileSizeHeader, "file_size")
			cleaned = strings.ReplaceAll(cleaned, c.TableColumns.DateHeader, "date")
		}

	}

	var items []models.Item
	err = json.Unmarshal([]byte(cleaned), &items)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %v", err)
	}

	// Skip the header row(s)
	switch c.HostType {
	case models.HostTypes.APACHE,
		models.HostTypes.MEGATHREAD:
		{
			if len(items) > 1 {
				return items[1:], nil
			}
		}
	}

	return nil, nil
}

func (c *HttpTableClient) DownloadFile(remotePath, localPath, filename string) error {
	return HttpDownload(c.RootURL, remotePath, localPath, filename)
}

func (c *HttpTableClient) DownloadFileRename(remotePath, localPath, filename, rename string) error {
	panic("not implemented")
}
