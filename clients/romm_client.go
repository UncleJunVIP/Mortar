package clients

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mortar/models"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type RomMClient struct {
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type RomMPlatform struct {
	ID          int           `json:"id"`
	Slug        string        `json:"slug"`
	FsSlug      string        `json:"fs_slug"`
	RomCount    int           `json:"rom_count"`
	Name        string        `json:"name"`
	CustomName  string        `json:"custom_name"`
	IgdbID      int           `json:"igdb_id"`
	SgdbID      interface{}   `json:"sgdb_id"`
	MobyID      int           `json:"moby_id"`
	SsID        int           `json:"ss_id"`
	Category    string        `json:"category"`
	Generation  int           `json:"generation"`
	FamilyName  string        `json:"family_name"`
	FamilySlug  string        `json:"family_slug"`
	URL         string        `json:"url"`
	URLLogo     string        `json:"url_logo"`
	LogoPath    string        `json:"logo_path"`
	Firmware    []interface{} `json:"firmware"`
	AspectRatio string        `json:"aspect_ratio"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	DisplayName string        `json:"display_name"`
}

type RomMRom struct {
	ID                  int         `json:"id"`
	IgdbID              int         `json:"igdb_id"`
	SgdbID              interface{} `json:"sgdb_id"`
	MobyID              interface{} `json:"moby_id"`
	SsID                interface{} `json:"ss_id"`
	PlatformID          int         `json:"platform_id"`
	PlatformSlug        string      `json:"platform_slug"`
	PlatformFsSlug      string      `json:"platform_fs_slug"`
	PlatformName        string      `json:"platform_name"`
	PlatformCustomName  string      `json:"platform_custom_name"`
	PlatformDisplayName string      `json:"platform_display_name"`
	FsName              string      `json:"fs_name"`
	FsNameNoTags        string      `json:"fs_name_no_tags"`
	FsNameNoExt         string      `json:"fs_name_no_ext"`
	FsExtension         string      `json:"fs_extension"`
	FsPath              string      `json:"fs_path"`
	FsSizeBytes         int         `json:"fs_size_bytes"`
	Name                string      `json:"name"`
	Slug                string      `json:"slug"`
	Summary             string      `json:"summary"`
	FirstReleaseDate    int64       `json:"first_release_date"`
	YoutubeVideoID      string      `json:"youtube_video_id"`
	AverageRating       float64     `json:"average_rating"`
	AlternativeNames    []string    `json:"alternative_names"`
	Genres              []string    `json:"genres"`
	Franchises          []string    `json:"franchises"`
	MetaCollections     []string    `json:"meta_collections"`
	Companies           []string    `json:"companies"`
	GameModes           []string    `json:"game_modes"`
	AgeRatings          []string    `json:"age_ratings"`
	IgdbMetadata        struct {
		TotalRating      string   `json:"total_rating"`
		AggregatedRating string   `json:"aggregated_rating"`
		FirstReleaseDate int      `json:"first_release_date"`
		YoutubeVideoID   string   `json:"youtube_video_id"`
		Genres           []string `json:"genres"`
		Franchises       []string `json:"franchises"`
		AlternativeNames []string `json:"alternative_names"`
		Collections      []string `json:"collections"`
		Companies        []string `json:"companies"`
		GameModes        []string `json:"game_modes"`
		AgeRatings       []struct {
			Rating         string `json:"rating"`
			Category       string `json:"category"`
			RatingCoverURL string `json:"rating_cover_url"`
		} `json:"age_ratings"`
		Platforms []struct {
			IgdbID int    `json:"igdb_id"`
			Name   string `json:"name"`
		} `json:"platforms"`
		Expansions []interface{} `json:"expansions"`
		Dlcs       []interface{} `json:"dlcs"`
		Remasters  []interface{} `json:"remasters"`
		Remakes    []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Slug     string `json:"slug"`
			Type     string `json:"type"`
			CoverURL string `json:"cover_url"`
		} `json:"remakes"`
		ExpandedGames []interface{} `json:"expanded_games"`
		Ports         []interface{} `json:"ports"`
		SimilarGames  []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Slug     string `json:"slug"`
			Type     string `json:"type"`
			CoverURL string `json:"cover_url"`
		} `json:"similar_games"`
	} `json:"igdb_metadata"`
	MobyMetadata struct {
	} `json:"moby_metadata"`
	SsMetadata     interface{}   `json:"ss_metadata"`
	PathCoverSmall string        `json:"path_cover_small"`
	PathCoverLarge string        `json:"path_cover_large"`
	URLCover       string        `json:"url_cover"`
	HasManual      bool          `json:"has_manual"`
	PathManual     interface{}   `json:"path_manual"`
	URLManual      interface{}   `json:"url_manual"`
	IsUnidentified bool          `json:"is_unidentified"`
	Revision       string        `json:"revision"`
	Regions        []interface{} `json:"regions"`
	Languages      []interface{} `json:"languages"`
	Tags           []interface{} `json:"tags"`
	CrcHash        string        `json:"crc_hash"`
	Md5Hash        string        `json:"md5_hash"`
	Sha1Hash       string        `json:"sha1_hash"`
	Multi          bool          `json:"multi"`
	Files          []struct {
		ID            int         `json:"id"`
		RomID         int         `json:"rom_id"`
		FileName      string      `json:"file_name"`
		FilePath      string      `json:"file_path"`
		FileSizeBytes int         `json:"file_size_bytes"`
		FullPath      string      `json:"full_path"`
		CreatedAt     time.Time   `json:"created_at"`
		UpdatedAt     time.Time   `json:"updated_at"`
		LastModified  time.Time   `json:"last_modified"`
		CrcHash       string      `json:"crc_hash"`
		Md5Hash       string      `json:"md5_hash"`
		Sha1Hash      string      `json:"sha1_hash"`
		Category      interface{} `json:"category"`
	} `json:"files"`
	FullPath    string        `json:"full_path"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	SiblingRoms []interface{} `json:"sibling_roms"`
	RomUser     struct {
		ID              int         `json:"id"`
		UserID          int         `json:"user_id"`
		RomID           int         `json:"rom_id"`
		CreatedAt       time.Time   `json:"created_at"`
		UpdatedAt       time.Time   `json:"updated_at"`
		LastPlayed      interface{} `json:"last_played"`
		NoteRawMarkdown string      `json:"note_raw_markdown"`
		NoteIsPublic    bool        `json:"note_is_public"`
		IsMainSibling   bool        `json:"is_main_sibling"`
		Backlogged      bool        `json:"backlogged"`
		NowPlaying      bool        `json:"now_playing"`
		Hidden          bool        `json:"hidden"`
		Rating          int         `json:"rating"`
		Difficulty      int         `json:"difficulty"`
		Completion      int         `json:"completion"`
		Status          interface{} `json:"status"`
		UserUsername    string      `json:"user__username"`
	} `json:"rom_user"`
	SortComparator string `json:"sort_comparator"`
}

const PlatformEndpoint = "/api/platforms/"
const RomsEndpoint = "/api/roms/"

func NewRomMClient(hostname string, port int, username string, password string) *RomMClient {
	return &RomMClient{
		Hostname: hostname,
		Port:     port,
		Username: username,
		Password: password,
	}
}

func (c *RomMClient) Close() error {
	return nil
}

func (c *RomMClient) buildRootURL() string {
	if c.Port != 0 {
		return c.Hostname + ":" + strconv.Itoa(c.Port)
	}

	return c.Hostname
}

func (c *RomMClient) ListDirectory(section models.Section) ([]models.Item, error) {
	auth := c.Username + ":" + c.Password
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	u, err := url.Parse(c.buildRootURL() + RomsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse rom endpoint URL for listing: %v", err)
	}

	params := url.Values{}
	params.Add("platform_id", section.RomMPlatformID)

	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to build rom list request: %v", err)
	}

	req.Header.Add("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to call roms list endpoint: %v", err)
	}
	defer resp.Body.Close()

	var rawItems []RomMRom
	err = json.NewDecoder(resp.Body).Decode(&rawItems)
	if err != nil {
		return nil, fmt.Errorf("failed to decode roms list JSON: %w", err)
	}

	var items []models.Item
	for _, rawItem := range rawItems {
		items = append(items, models.Item{
			Filename: rawItem.FsName,
			FileSize: strconv.Itoa(rawItem.FsSizeBytes),
			Date:     rawItem.UpdatedAt.String(),
			RomID:    strconv.Itoa(rawItem.ID),
		})
	}

	return items, nil
}

func (c *RomMClient) DownloadFile(remotePath, localPath, filename string) error {
	auth := c.Username + ":" + c.Password
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	u, err := url.Parse(c.buildRootURL() + RomsEndpoint + remotePath + "/content/" + filename)
	if err != nil {
		return fmt.Errorf("unable to parse url for rom download: %v", err)
	}

	params := url.Values{}
	params.Add("path", remotePath)

	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to build rom download request: %v", err)
	}

	req.Header.Add("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to call rom download endpoint: %v", err)
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(filepath.Join(localPath, filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
