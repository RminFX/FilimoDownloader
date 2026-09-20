package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"FilimoDownloader-GholamTaksir/internal/helper"
)

// GetEpisodeWatch دریافت اطلاعات یه قسمت - با error به جای panic
func GetEpisodeWatch(client helper.HttpClient, episodeID string) (Watch, error) {
	url := fmt.Sprintf("https://api.filimo.com/api/fa/v1/movie/watch/watch/uid/%s", episodeID)
	response, err := client.Get(url)
	if err != nil {
		return Watch{}, fmt.Errorf("failed to get episode %s: %w", episodeID, err)
	}
	var watch Watch
	if err := json.Unmarshal([]byte(response), &watch); err != nil {
		return Watch{}, fmt.Errorf("failed to parse episode %s: %w", episodeID, err)
	}
	return watch, nil
}

type Episode struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Season      int    `json:"season"`
	Number      int    `json:"number"`
	Duration    string `json:"duration"`
	SeasonTitle string `json:"seasonTitle"`
	Cover       string `json:"cover"`
	Poster      string `json:"poster"`
	Description string `json:"description"`
}

func WatchMeta(client helper.HttpClient, id string) (title string, season int, episode int, series bool, err error) {
	url := fmt.Sprintf("https://api.filimo.com/api/fa/v1/movie/watch/watch/uid/%s", id)
	response, err := client.Get(url)
	if err != nil {
		return "", 0, 0, false, err
	}
	var result struct {
		Data struct {
			Attributes struct {
				MovieTitle    string `json:"movie_title"`
				SeasonNumber  int    `json:"seasonNumber"`
				EpisodeNumber int    `json:"episodeNumber"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "", 0, 0, false, err
	}
	title = result.Data.Attributes.MovieTitle
	season = result.Data.Attributes.SeasonNumber
	episode = result.Data.Attributes.EpisodeNumber
	series = season > 0 || episode > 0 || (strings.Contains(title, "فصل") && strings.Contains(title, "قسمت"))
	return title, season, episode, series, nil
}

func ListEpisodesFromWatch(watch Watch) []Episode {
	seen := map[string]Episode{}
	for seasonIdx, season := range watch.Data.Attributes.SeriesData {
		seasonNo := parseSeasonNumber(season.Title)
		if seasonNo == 0 {
			seasonNo = seasonIdx + 1
		}
		seasonTitle := strings.TrimSpace(season.Title)
		if seasonTitle == "" {
			seasonTitle = fmt.Sprintf("فصل %d", seasonNo)
		}
		seasonCover := firstNonEmptyStr(season.Cover, season.Image, season.Thumb)
		for epIdx, raw := range season.Episodes {
			id := strings.TrimSpace(raw.UID)
			if id == "" {
				continue
			}
			title := strings.TrimSpace(raw.Title)
			if title == "" {
				title = fmt.Sprintf("قسمت %d", epIdx+1)
			}
			num := parseEpisodeNumber(title)
			if num == 0 {
				num = epIdx + 1
			}
			poster := upgradeThumbURL(firstNonEmptyStr(raw.Image, seasonCover))
			thumb := upgradeThumbURL(strings.TrimSpace(raw.Thumbplay))
			// Don't trust series-shared posters as episode art; Enrich fills real video frames.
			seen[id] = Episode{
				ID:          id,
				Title:       title,
				Season:      seasonNo,
				Number:      num,
				Duration:    strings.TrimSpace(raw.Duration),
				SeasonTitle: seasonTitle,
				Cover:       thumb,
				Poster:      poster,
				Description: strings.TrimSpace(raw.Description),
			}
		}
	}
	list := make([]Episode, 0, len(seen))
	for _, ep := range seen {
		list = append(list, ep)
	}
	sortEpisodes(list)
	return list
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// upgradeThumbURL asks Filimo CDN for a larger thumb so episode art is distinguishable.
func upgradeThumbURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	u = strings.Replace(u, "width=165", "width=400", 1)
	u = strings.Replace(u, "width=220", "width=480", 1)
	u = strings.Replace(u, "width=300", "width=480", 1)
	return u
}

func isGenericEpisodeTitle(title string) bool {
	t := strings.TrimSpace(toLatinDigits(title))
	if t == "" {
		return true
	}
	if episodeNumRe.MatchString(t) && len([]rune(t)) <= 12 {
		return true
	}
	return false
}

// EnrichEpisodes loads per-episode titles and unique video-frame covers from Filimo watch thumbs.
func EnrichEpisodes(client helper.HttpClient, eps []Episode) []Episode {
	if len(eps) == 0 {
		return eps
	}
	type result struct {
		idx   int
		title string
		cover string
	}
	out := make(chan result, len(eps))
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	for i := range eps {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			title, cover := enrichOneEpisode(client, eps[i])
			out <- result{idx: i, title: title, cover: cover}
		}(i)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	for r := range out {
		ep := &eps[r.idx]
		if r.title != "" && (isGenericEpisodeTitle(ep.Title) || len([]rune(r.title)) > len([]rune(ep.Title))) {
			ep.Title = r.title
		}
		if r.cover != "" {
			ep.Cover = r.cover
		}
	}
	return eps
}

func enrichOneEpisode(client helper.HttpClient, ep Episode) (title, cover string) {
	watch, err := GetEpisodeWatch(client, ep.ID)
	if err == nil {
		title = strings.TrimSpace(watch.Data.Attributes.Name)
		if title == "" {
			title = strings.TrimSpace(watch.Data.Attributes.MovieName)
		}
		if c, err := EpisodeCoverFromWatch(client, watch); err == nil && c != "" {
			cover = c
		}
	}
	if cover == "" {
		// Last resort: keep thumbplay only (may be wrong on rare CDN bugs).
		cover = strings.TrimSpace(ep.Cover)
	}
	if title == "" {
		cardTitle, _ := fetchEpisodeCard(client, ep.ID)
		title = cardTitle
	}
	return title, cover
}

func IsSeriesWatch(watch Watch) bool {
	a := watch.Data.Attributes
	if len(a.SeriesData) > 0 {
		return true
	}
	if a.SeasonNumber > 0 || a.EpisodeNumber > 0 {
		return true
	}
	if strings.TrimSpace(a.PreTitle) == "سریال" {
		return true
	}
	title := a.Name
	return strings.Contains(title, "فصل") && (strings.Contains(title, "قسمت") || strings.Contains(title, "جلسه"))
}

func SeriesTitle(watch Watch) string {
	name := strings.TrimSpace(watch.Data.Attributes.Name)
	if name == "" {
		name = strings.TrimSpace(watch.Data.Attributes.MovieName)
	}
	cleaned := episodeSuffix.ReplaceAllString(toLatinDigits(name), "")
	cleaned = strings.TrimSpace(strings.Trim(cleaned, ":-–— "))
	if cleaned != "" {
		return cleaned
	}
	return name
}

var (
	digitReplacer = strings.NewReplacer(
		"۰", "0", "۱", "1", "۲", "2", "۳", "3", "۴", "4",
		"۵", "5", "۶", "6", "۷", "7", "۸", "8", "۹", "9",
		"٠", "0", "١", "1", "٢", "2", "٣", "3", "٤", "4",
		"٥", "5", "٦", "6", "٧", "7", "٨", "8", "٩", "9",
	)
	episodeNumRe  = regexp.MustCompile(`(?:قسمت|جلسه|episode|ep)\s*(\d+)`)
	seasonNumRe   = regexp.MustCompile(`فصل\s*(\d+)`)
	episodeSuffix = regexp.MustCompile(`[:\s]*(?:جلسه|قسمت|episode|ep)\s*\d+\s*$`)
)

func toLatinDigits(s string) string {
	return digitReplacer.Replace(s)
}

func parseEpisodeNumber(title string) int {
	m := episodeNumRe.FindStringSubmatch(toLatinDigits(title))
	if len(m) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

func parseSeasonNumber(title string) int {
	t := toLatinDigits(strings.TrimSpace(title))
	if m := seasonNumRe.FindStringSubmatch(t); len(m) >= 2 {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	words := map[string]int{
		"اول": 1, "دوم": 2, "سوم": 3, "چهارم": 4, "پنجم": 5,
		"ششم": 6, "هفتم": 7, "هشتم": 8, "نهم": 9, "دهم": 10,
	}
	for word, n := range words {
		if strings.Contains(t, word) {
			return n
		}
	}
	return 0
}

func ListEpisodes(client helper.HttpClient, id string) []Episode {
	seen := map[string]Episode{}
	urls := []string{
		fmt.Sprintf("https://api.filimo.com/api/fa/v1/movie/movie/one/uid/%s", id),
		fmt.Sprintf("https://www.filimo.com/api/fa/v1/movie/movie/one/uid/%s", id),
		fmt.Sprintf("https://api.filimo.com/api/fa/v1/movie/serial/each/parent_id/%s", id),
		fmt.Sprintf("https://www.filimo.com/api/fa/v1/movie/serial/each/id/0/parent_id/%s", id),
	}
	for _, u := range urls {
		body, err := client.Get(u)
		if err != nil || strings.TrimSpace(body) == "" {
			continue
		}
		var raw any
		if err := json.Unmarshal([]byte(body), &raw); err != nil {
			continue
		}
		walkEpisodes(raw, seen)
	}
	list := make([]Episode, 0, len(seen))
	for _, ep := range seen {
		list = append(list, ep)
	}
	sortEpisodes(list)
	return list
}

func walkEpisodes(node any, seen map[string]Episode) {
	switch n := node.(type) {
	case []any:
		for _, child := range n {
			walkEpisodes(child, seen)
		}
	case map[string]any:
		if ep, ok := episodeFromMap(n); ok {
			seen[ep.ID] = ep
		}
		for _, child := range n {
			walkEpisodes(child, seen)
		}
	}
}

func episodeFromMap(m map[string]any) (Episode, bool) {
	attrs, _ := m["attributes"].(map[string]any)
	id := firstString(m, "uid", "link_key", "id", "movie_uid")
	if id == "" && attrs != nil {
		id = firstString(attrs, "uid", "link_key", "id", "movie_uid")
	}
	if id == "" || looksLikeNonContentID(id) {
		return Episode{}, false
	}
	title := firstString(m, "movie_title", "title", "name")
	if title == "" && attrs != nil {
		title = firstString(attrs, "movie_title", "title", "name")
	}
	season := firstInt(m, "seasonNumber", "season", "season_number")
	episode := firstInt(m, "episodeNumber", "episode", "episode_number")
	if attrs != nil {
		if season == 0 {
			season = firstInt(attrs, "seasonNumber", "season", "season_number")
		}
		if episode == 0 {
			episode = firstInt(attrs, "episodeNumber", "episode", "episode_number")
		}
	}
	if episode == 0 && !(strings.Contains(title, "قسمت") || strings.Contains(title, "فصل")) {
		return Episode{}, false
	}
	if title == "" {
		title = id
	}
	cover := firstString(m, "image", "thumbplay", "poster", "cover", "thumb", "movie_img_m", "movie_img_s")
	if cover == "" && attrs != nil {
		cover = firstString(attrs, "image", "thumbplay", "poster", "cover", "thumb", "movie_img_m", "movie_img_s")
	}
	duration := firstString(m, "duration", "movie_duration", "time")
	if duration == "" && attrs != nil {
		duration = firstString(attrs, "duration", "movie_duration", "time")
	}
	return Episode{ID: id, Title: title, Season: season, Number: episode, Cover: cover, Duration: duration}, true
}

func looksLikeNonContentID(id string) bool {
	if strings.Contains(id, "http") || strings.Contains(id, "/") {
		return true
	}
	if len(id) < 3 {
		return true
	}
	return false
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		switch v := m[key].(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		case float64:
			return fmt.Sprintf("%.0f", v)
		case json.Number:
			return v.String()
		}
	}
	return ""
}

func firstInt(m map[string]any, keys ...string) int {
	for _, key := range keys {
		switch v := m[key].(type) {
		case float64:
			return int(v)
		case int:
			return v
		case json.Number:
			n, _ := v.Int64()
			return int(n)
		case string:
			n := 0
			fmt.Sscanf(v, "%d", &n)
			if n > 0 {
				return n
			}
		}
	}
	return 0
}

func sortEpisodes(list []Episode) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Season != list[j].Season {
			return list[i].Season < list[j].Season
		}
		return list[i].Number < list[j].Number
	})
}

// fetchEpisodeCard loads unique title + poster for one episode uid.
func fetchEpisodeCard(client helper.HttpClient, id string) (title, cover string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", ""
	}
	body, err := client.Get(fmt.Sprintf("https://www.filimo.com/api/fa/v1/movie/movie/one/uid/%s", id))
	if err != nil || strings.TrimSpace(body) == "" {
		body, err = client.Get(fmt.Sprintf("https://api.filimo.com/api/fa/v1/movie/movie/one/uid/%s", id))
		if err != nil || strings.TrimSpace(body) == "" {
			return "", ""
		}
	}
	var one struct {
		Data struct {
			Attributes struct {
				General struct {
					Title   string `json:"title"`
					TitleFA string `json:"title_fa"`
					Thumbnails struct {
						S string `json:"movie_img_s"`
						M string `json:"movie_img_m"`
						B string `json:"movie_img_b"`
					} `json:"thumbnails"`
				} `json:"General"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &one); err != nil {
		return "", ""
	}
	g := one.Data.Attributes.General
	title = strings.TrimSpace(firstNonEmptyStr(g.TitleFA, g.Title))
	cover = firstNonEmptyStr(g.Thumbnails.M, g.Thumbnails.B, g.Thumbnails.S)
	return title, cover
}

// IsSeries چک میکنه آیا ID یه قسمت سریاله
func IsSeries(client helper.HttpClient, id string) bool {
	url := fmt.Sprintf("https://api.filimo.com/api/fa/v1/movie/watch/watch/uid/%s", id)
	response, err := client.Get(url)
	if err != nil {
		return false
	}

	var result struct {
		Data struct {
			Attributes struct {
				MovieTitle    string `json:"movie_title"`
				SeasonNumber  int    `json:"seasonNumber"`
				EpisodeNumber int    `json:"episodeNumber"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return false
	}

	if result.Data.Attributes.SeasonNumber > 0 || result.Data.Attributes.EpisodeNumber > 0 {
		return true
	}

	title := result.Data.Attributes.MovieTitle
	return strings.Contains(title, "فصل") && strings.Contains(title, "قسمت")
}
