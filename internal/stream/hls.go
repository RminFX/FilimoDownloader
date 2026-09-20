package stream

import (
	"fmt"
	"regexp"
	"strings"

	"FilimoDownloader-GholamTaksir/internal/api"
	"FilimoDownloader-GholamTaksir/internal/helper"
)

type Hls struct {
	Variants []HlsVideoVariant
	Tracks   []HlsAudioTrack
}

type HlsVideoVariant struct {
	Quality    string
	Resolution string
	Link       string
}

type HlsAudioTrack struct {
	Language string
	Link     string
}

func GetHls(client helper.HttpClient, watch api.Watch) Hls {
	hls, err := GetHlsErr(client, watch)
	if err != nil {
		helper.ShowErrorAndExit(err.Error())
	}
	return hls
}

func GetHlsErr(client helper.HttpClient, watch api.Watch) (Hls, error) {
	l := extractHlsLink(watch)
	if l == "" {
		return Hls{}, fmt.Errorf("لینک پخش پیدا نشد")
	}
	hlsContent, err := client.Get(l)
	if err != nil {
		return Hls{}, fmt.Errorf("خواندن کیفیت‌ها ممکن نشد: %w", err)
	}
	variants, err := parseVariants(hlsContent)
	if err != nil {
		return Hls{}, err
	}
	return Hls{
		Variants: variants,
		Tracks:   parseTracks(hlsContent),
	}, nil
}

func extractHlsLink(watch api.Watch) string {
	for _, list := range watch.Data.Attributes.Sources {
		for _, source := range list {
			if source.Type == "application/vnd.apple.mpegurl" {
				return source.Link
			}
		}
	}
	return ""
}

// parseVariants - FIX: more robust regex that handles various M3U8 formats
func parseVariants(hls string) ([]HlsVideoVariant, error) {
	list := []HlsVideoVariant{}

	lines := strings.Split(hls, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			continue
		}

		// Extract BANDWIDTH or NAME as quality label
		quality := ""
		bwPattern := regexp.MustCompile(`BANDWIDTH=(\d+)`)
		if m := bwPattern.FindStringSubmatch(line); len(m) > 1 {
			bw := m[1]
			// Convert bandwidth to human-readable
			switch {
			case len(bw) >= 7:
				quality = "1080p"
			case len(bw) >= 6:
				quality = "720p"
			default:
				quality = "360p"
			}
		}

		// Extract RESOLUTION
		resolution := ""
		resPattern := regexp.MustCompile(`RESOLUTION=([0-9x]+)`)
		if m := resPattern.FindStringSubmatch(line); len(m) > 1 {
			resolution = m[1]
			// Derive quality from resolution height
			parts := strings.Split(resolution, "x")
			if len(parts) == 2 {
				quality = parts[1] + "p"
			}
		}

		// Next non-empty line is the URL
		link := ""
		for j := i + 1; j < len(lines); j++ {
			candidate := strings.TrimSpace(lines[j])
			if candidate != "" && !strings.HasPrefix(candidate, "#") {
				link = candidate
				break
			}
		}

		if link == "" {
			continue
		}

		list = append(list, HlsVideoVariant{
			Quality:    quality,
			Resolution: resolution,
			Link:       link,
		})
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("هیچ کیفیتی در پخش پیدا نشد")
	}
	return list, nil
}

// parseTracks - FIX: more robust regex for audio tracks
func parseTracks(hls string) []HlsAudioTrack {
	list := []HlsAudioTrack{}
	seen := make(map[string]bool)

	pattern := regexp.MustCompile(`#EXT-X-MEDIA:TYPE=AUDIO[^\n]*LANGUAGE="([^"]*)"[^\n]*URI="([^"]*)"`)
	for _, m := range pattern.FindAllStringSubmatch(hls, -1) {
		if len(m) < 3 {
			continue
		}
		lang := m[1]
		link := m[2]
		if link == "" || seen[lang] {
			continue
		}
		seen[lang] = true
		list = append(list, HlsAudioTrack{
			Language: lang,
			Link:     link,
		})
	}
	return list
}
