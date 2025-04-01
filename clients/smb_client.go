package clients

import (
	"fmt"
	"github.com/hirochachacha/go-smb2"
	"mortar/models"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type SMBClient struct {
	Hostname         string
	Port             int
	Username         string
	Password         string
	ShareName        string
	ExtensionFilters []string

	Connection net.Conn
	Session    *smb2.Session
	Mount      *smb2.Share
}

func NewSMBClient(hostname string, port int, username, password, shareName string, extensionFilters []string) (*SMBClient, error) {
	c := &SMBClient{
		Hostname:         hostname,
		Port:             port,
		Username:         username,
		Password:         password,
		ShareName:        shareName,
		ExtensionFilters: extensionFilters,
	}

	var err error

	address := fmt.Sprintf("%s:%d", hostname, port)

	c.Connection, err = net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User: "GUEST",
		},
	}

	c.Session, err = d.Dial(c.Connection)
	if err != nil {
		return nil, err
	}

	c.Mount, err = c.Session.Mount(c.ShareName)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *SMBClient) Close() error {
	err := c.Mount.Umount()
	if err != nil {
		return err
	}

	err = c.Session.Logoff()
	if err != nil {
		return err
	}

	err = c.Connection.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *SMBClient) ListDirectory(section models.Section) ([]models.Item, error) {
	ls, err := c.Mount.ReadDir(section.HostSubdirectory)
	if err != nil {
		return nil, err
	}

	filenames := make([]models.Item, 0, len(ls))

	if len(c.ExtensionFilters) > 0 {
		for _, l := range ls {
			for _, f := range c.ExtensionFilters {
				if !strings.Contains(l.Name(), f) {
					filenames = append(filenames, models.Item{Filename: l.Name()})
				}
			}
		}
	} else {
		for _, l := range ls {
			filenames = append(filenames, models.Item{Filename: l.Name()})
		}
	}

	return filenames, nil
}

func (c *SMBClient) DownloadFile(remotePath, localPath, filename string) error {
	bytes, err := c.Mount.ReadFile(filepath.Join(remotePath, filename))
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(localPath, filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
