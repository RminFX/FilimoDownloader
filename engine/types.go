package engine

import "FilimoDownloader-GholamTaksir/internal/api"

type Session struct {
	LoggedIn bool   `json:"loggedIn"`
	Username string `json:"username"`
	HasToken bool   `json:"hasToken"`
}

type Quality struct {
	Quality    string `json:"quality"`
	Resolution string `json:"resolution"`
}

type InspectResult struct {
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	Kind             string        `json:"kind"`
	Season           int           `json:"season"`
	Episode          int           `json:"episode"`
	Cover            string        `json:"cover"`
	Description      string        `json:"description"`
	Categories       []string      `json:"categories"`
	Qualities        []Quality     `json:"qualities"`
	Audio            []string      `json:"audio"`
	Subtitles        []string      `json:"subtitles"`
	Episodes         []api.Episode `json:"episodes"`
	SeasonIncomplete bool          `json:"seasonIncomplete"`
	SeriesTitle      string        `json:"seriesTitle"`
	ParentID         string        `json:"parentId"`
}

type QueueItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Kind        string `json:"kind"`
	SeriesKey   string `json:"seriesKey"`
	SeriesTitle string `json:"seriesTitle"`
	Season      int    `json:"season"`
	Episode     int    `json:"episode"`
	SeasonTitle string `json:"seasonTitle"`
	Cover       string `json:"cover"`
}

type EnqueueRequest struct {
	Items     []QueueItem `json:"items"`
	Quality   string      `json:"quality"`
	Audio     []string    `json:"audio"`
	Subtitles []string    `json:"subtitles"`
	Format    string      `json:"format"`
}

type Job struct {
	ID          string   `json:"id"`
	ContentID   string   `json:"contentId"`
	Title       string   `json:"title"`
	Quality     string   `json:"quality"`
	Format      string   `json:"format"`
	Status      string   `json:"status"`
	Progress    float64  `json:"progress"`
	Message     string   `json:"message"`
	FilePath    string   `json:"filePath"`
	Error       string   `json:"error"`
	Kind        string   `json:"kind"`
	SeriesKey   string   `json:"seriesKey"`
	SeriesTitle string   `json:"seriesTitle"`
	Season      int      `json:"season"`
	Episode     int      `json:"episode"`
	SeasonTitle string   `json:"seasonTitle"`
	Cover       string   `json:"cover"`
	Audio       []string `json:"audio"`
	Subs        []string `json:"subs"`
	stopMode    string
}

type Settings struct {
	DefaultQuality       string `json:"defaultQuality"`
	DefaultFormat        string `json:"defaultFormat"`
	DownloadPath         string `json:"downloadPath"`
	AutoOpenFolder       bool   `json:"autoOpenFolder"`
	ShowInfoBeforeDL     bool   `json:"showInfoBeforeDL"`
	ConcurrentDownloads  int    `json:"concurrentDownloads"`
	Theme                string `json:"theme"`
	Language             string `json:"language"`
	Notifications        bool   `json:"notifications"`
	CheckUpdatesOnLaunch bool   `json:"checkUpdatesOnLaunch"`
}

type HistoryItem struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Quality      string  `json:"quality"`
	Format       string  `json:"format"`
	SizeMB       float64 `json:"sizeMb"`
	DownloadedAt string  `json:"downloadedAt"`
	FilePath     string  `json:"filePath"`
}

type LibrarySeries struct {
	Key          string  `json:"key"`
	Title        string  `json:"title"`
	Cover        string  `json:"cover"`
	Kind         string  `json:"kind"`
	SeasonCount  int     `json:"seasonCount"`
	EpisodeCount int     `json:"episodeCount"`
	DoneCount    int     `json:"doneCount"`
	ActiveCount  int     `json:"activeCount"`
	Progress     float64 `json:"progress"`
}

type LibrarySeason struct {
	SeriesKey    string  `json:"seriesKey"`
	Season       int     `json:"season"`
	Title        string  `json:"title"`
	Cover        string  `json:"cover"`
	EpisodeCount int     `json:"episodeCount"`
	DoneCount    int     `json:"doneCount"`
	Progress     float64 `json:"progress"`
}
