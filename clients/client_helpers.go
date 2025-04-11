package clients

import (
	sharedModels "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"mortar/models"
)

func BuildClient(host models.Host) (models.Client, error) {
	switch host.HostType {
	case sharedModels.HostTypes.APACHE,
		sharedModels.HostTypes.MEGATHREAD,
		sharedModels.HostTypes.CUSTOM:
		return NewHttpTableClient(
			host.RootURI,
			host.HostType,
			host.TableColumns,
			host.SourceReplacements,
			host.Filters,
		), nil
	case sharedModels.HostTypes.NGINX:
		return NewNginxJsonClient(host.RootURI, host.Filters), nil
	case sharedModels.HostTypes.SMB:
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
	case sharedModels.HostTypes.ROMM:
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
