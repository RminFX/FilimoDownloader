package stream

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"FilimoDownloader-GholamTaksir/internal/helper"
)

type Playlist struct {
	Content string
	Urls    []*url.URL
}

func GetPlaylist(client helper.HttpClient, link string) Playlist {
	playlist, err := GetPlaylistErr(client, link)
	if err != nil {
		helper.ShowErrorAndExit(err.Error())
	}
	return playlist
}

func GetPlaylistErr(client helper.HttpClient, link string) (Playlist, error) {
	playlist, err := client.Get(link)
	if err != nil {
		return Playlist{}, fmt.Errorf("خواندن لیست قطعه‌ها ممکن نشد: %w", err)
	}
	urls, err := extractUrls(playlist, link)
	if err != nil {
		return Playlist{}, err
	}
	return Playlist{
		Content: cleanContent(playlist),
		Urls:    urls,
	}, nil
}

func cleanContent(playlist string) string {
	pattern := `\?([^"\n]*)`
	return regexp.MustCompile(pattern).ReplaceAllString(playlist, "")
}

func extractUrls(playlist string, link string) ([]*url.URL, error) {
	urls := []*url.URL{}

	keyPattern := regexp.MustCompile(`#EXT-X-KEY[^\n]*URI="([^"]*)"`)
	keyMatches := keyPattern.FindStringSubmatch(playlist)

	if len(keyMatches) < 2 || keyMatches[1] == "" {
		return nil, fmt.Errorf("کلید رمز قطعه پیدا نشد")
	}
	urls = append(urls, helper.AbsoluteUrl(link, keyMatches[1]))

	chunksPattern := regexp.MustCompile(`(?m)^([^#\s].*)`)
	for _, chunk := range chunksPattern.FindAllString(playlist, -1) {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		urls = append(urls, helper.AbsoluteUrl(link, chunk))
	}

	if len(urls) <= 1 {
		return nil, fmt.Errorf("قطعه‌ای برای دانلود پیدا نشد")
	}

	return urls, nil
}
