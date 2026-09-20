package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"regexp"
	"strconv"
	"strings"
	"time"

	"FilimoDownloader-GholamTaksir/internal/helper"
)

var (
	vttCueRe = regexp.MustCompile(`(?m)^(\d{2}):(\d{2})\.(\d{3})\s+-->\s+.*\n([^\s#]+)#xywh=(\d+),(\d+),(\d+),(\d+)`)
)

// EpisodeCoverFromWatch extracts a real video-frame thumbnail from Filimo seek sprites.
// This is per-episode and avoids shared series posters / broken thumbplay assets.
func EpisodeCoverFromWatch(client helper.HttpClient, watch Watch) (string, error) {
	th := watch.Data.Attributes.Thumbs
	if th == nil {
		return "", fmt.Errorf("no thumbs")
	}
	// Prefer JPEG sprite sheet (no extra deps); webp is fallback.
	vttURL := firstNonEmptyStr(th.Src, th.WebpMSrc, th.WebpTSrc)
	if vttURL == "" {
		return "", fmt.Errorf("empty thumbs src")
	}
	vttURL = strings.ReplaceAll(vttURL, " ", "")
	body, err := client.Get(vttURL)
	if err != nil {
		return "", err
	}
	spriteRel, x, y, w, h, ok := pickVTTCue(body, 25*time.Second)
	if !ok {
		return "", fmt.Errorf("no vtt cue")
	}
	spriteURL := resolveSiblingURL(vttURL, spriteRel)
	raw, err := client.GetBytes(spriteURL)
	if err != nil {
		return "", err
	}
	img, err := decodeAnyImage(raw)
	if err != nil {
		return "", err
	}
	cropped, err := cropImage(img, x, y, w, h)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, cropped, &jpeg.Options{Quality: 85}); err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func pickVTTCue(vtt string, at time.Duration) (file string, x, y, w, h int, ok bool) {
	matches := vttCueRe.FindAllStringSubmatch(vtt, -1)
	if len(matches) == 0 {
		return "", 0, 0, 0, 0, false
	}
	targetMs := int(at / time.Millisecond)
	best := matches[0]
	bestDiff := int(1 << 30)
	for _, m := range matches {
		startMs := vttTimeMs(m[1], m[2], m[3])
		diff := startMs - targetMs
		if diff < 0 {
			diff = -diff
		}
		if diff < bestDiff {
			bestDiff = diff
			best = m
		}
	}
	x, _ = strconv.Atoi(best[5])
	y, _ = strconv.Atoi(best[6])
	w, _ = strconv.Atoi(best[7])
	h, _ = strconv.Atoi(best[8])
	return best[4], x, y, w, h, w > 0 && h > 0
}

func vttTimeMs(mm, ss, ms string) int {
	m, _ := strconv.Atoi(mm)
	s, _ := strconv.Atoi(ss)
	milli, _ := strconv.Atoi(ms)
	return (m*60+s)*1000 + milli
}

func resolveSiblingURL(vttURL, rel string) string {
	rel = strings.TrimSpace(rel)
	if strings.HasPrefix(rel, "http://") || strings.HasPrefix(rel, "https://") {
		return rel
	}
	// Filimo sometimes returns https://host//filimo-video/...
	base := strings.TrimSpace(vttURL)
	if i := strings.LastIndex(base, "/"); i >= 0 {
		return base[:i+1] + rel
	}
	return rel
}

func decodeAnyImage(raw []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func cropImage(img image.Image, x, y, w, h int) (image.Image, error) {
	b := img.Bounds()
	r := image.Rect(x, y, x+w, y+h).Intersect(b)
	if r.Empty() {
		return nil, fmt.Errorf("empty crop")
	}
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	if s, ok := img.(subImager); ok {
		return s.SubImage(r), nil
	}
	dst := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for yy := r.Min.Y; yy < r.Max.Y; yy++ {
		for xx := r.Min.X; xx < r.Max.X; xx++ {
			dst.Set(xx-r.Min.X, yy-r.Min.Y, img.At(xx, yy))
		}
	}
	return dst, nil
}
