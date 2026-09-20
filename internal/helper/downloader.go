package helper

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

type DownloadStats struct {
	TotalSize     int64
	Downloaded    int64
	StartTime     time.Time
	LastSpeedTime time.Time
	LastSpeedPos  int64
	Speed         float64
}

func (ds *DownloadStats) UpdateSpeed() {
	now := time.Now()
	elapsed := now.Sub(ds.LastSpeedTime).Seconds()
	if elapsed >= 0.5 {
		bytesDownloaded := ds.Downloaded - ds.LastSpeedPos
		ds.Speed = float64(bytesDownloaded) / elapsed / 1024 / 1024
		ds.LastSpeedTime = now
		ds.LastSpeedPos = ds.Downloaded
	}
}

func (ds *DownloadStats) PrintProgress() {
	if ds.TotalSize <= 0 {
		return
	}
	percentage := float64(ds.Downloaded) / float64(ds.TotalSize) * 100
	var eta float64
	if ds.Speed > 0 {
		remaining := float64(ds.TotalSize-ds.Downloaded) / 1024 / 1024
		eta = remaining / ds.Speed
	}
	fmt.Printf("\r  [%.1f%%] %.2f / %.2f MB | %.2f MB/s | ETA: %ds     ",
		percentage,
		float64(ds.Downloaded)/1024/1024,
		float64(ds.TotalSize)/1024/1024,
		ds.Speed,
		int(eta),
	)
}

// DownloadFileWithResume دانلود با قابلیت ادامه + retry داخلی
func DownloadFileWithResume(url string, filepath string, token string) error {
	dir := path.Dir(filepath)
	MakeDirectories(dir)

	const maxRetries = 5
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := downloadAttempt(url, filepath, token)
		if err == nil {
			return nil
		}
		if attempt < maxRetries {
			fmt.Printf("\n  Retry %d/%d in 3s... (%v)\n", attempt, maxRetries, err)
			time.Sleep(3 * time.Second)
		} else {
			return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
		}
	}
	return nil
}

func downloadAttempt(url string, filepath string, token string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	stat, _ := file.Stat()
	existingSize := stat.Size()

	// ساخت request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("User-Agent", GetUserAgent())
	if existingSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingSize))
	}

	// timeout بالاتر = مشکل سرعت حل میشه
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	// اگه سرور resume نداشت از اول شروع کن
	if existingSize > 0 && resp.StatusCode != 206 {
		file.Truncate(0)
		file.Seek(0, 0)
		existingSize = 0
	}

	if resp.StatusCode != 200 && resp.StatusCode != 206 {
		if resp.StatusCode == 403 {
			return fmt.Errorf("HTTP 403 (دسترسی رد شد — معمولاً لینک/کلید منقضی یا بدون امضا)")
		}
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	totalSize := existingSize + resp.ContentLength

	stats := &DownloadStats{
		TotalSize:     totalSize,
		Downloaded:    existingSize,
		StartTime:     time.Now(),
		LastSpeedTime: time.Now(),
		LastSpeedPos:  existingSize,
	}

	file.Seek(existingSize, 0)

	// buffer بزرگتر = سرعت بهتر
	buffer := make([]byte, 256*1024)
	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("write: %w", writeErr)
			}
			stats.Downloaded += int64(n)
			stats.UpdateSpeed()
			if totalSize > 500*1024 {
				stats.PrintProgress()
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("read: %w", readErr)
		}
	}

	if totalSize > 500*1024 {
		fmt.Println()
	}
	return nil
}
