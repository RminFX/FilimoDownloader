package config

import (
	"encoding/json"
	"os"
	"path"

	"FilimoDownloader-GholamTaksir/internal/helper"
)

type Settings struct {
	DefaultQuality       string `json:"default_quality"`
	DefaultFormat        string `json:"default_format"`
	DownloadPath         string `json:"download_path"`
	MaxThreads           int    `json:"max_threads"`
	AutoOpenFolder       bool   `json:"auto_open_folder"`
	ShowInfoBeforeDL     bool   `json:"show_info_before_dl"`
	Theme                string `json:"theme"`
	Language             string `json:"language"`
	Notifications        bool   `json:"notifications"`
	CheckUpdatesOnLaunch bool   `json:"check_updates_on_launch"`
}

var DefaultSettings = Settings{
	DefaultQuality:       "720p",
	DefaultFormat:        "mp4",
	DownloadPath:         "Downloads",
	MaxThreads:           2,
	AutoOpenFolder:       true,
	ShowInfoBeforeDL:     true,
	Theme:                "system",
	Language:             "fa",
	Notifications:        true,
	CheckUpdatesOnLaunch: true,
}

// DataDir returns the data subdirectory path
func DataDir(basePath string) string {
	return path.Join(basePath, "data")
}

func GetConfigPath(basePath string) string {
	return path.Join(DataDir(basePath), "config.json")
}

func Load(basePath string) *Settings {
	// Ensure data directory exists
	helper.MakeDirectories(DataDir(basePath))

	configPath := GetConfigPath(basePath)
	if !helper.IsFileExists(configPath) {
		Save(basePath, &DefaultSettings)
		return &DefaultSettings
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return &DefaultSettings
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return &DefaultSettings
	}
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)
	if _, ok := raw["notifications"]; !ok {
		settings.Notifications = true
	}
	if _, ok := raw["check_updates_on_launch"]; !ok {
		settings.CheckUpdatesOnLaunch = true
	}
	if _, ok := raw["theme"]; !ok {
		settings.Theme = "system"
	}
	normalizeSettings(&settings)
	return &settings
}

func normalizeSettings(s *Settings) {
	if s.DefaultQuality == "" {
		s.DefaultQuality = DefaultSettings.DefaultQuality
	}
	if s.DefaultFormat == "" {
		s.DefaultFormat = DefaultSettings.DefaultFormat
	}
	if s.MaxThreads < 1 {
		s.MaxThreads = DefaultSettings.MaxThreads
	}
	if s.MaxThreads > 5 {
		s.MaxThreads = 5
	}
	if s.Theme == "" {
		s.Theme = DefaultSettings.Theme
	}
	switch s.Theme {
	case "light", "dark", "system":
	default:
		s.Theme = "system"
	}
	if s.Language == "" {
		s.Language = DefaultSettings.Language
	}
	switch s.Language {
	case "fa", "en", "ar", "tr":
	default:
		s.Language = "fa"
	}
}

func Save(basePath string, settings *Settings) {
	helper.MakeDirectories(DataDir(basePath))
	normalizeSettings(settings)
	configPath := GetConfigPath(basePath)
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return
	}
	helper.WriteFile(configPath, string(data))
}
