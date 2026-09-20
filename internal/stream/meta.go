package stream

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"FilimoDownloader-GholamTaksir/internal/api"
	"FilimoDownloader-GholamTaksir/internal/helper"
)

func SaveSidecarMeta(dir string, meta api.ContentMeta, coverPath string) error {
	helper.MakeDirectories(dir)
	infoPath := path.Join(dir, "info.json")
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(infoPath, data, 0644); err != nil {
		return err
	}

	nfo := buildNFO(meta, coverPath)
	nfoName := "movie.nfo"
	if meta.IsSeries {
		nfoName = "episode.nfo"
	}
	return os.WriteFile(path.Join(dir, nfoName), []byte(nfo), 0644)
}

func DownloadCover(url, destDir string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", fmt.Errorf("آدرس پوستر خالی است")
	}
	helper.MakeDirectories(destDir)
	dest := path.Join(destDir, "cover.jpg")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", helper.GetUserAgent())
	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d برای پوستر", resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}
	return dest, nil
}

func ApplyFileMetadata(mediaPath string, meta api.ContentMeta, coverPath string) (string, error) {
	if !helper.IsFileExists(mediaPath) {
		return "", fmt.Errorf("فایل رسانه پیدا نشد")
	}
	ext := filepath.Ext(mediaPath)
	tmp := strings.TrimSuffix(mediaPath, ext) + ".meta" + ext

	args := []string{"-y", "-i", mediaPath}
	hasCover := coverPath != "" && helper.IsFileExists(coverPath)
	if hasCover {
		args = append(args, "-i", coverPath)
	}
	args = append(args, "-map", "0")
	if hasCover {
		args = append(args, "-map", "1", "-c", "copy", "-c:v:1", "mjpeg", "-disposition:v:1", "attached_pic")
	} else {
		args = append(args, "-c", "copy")
	}

	title := meta.Title
	show := meta.SeriesTitle
	if show == "" {
		show = title
	}
	args = append(args,
		"-metadata", "title="+title,
		"-metadata", "show="+show,
		"-metadata", "artist="+strings.Join(meta.Directors, ", "),
		"-metadata", "genre="+strings.Join(meta.Categories, ", "),
		"-metadata", "comment="+meta.Description,
		"-metadata", "description="+meta.Description,
		"-metadata", "synopsis="+meta.Description,
		"-metadata", "date="+meta.Year,
	)
	if meta.IsSeries {
		if meta.Season > 0 {
			args = append(args, "-metadata", fmt.Sprintf("season_number=%d", meta.Season))
		}
		if meta.Episode > 0 {
			args = append(args, "-metadata", fmt.Sprintf("episode_id=%d", meta.Episode))
			args = append(args, "-metadata", fmt.Sprintf("track=%d", meta.Episode))
		}
		args = append(args, "-metadata", "media_type=10") // TV show
	} else {
		args = append(args, "-metadata", "media_type=9") // Movie
	}
	args = append(args, tmp)

	if err := runFFmpeg("", args); err != nil {
		_ = os.Remove(tmp)
		return mediaPath, fmt.Errorf("نوشتن متادیتا ممکن نشد: %w", err)
	}
	if err := os.Rename(tmp, mediaPath); err != nil {
		_ = os.Remove(tmp)
		return mediaPath, err
	}
	return mediaPath, nil
}

func buildNFO(meta api.ContentMeta, coverPath string) string {
	var b strings.Builder
	root := "movie"
	if meta.IsSeries {
		root = "episodedetails"
	}
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>\n")
	b.WriteString("<" + root + ">\n")
	b.WriteString("  <title>" + xmlEscape(meta.Title) + "</title>\n")
	if meta.SeriesTitle != "" {
		b.WriteString("  <showtitle>" + xmlEscape(meta.SeriesTitle) + "</showtitle>\n")
	}
	if meta.Season > 0 {
		b.WriteString(fmt.Sprintf("  <season>%d</season>\n", meta.Season))
	}
	if meta.Episode > 0 {
		b.WriteString(fmt.Sprintf("  <episode>%d</episode>\n", meta.Episode))
	}
	if meta.Description != "" {
		b.WriteString("  <plot>" + xmlEscape(meta.Description) + "</plot>\n")
	}
	for _, g := range meta.Categories {
		b.WriteString("  <genre>" + xmlEscape(g) + "</genre>\n")
	}
	for _, c := range meta.Countries {
		b.WriteString("  <country>" + xmlEscape(c) + "</country>\n")
	}
	for _, d := range meta.Directors {
		b.WriteString("  <director>" + xmlEscape(d) + "</director>\n")
	}
	for _, a := range meta.Actors {
		b.WriteString("  <actor><name>" + xmlEscape(a) + "</name></actor>\n")
	}
	if coverPath != "" {
		b.WriteString("  <thumb>" + xmlEscape(filepath.Base(coverPath)) + "</thumb>\n")
	} else if meta.PosterURL != "" {
		b.WriteString("  <thumb>" + xmlEscape(meta.PosterURL) + "</thumb>\n")
	}
	b.WriteString("  <uniqueid type=\"filimo\">" + xmlEscape(meta.ID) + "</uniqueid>\n")
	b.WriteString("</" + root + ">\n")
	return b.String()
}

func xmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}
