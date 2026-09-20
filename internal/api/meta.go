package api

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"

	"FilimoDownloader-GholamTaksir/internal/helper"
)

type ContentMeta struct {
	ID          string   `json:"id"`
	UID         string   `json:"uid"`
	Title       string   `json:"title"`
	TitleEN     string   `json:"titleEn"`
	Description string   `json:"description"`
	Year        string   `json:"year"`
	Season      int      `json:"season"`
	Episode     int      `json:"episode"`
	SeasonText  string   `json:"seasonText"`
	EpisodeText string   `json:"episodeText"`
	SeriesTitle string   `json:"seriesTitle"`
	Categories  []string `json:"categories"`
	Countries   []string `json:"countries"`
	Directors   []string `json:"directors"`
	Actors      []string `json:"actors"`
	PosterURL   string   `json:"posterUrl"`
	ThumbURL    string   `json:"thumbUrl"`
	CoverURL    string   `json:"coverUrl"`
	ParentID    string   `json:"parentId"`
	IsSeries    bool     `json:"isSeries"`
}

var stripTags = regexp.MustCompile(`<[^>]*>`)

func FetchContentMeta(client helper.HttpClient, id string, watch Watch) ContentMeta {
	meta := ContentMeta{
		ID:        id,
		Title:     strings.TrimSpace(watch.Data.Attributes.Name),
		PosterURL: firstNonEmpty(watch.Data.Attributes.Poster, watch.Data.Attributes.Cover, watch.Data.Attributes.Thumb),
		CoverURL:  firstNonEmpty(watch.Data.Attributes.Cover, watch.Data.Attributes.Poster),
		ThumbURL:  watch.Data.Attributes.Thumb,
		IsSeries:  IsSeriesWatch(watch),
		Season:    watch.Data.Attributes.SeasonNumber,
		Episode:   watch.Data.Attributes.EpisodeNumber,
	}
	if meta.Title == "" {
		meta.Title = strings.TrimSpace(watch.Data.Attributes.MovieName)
	}
	if meta.IsSeries {
		meta.SeriesTitle = SeriesTitle(watch)
	}

	// تصویر قسمت از seriesData
	for _, season := range watch.Data.Attributes.SeriesData {
		for _, ep := range season.Episodes {
			if ep.UID == id || (meta.UID != "" && ep.UID == meta.UID) {
				if ep.Image != "" {
					meta.ThumbURL = ep.Image
					if meta.PosterURL == "" {
						meta.PosterURL = ep.Image
					}
				}
				if ep.Thumbplay != "" && meta.CoverURL == "" {
					meta.CoverURL = ep.Thumbplay
				}
			}
		}
	}

	body, err := client.Get(fmt.Sprintf("https://api.filimo.com/api/fa/v1/movie/movie/one/uid/%s", id))
	if err != nil || strings.TrimSpace(body) == "" {
		return meta
	}
	var one struct {
		Data struct {
			Attributes struct {
				General struct {
					UID                string `json:"uid"`
					Title              string `json:"title"`
					TitleEN            string `json:"title_en"`
					TitleFA            string `json:"title_fa"`
					About              string `json:"about_movie"`
					SeasonCount        int    `json:"season_count"`
					EpisodeSeasonCount string `json:"episode_season_count"`
					Thumbnails         struct {
						S string `json:"movie_img_s"`
						M string `json:"movie_img_m"`
						B string `json:"movie_img_b"`
					} `json:"thumbnails"`
					Serial struct {
						Enable      bool   `json:"enable"`
						ParentID    string `json:"parent_id"`
						SeasonText  string `json:"season_text"`
						EpisodeText string `json:"episode_text"`
						SeasonID    string `json:"season_id"`
						SerialPart  string `json:"serial_part"`
					} `json:"serial"`
					Categories []struct {
						Title string `json:"title"`
					} `json:"categories"`
					Countries []struct {
						Title string `json:"title"`
					} `json:"countries"`
					Director []struct {
						Name string `json:"name"`
					} `json:"director"`
					Actors []struct {
						Name string `json:"name"`
					} `json:"actors"`
				} `json:"General"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &one); err != nil {
		return meta
	}
	g := one.Data.Attributes.General
	if g.UID != "" {
		meta.UID = g.UID
	}
	if t := firstNonEmpty(g.TitleFA, g.Title); t != "" {
		meta.Title = t
	}
	meta.TitleEN = g.TitleEN
	meta.Description = cleanHTML(g.About)
	meta.SeasonText = g.Serial.SeasonText
	meta.EpisodeText = g.Serial.EpisodeText
	meta.ParentID = g.Serial.ParentID
	if g.Serial.Enable {
		meta.IsSeries = true
	}
	if meta.Season == 0 {
		meta.Season = atoiSafe(g.Serial.SeasonID)
	}
	if meta.Episode == 0 {
		meta.Episode = atoiSafe(g.Serial.SerialPart)
	}
	if meta.SeriesTitle == "" && meta.IsSeries {
		meta.SeriesTitle = SeriesTitle(watch)
		if meta.SeriesTitle == "" {
			meta.SeriesTitle = stripEpisodeSuffix(meta.Title)
		}
	}
	if g.Thumbnails.B != "" {
		meta.PosterURL = g.Thumbnails.B
	} else if g.Thumbnails.M != "" && meta.PosterURL == "" {
		meta.PosterURL = g.Thumbnails.M
	}
	if g.Thumbnails.S != "" {
		meta.ThumbURL = g.Thumbnails.S
	}
	for _, c := range g.Categories {
		if c.Title != "" {
			meta.Categories = append(meta.Categories, c.Title)
		}
	}
	for _, c := range g.Countries {
		if c.Title != "" {
			meta.Countries = append(meta.Countries, c.Title)
		}
	}
	for _, d := range g.Director {
		if d.Name != "" {
			meta.Directors = append(meta.Directors, d.Name)
		}
	}
	for _, a := range g.Actors {
		if a.Name != "" {
			meta.Actors = append(meta.Actors, a.Name)
		}
	}
	return meta
}

func cleanHTML(s string) string {
	s = html.UnescapeString(s)
	s = stripTags.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func atoiSafe(s string) int {
	s = strings.TrimSpace(toLatinDigits(s))
	n := 0
	fmt.Sscanf(s, "%d", &n)
	return n
}

func stripEpisodeSuffix(title string) string {
	cleaned := episodeSuffix.ReplaceAllString(toLatinDigits(title), "")
	return strings.TrimSpace(strings.Trim(cleaned, ":-–— "))
}
