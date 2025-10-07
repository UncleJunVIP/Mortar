package utils

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"mortar/clients"
	"mortar/models"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/disintegration/imaging"
	"qlova.tech/sum"
)

const arcadeMappingFile = "resources/arcade_mapping.txt"

var ArcadeMapping map[string]string

func init() {
	ArcadeMapping, _ = LoadArcadeMapping(arcadeMappingFile)
}

func IsDev() bool {
	return os.Getenv("ENVIRONMENT") == "DEV"
}

func GetRomDirectory() string {
	if IsDev() {
		return os.Getenv("ROM_DIRECTORY")
	}

	return common.RomDirectory
}

func UnzipGame(platform models.Platform, game shared.Item) ([]string, error) {
	logger := gaba.GetLoggerInstance()

	zipPath := filepath.Join(platform.LocalDirectory, game.Filename)
	romDirectory := platform.LocalDirectory

	if IsDev() {
		romDirectory = strings.ReplaceAll(platform.LocalDirectory, common.RomDirectory, GetRomDirectory())
		zipPath = filepath.Join(romDirectory, game.Filename)
	}

	extractedFiles, err := gaba.ProcessMessage(fmt.Sprintf("%s %s...", "Unzipping", game.DisplayName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		extractedFiles, err := Unzip(zipPath, romDirectory)
		if err != nil {
			return nil, err
		}

		logger.Debug("Extracted files", "files", extractedFiles)

		return extractedFiles, nil
	})

	if err != nil {
		gaba.ProcessMessage(fmt.Sprintf("Unable to unzip %s", game.DisplayName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(3 * time.Second)
			return nil, nil
		})
		logger.Error("Unable to unzip pak", "error", err)
		return nil, err
	} else {
		deleted := common.DeleteFile(zipPath)
		if !deleted {
			return nil, errors.New("unable to delete zip file")
		}
	}

	return extractedFiles.Result.([]string), nil
}

func Unzip(src, dest string) ([]string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return nil, err
	}

	extractedFiles := []string{}

	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(path, f.Mode())
			if err != nil {
				return err
			}
		} else {
			err := os.MkdirAll(filepath.Dir(path), f.Mode())
			if err != nil {
				return err
			}

			tempPath := path + ".tmp"
			tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			_, err = io.Copy(tempFile, rc)
			tempFile.Close() // Close the file before attempting to rename it

			if err != nil {
				common.DeleteFile(tempPath)
				return err
			}

			// Now rename the temporary file to the target path
			err = os.Rename(tempPath, path)
			if err != nil {
				common.DeleteFile(tempPath)
				return err
			}

			extractedFiles = append(extractedFiles, path)
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return extractedFiles, err
		}
	}

	return extractedFiles, nil
}

func ListZipContents(platform models.Platform, game shared.Item) ([]string, error) {
	zipPath := filepath.Join(platform.LocalDirectory, game.Filename)
	romDirectory := platform.LocalDirectory

	if IsDev() {
		romDirectory = strings.ReplaceAll(platform.LocalDirectory, common.RomDirectory, GetRomDirectory())
		zipPath = filepath.Join(romDirectory, game.Filename)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	filenames := make([]string, 0, len(reader.File))

	for _, file := range reader.File {
		filenames = append(filenames, file.Name)
	}

	return filenames, nil
}

func HasBinCue(platform models.Platform, game shared.Item) bool {
	filenames, err := ListZipContents(platform, game)
	if err != nil {
		return false
	}
	for _, filename := range filenames {
		if strings.HasSuffix(filename, ".bin") || strings.HasSuffix(filename, ".cue") {
			return true
		}
	}

	return false
}

func IsMultiDisc(platform models.Platform, game shared.Item) bool {
	if filepath.Ext(game.Filename) == ".zip" {
		filenames, err := ListZipContents(platform, game)
		if err != nil {
			return false
		}

		for _, filename := range filenames {
			if strings.Contains(filename, "(Disc") || strings.Contains(filename, "(Disk") {
				return true
			}
		}

		return false
	}

	return strings.Contains(game.Filename, "(Disc") || strings.Contains(game.Filename, "(Disk")
}

func GroupBinCue(platform models.Platform, game shared.Item) {
	logger := gaba.GetLoggerInstance()

	unzipped, err := UnzipGame(platform, game)

	if err == nil && len(unzipped) > 0 {

		gaba.ProcessMessage(fmt.Sprintf("Grouping BIN/CUE for %s", game.DisplayName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(1500 * time.Millisecond)
			logger.Debug("Grouping BIN / CUE ROMs")

			// Find all CUE files in the unzipped files
			cueFiles := []string{}
			for _, file := range unzipped {
				if strings.HasSuffix(file, ".cue") {
					cueFiles = append(cueFiles, file)
				}
			}

			// For each CUE file, create a directory and move related files
			for _, cueFile := range cueFiles {
				baseName := filepath.Base(cueFile)
				dirName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
				dirPath := filepath.Join(filepath.Dir(cueFile), dirName)

				// Create directory with the same name as the CUE file
				err := os.MkdirAll(dirPath, 0755)
				if err != nil {
					logger.Error("Failed to create directory for BIN/CUE grouping",
						"directory", dirPath,
						"error", err)
					continue
				}

				// Move all related files (both BIN and CUE) to the new directory
				for _, file := range unzipped {
					// Check if file is in the same directory as the CUE file
					if filepath.Dir(file) == filepath.Dir(cueFile) {
						fileBaseName := filepath.Base(file)
						// Check if it's a BIN file or the CUE file itself
						if strings.HasSuffix(file, ".bin") || file == cueFile {
							newPath := filepath.Join(dirPath, fileBaseName)
							err := os.Rename(file, newPath)
							if err != nil {
								logger.Error("Failed to move file to BIN/CUE group directory",
									"file", file,
									"destination", newPath,
									"error", err)
							}
						}
					}
				}

				logger.Debug("Successfully grouped BIN/CUE files",
					"cueFile", baseName,
					"directory", dirPath)
			}

			return nil, nil
		})
	}
}

func GroupMultiDisk(platform models.Platform, game shared.Item) error {
	logger := gaba.GetLoggerInstance()

	gameFolderName := game.DisplayName
	diskIndex := strings.Index(gameFolderName, "(Disk")
	discIndex := strings.Index(gameFolderName, "(Disc")

	trimIndex := -1
	if diskIndex != -1 && discIndex != -1 {
		trimIndex = min(diskIndex, discIndex)
	} else if diskIndex != -1 {
		trimIndex = diskIndex
	} else if discIndex != -1 {
		trimIndex = discIndex
	}

	if trimIndex != -1 {
		gameFolderName = gameFolderName[:trimIndex]
	}

	gameFolderName = strings.TrimSpace(gameFolderName)
	gameFolderPath := filepath.Join(platform.LocalDirectory, gameFolderName)

	if IsDev() {
		romDirectory := strings.ReplaceAll(platform.LocalDirectory, common.RomDirectory, GetRomDirectory())
		gameFolderPath = filepath.Join(romDirectory, gameFolderName)
	}

	if _, err := os.Stat(gameFolderPath); os.IsNotExist(err) {
		err := os.MkdirAll(gameFolderPath, 0755)
		if err != nil {
			logger.Error("Failed to create game directory", "error", err)
			return err
		}
		logger.Debug("Created new game directory", "path", gameFolderPath)
	} else {
		logger.Debug("Game directory already exists, skipping creation", "path", gameFolderPath)
	}

	var extractedFiles []string

	if filepath.Ext(game.Filename) == ".zip" {
		var err error
		extractedFiles, err = UnzipGame(platform, game)
		if err != nil {
			logger.Error("Failed to unzip game", "error", err)
			return err
		}
	} else {
		romDirectory := platform.LocalDirectory

		if IsDev() {
			romDirectory = strings.ReplaceAll(platform.LocalDirectory, common.RomDirectory, GetRomDirectory())
		}

		extractedFiles = append(extractedFiles, filepath.Join(romDirectory, game.Filename))
	}

	_, err := gaba.ProcessMessage(fmt.Sprintf("Wrangling multi-disk game %s", game.DisplayName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		time.Sleep(1500 * time.Millisecond)
		for _, filePath := range extractedFiles {
			fileName := filepath.Base(filePath)

			destPath := filepath.Join(gameFolderPath, fileName)

			err := os.Rename(filePath, destPath)
			if err != nil {
				logger.Error("Failed to move file", "source", filePath, "destination", destPath, "error", err)
				return nil, err
			}
		}

		// Create or append to M3U file with the game's display name
		m3uFileName := fmt.Sprintf("%s.m3u", gameFolderName)
		m3uFilePath := filepath.Join(gameFolderPath, m3uFileName)

		// Find all .cue, .chd, and .pbp files in the new directory and add them to the M3U
		var discFiles []string
		for _, filePath := range extractedFiles {
			fileName := filepath.Base(filePath)
			fileNameLower := strings.ToLower(fileName)
			if strings.HasSuffix(fileNameLower, ".cue") ||
				strings.HasSuffix(fileNameLower, ".chd") ||
				strings.HasSuffix(fileNameLower, ".pbp") {
				discFiles = append(discFiles, fileName)
			}
		}

		// Check if there are any disc files to add
		if len(discFiles) > 0 {
			// Open the M3U file for appending (or create if it doesn't exist)
			m3uFile, err := os.OpenFile(m3uFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				logger.Error("Failed to open M3U file", "error", err)
				return nil, err
			}
			defer m3uFile.Close()

			// Write each disc file to the M3U, one per line
			for _, discFile := range discFiles {
				_, err := m3uFile.WriteString(discFile + "\n")
				if err != nil {
					logger.Error("Failed to write to M3U file", "error", err)
					return nil, err
				}
			}

			logger.Debug("Successfully appended to M3U file",
				"m3u_path", m3uFilePath,
				"disc_files", discFiles)
		} else {
			logger.Debug("No .cue, .chd, or .pbp files found to add to M3U file")
		}

		logger.Debug("Successfully processed game",
			"folder", gameFolderPath,
			"m3u_path", m3uFilePath)

		return nil, nil
	})

	return err
}

func FindArt(platform models.Platform, game shared.Item, downloadType sum.Int[shared.ArtDownloadType]) string {
	logger := gaba.GetLoggerInstance()

	artDirectory := ""

	if IsDev() {
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

	client := common.NewThumbnailClient(downloadType)
	section := client.BuildThumbnailSection(platform.SystemTag)

	artList, err := client.ListDirectory(section.HostSubdirectory)

	if err != nil {
		logger.Debug("Unable to fetch artlist", "error", err)
		return ""
	}

	// toastd's trick for Libretro Thumbnail Naming
	cleanedName := strings.ReplaceAll(game.DisplayName, "&", "_")

	var matched shared.Item

	// naive search first
	for _, art := range artList {
		if strings.Contains(strings.ToLower(art.Filename), strings.ToLower(cleanedName)) {
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
			logger.Error("Unable to open last saved art", "error", err)
			return ""
		}

		dst := imaging.Resize(src, 500, 0, imaging.Lanczos)

		err = imaging.Save(dst, lastSavedArtPath)
		if err != nil {
			logger.Error("Unable to save resized last saved art", "error", err)
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
	logger := gaba.GetLoggerInstance()
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(cwd, ".cache"))
	if err != nil {
		logger.Error("Unable to delete cache", "error", err)
		return err
	}

	logger.Debug("Cache deleted")
	return nil
}

func LoadArcadeMapping(filePath string) (map[string]string, error) {
	mapping := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			romFile := strings.TrimSpace(parts[0])
			displayName := strings.TrimSpace(parts[1])
			mapping[romFile] = displayName
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return mapping, nil
}

func IsConnectedToInternet() bool {
	timeout := 5 * time.Second
	_, err := net.DialTimeout("tcp", "8.8.8.8:53", timeout)
	return err == nil
}

func AllPlatformsHaveLocalFolders(config *models.Config) []string {
	var missingPlatforms []string

	for _, h := range config.Hosts {
		for _, p := range h.Platforms {
			if p.LocalDirectory == "" {
				missingPlatforms = append(missingPlatforms, p.Name)
			}
		}
	}

	return missingPlatforms
}
