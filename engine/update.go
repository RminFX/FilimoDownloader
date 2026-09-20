package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const AppVersion = "2.2.0"

type UpdateInfo struct {
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	URL       string `json:"url"`
	Available bool   `json:"available"`
	Message   string `json:"message"`
}

func Version() string {
	return AppVersion
}

func CheckForUpdates() UpdateInfo {
	info := UpdateInfo{
		Current: AppVersion,
		URL:     "https://github.com/Gholam-Taksir/FilimoDownloader/releases",
		Message: "بررسی شد",
	}
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/Gholam-Taksir/FilimoDownloader/releases/latest", nil)
	if err != nil {
		info.Message = "بررسی آپدیت ممکن نبود"
		return info
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "FilimoDownloader/"+AppVersion)
	resp, err := client.Do(req)
	if err != nil {
		info.Message = "اتصال برای بررسی آپدیت برقرار نشد"
		return info
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		info.Message = "هنوز ریلیزی منتشر نشده"
		return info
	}
	if resp.StatusCode >= 400 {
		info.Message = fmt.Sprintf("پاسخ نامعتبر از گیت‌هاب (%d)", resp.StatusCode)
		return info
	}
	var payload struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Name    string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		info.Message = "خواندن اطلاعات آپدیت ممکن نبود"
		return info
	}
	latest := strings.TrimPrefix(strings.TrimSpace(payload.TagName), "v")
	info.Latest = latest
	if payload.HTMLURL != "" {
		info.URL = payload.HTMLURL
	}
	if latest != "" && versionLess(AppVersion, latest) {
		info.Available = true
		info.Message = fmt.Sprintf("نسخه جدید %s آماده است", latest)
	} else {
		info.Message = "روی آخرین نسخه هستید"
	}
	return info
}

func versionLess(a, b string) bool {
	ap := splitVersion(a)
	bp := splitVersion(b)
	n := len(ap)
	if len(bp) > n {
		n = len(bp)
	}
	for i := 0; i < n; i++ {
		ai, bi := 0, 0
		if i < len(ap) {
			ai = ap[i]
		}
		if i < len(bp) {
			bi = bp[i]
		}
		if ai < bi {
			return true
		}
		if ai > bi {
			return false
		}
	}
	return false
}

func splitVersion(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n := 0
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				break
			}
			n = n*10 + int(ch-'0')
		}
		out = append(out, n)
	}
	return out
}
