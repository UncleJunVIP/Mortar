package clients

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"mortar/models"
)

func BuildClient(host models.Host) (shared.Client, error) {
	switch host.HostType {
	case shared.HostTypes.APACHE,
		shared.HostTypes.MEGATHREAD,
		shared.HostTypes.CUSTOM:
		return NewHttpTableClient(
			host.RootURI,
			host.HostType,
			host.TableColumns,
			host.SourceReplacements,
		), nil
	case shared.HostTypes.NGINX:
		return NewNginxJsonClient(host.RootURI), nil
	case shared.HostTypes.SMB:
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
	case shared.HostTypes.ROMM:
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
