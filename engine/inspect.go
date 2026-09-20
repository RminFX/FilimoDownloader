package engine

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"FilimoDownloader-GholamTaksir/internal/api"
	"FilimoDownloader-GholamTaksir/internal/helper"
	"FilimoDownloader-GholamTaksir/internal/stream"
)

func getUser(client helper.HttpClient) (string, error) {
	return api.GetUserNameErr(client)
}

func ParseContentID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err == nil && u.Path != "" {
			return path.Base(strings.TrimSuffix(u.Path, "/"))
		}
	}
	return raw
}

func (m *Manager) Inspect(query string) (*InspectResult, error) {
	client, err := m.client()
	if err != nil {
		return nil, err
	}
	id := ParseContentID(query)
	if id == "" {
		return nil, fmt.Errorf("لینک یا شناسه را بگذارید")
	}

	watch, err := api.GetEpisodeWatch(client, id)
	if err != nil {
		return nil, err
	}
	title := api.SeriesTitle(watch)
	if title == "" {
		title = id
	}
	season := watch.Data.Attributes.SeasonNumber
	episode := watch.Data.Attributes.EpisodeNumber
	kind := "movie"
	if api.IsSeriesWatch(watch) {
		kind = "series"
	}

	cover := watch.Data.Attributes.Cover
	if cover == "" {
		cover = watch.Data.Attributes.Thumb
	}
	if cover == "" {
		cover = watch.Data.Attributes.Poster
	}
	meta := api.FetchContentMeta(client, id, watch)
	if meta.PosterURL != "" {
		cover = meta.PosterURL
	} else if meta.ThumbURL != "" {
		cover = meta.ThumbURL
	}
	if meta.Title != "" {
		title = meta.Title
		if meta.SeriesTitle != "" && meta.IsSeries {
			title = meta.SeriesTitle
		}
	}

	result := &InspectResult{
		ID:          id,
		Title:       title,
		Kind:        kind,
		Season:      season,
		Episode:     episode,
		Cover:       cover,
		Description: meta.Description,
		Categories:  meta.Categories,
		SeriesTitle: meta.SeriesTitle,
		ParentID:    meta.ParentID,
		Qualities:   []Quality{},
		Audio:       []string{},
		Subtitles:   []string{},
		Episodes:    []api.Episode{},
	}
	if result.SeriesTitle == "" && kind == "series" {
		result.SeriesTitle = title
	}
	if result.ParentID == "" {
		result.ParentID = id
	}

	hls, err := stream.GetHlsErr(client, watch)
	if err == nil {
		for _, v := range hls.Variants {
			result.Qualities = append(result.Qualities, Quality{Quality: v.Quality, Resolution: v.Resolution})
		}
		for _, t := range hls.Tracks {
			result.Audio = append(result.Audio, t.Language)
		}
	}
	for _, s := range watch.Data.Attributes.Subtitles {
		result.Subtitles = append(result.Subtitles, s.Language)
	}

	if kind == "series" {
		eps := api.ListEpisodesFromWatch(watch)
		if len(eps) == 0 {
			eps = api.ListEpisodes(client, id)
		}
		if len(eps) == 0 {
			eps = []api.Episode{{
				ID:     id,
				Title:  strings.TrimSpace(watch.Data.Attributes.Name),
				Season: season,
				Number: episode,
				Cover:  cover,
			}}
			result.SeasonIncomplete = true
		} else {
			eps = api.EnrichEpisodes(client, eps)
		}
		result.Episodes = eps
		m.syncJobCovers(eps)
		if result.Season == 0 && len(eps) > 0 {
			result.Season = eps[0].Season
		}
	}

	return result, nil
}

func (m *Manager) syncJobCovers(eps []api.Episode) {
	if len(eps) == 0 {
		return
	}
	byID := map[string]string{}
	for _, ep := range eps {
		if strings.TrimSpace(ep.Cover) != "" {
			byID[ep.ID] = ep.Cover
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	changed := false
	for _, job := range m.jobs {
		if c, ok := byID[job.ContentID]; ok && c != "" && job.Cover != c {
			job.Cover = c
			changed = true
		}
	}
	if changed {
		m.saveQueueLocked()
	}
}
