package clients

import (
	"encoding/json"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	htmltableparser "github.com/activcoding/HTML-Table-to-JSON"
	"net/http"
	"net/url"
	"qlova.tech/sum"
	"strings"
)

type HttpTableClient struct {
	RootURL            string
	HostType           sum.Int[shared.HostType]
	TableColumns       shared.TableColumns
	SourceReplacements map[string]string
}

func NewHttpTableClient(rootURL string, hostType sum.Int[shared.HostType], tableColumns shared.TableColumns,
	sourceReplacements map[string]string) *HttpTableClient {
	return &HttpTableClient{
		RootURL:            rootURL,
		HostType:           hostType,
		TableColumns:       tableColumns,
		SourceReplacements: sourceReplacements,
	}
}

func (c *HttpTableClient) Close() error {
	return nil
}

func (c *HttpTableClient) ListDirectory(subdirectory string) (shared.Items, error) {
	params := url.Values{}

	switch c.HostType {
	case shared.HostTypes.APACHE:
		params.Add("F", "2") // To enable table mode for mod_autoindex
	}

	base, err := url.Parse(c.RootURL)
	if err != nil {
		return nil, fmt.Errorf("invalid root URL: %v", err)
	}

	ref, err := url.Parse(subdirectory)
	if err != nil {
		return nil, fmt.Errorf("invalid subdirectory: %v", err)
	}

	u := base.ResolveReference(ref)
	u.RawQuery = params.Encode()

	u.RawQuery = params.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch table:, %v", err)
	}
	defer resp.Body.Close()

	jsonBytes, err := htmltableparser.ConvertReaderToJSON(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse table into json: %v", err)
	}

	rawJson := string(jsonBytes)

	cleaned := rawJson

	switch c.HostType {
	case shared.HostTypes.APACHE:
		cleaned = strings.ReplaceAll(cleaned, "[[", "[")
		cleaned = strings.ReplaceAll(cleaned, "]]", "]")
		cleaned = strings.ReplaceAll(cleaned, "Name", "filename")
		cleaned = strings.ReplaceAll(cleaned, "Size", "file_size")
		cleaned = strings.ReplaceAll(cleaned, "Last modified", "date")
	case shared.HostTypes.MEGATHREAD:
		{
			cleaned = strings.ReplaceAll(cleaned, "  ↓", "")
			cleaned = strings.ReplaceAll(cleaned, "[[", "[")
			cleaned = strings.ReplaceAll(cleaned, "]]", "]")
			cleaned = strings.ReplaceAll(cleaned, "File Name", "filename")
			cleaned = strings.ReplaceAll(cleaned, "File Size", "file_size")
			cleaned = strings.ReplaceAll(cleaned, "Date", "date")
		}
	case shared.HostTypes.CUSTOM:
		{
			for oldValue, newValue := range c.SourceReplacements {
				cleaned = strings.ReplaceAll(cleaned, oldValue, newValue)
			}

			cleaned = strings.ReplaceAll(cleaned, c.TableColumns.FilenameHeader, "filename")
			cleaned = strings.ReplaceAll(cleaned, c.TableColumns.FileSizeHeader, "file_size")
			cleaned = strings.ReplaceAll(cleaned, c.TableColumns.DateHeader, "date")
		}

	}

	var items []shared.Item
	err = json.Unmarshal([]byte(cleaned), &items)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %v", err)
	}

	// Skip the header row(s)
	switch c.HostType {
	case shared.HostTypes.APACHE,
		shared.HostTypes.MEGATHREAD:
		{
			if len(items) > 1 {
				return items[1:], nil
			}
		}
	}

	return nil, nil
}

func (c *HttpTableClient) DownloadFile(remotePath, localPath, filename string) (string, error) {
	return common.HttpDownload(c.RootURL, remotePath, localPath, filename)
}

func (c *HttpTableClient) DownloadFileRename(remotePath, localPath, filename, rename string) (string, error) {
	panic("not implemented")
}
