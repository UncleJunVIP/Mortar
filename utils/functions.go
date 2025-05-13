package utils

import (
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/disintegration/imaging"
	"go.uber.org/zap"
	"mortar/clients"
	"mortar/models"
	"os"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
)

func GetRomDirectory() string {
	if os.Getenv("DEVELOPMENT") == "true" {
		return "/Users/btk/Desktop/Roms"
	}

	return common.RomDirectory
}

func FindArt(platform models.Platform, game shared.Item, downloadType sum.Int[shared.ArtDownloadType]) string {
	logger := common.GetLoggerInstance()

	artDirectory := ""

	if os.Getenv("DEVELOPMENT") == "true" {
		romDirectory := strings.ReplaceAll(platform.LocalDirectory, common.RomDirectory, GetRomDirectory())
		artDirectory = filepath.Join(romDirectory, ".media")
	} else {
		artDirectory = filepath.Join(platform.LocalDirectory, ".media")
	}

	host := platform.Host

	if host.HostType == shared.HostTypes.ROMM {
		// Skip all this silliness and grab the art from RoMM
		client, err := clients.BuildClient(host)
		if err != nil {
			return ""
		}

		rommClient := client.(*clients.RomMClient)

		if game.ArtURL == "" {
			return ""
		}

		slashIdx := strings.LastIndex(game.ArtURL, "/")
		artSubdirectory, artFilename := game.ArtURL[:slashIdx], game.ArtURL[slashIdx+1:]

		artFilename = strings.Split(artFilename, "?")[0] // For the query string caching stuff

		LastSavedArtPath, err := rommClient.DownloadArt(artSubdirectory,
			artDirectory, artFilename, game.Filename)

		if err != nil {
			return ""
		}

		return LastSavedArtPath
	}

	tag := common.TagRegex.FindStringSubmatch(platform.LocalDirectory)

	if tag == nil {
		return ""
	}

	client := common.NewThumbnailClient(downloadType)
	section := client.BuildThumbnailSection(tag[1])

	artList, err := client.ListDirectory(section.HostSubdirectory)

	if err != nil {
		logger.Info("Unable to fetch artlist", zap.Error(err))
		return ""
	}

	noExtension := strings.TrimSuffix(game.Filename, filepath.Ext(game.Filename))

	var matched shared.Item

	// naive search first
	for _, art := range artList {
		if strings.Contains(strings.ToLower(art.Filename), strings.ToLower(noExtension)) {
			matched = art
			break
		}
	}

	if matched.Filename != "" {
		lastSavedArtPath, err := client.DownloadArt(section.HostSubdirectory, artDirectory, matched.Filename, game.Filename)
		if err != nil {
			return ""
		}

		src, err := imaging.Open(lastSavedArtPath)
		if err != nil {
			logger.Error("Unable to open last saved art", zap.Error(err))
			return ""
		}

		dst := imaging.Resize(src, 500, 0, imaging.Lanczos)

		err = imaging.Save(dst, lastSavedArtPath)
		if err != nil {
			logger.Error("Unable to save resized last saved art", zap.Error(err))
			return ""
		}

		return lastSavedArtPath
	}

	return ""
}

func MapTagsToDirectories(items shared.Items) map[string]string {
	mapping := make(map[string]string)

	for _, entry := range items {
		if entry.IsDirectory {
			tag := strings.ReplaceAll(entry.Tag, "(", "")
			tag = strings.ReplaceAll(tag, ")", "")
			path := filepath.Join(common.RomDirectory, entry.Filename)
			mapping[tag] = path
		}
	}

	return mapping
}

func CachedMegaThreadJsonFilename(hostName, platformName string) string {
	return strings.ReplaceAll(fmt.Sprintf("%s_%s_%s.json",
		hostName, platformName, "megathread"), " ", "")
}

func CacheFolderExists() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	cachePath := filepath.Join(cwd, ".cache")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func DeleteCache() error {
	logger := common.GetLoggerInstance()
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(cwd, ".cache"))
	if err != nil {
		logger.Error("Unable to delete cache", zap.Error(err))
		return err
	}

	logger.Info("Cache deleted")
	return nil
}
