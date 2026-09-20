package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"FilimoDownloader-GholamTaksir/internal/api"
	"FilimoDownloader-GholamTaksir/internal/helper"
	"FilimoDownloader-GholamTaksir/internal/history"
	"FilimoDownloader-GholamTaksir/internal/stream"
)

func (m *Manager) runJob(job *Job) {
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	if job.Status == "canceled" || job.Status == "paused" {
		delete(m.active, job.ID)
		m.mu.Unlock()
		return
	}
	job.Status = "preparing"
	job.Message = "آماده‌سازی"
	job.Error = ""
	job.stopMode = ""
	if job.Progress < 1 {
		job.Progress = 1
	}
	m.active[job.ID] = true
	m.cancels[job.ID] = cancel
	m.mu.Unlock()
	m.notify()

	defer func() {
		m.mu.Lock()
		delete(m.active, job.ID)
		m.mu.Unlock()
	}()

	err := m.downloadJob(ctx, job)
	cancel()

	doneCopy := Job{}
	finishedOK := false
	m.mu.Lock()
	delete(m.cancels, job.ID)
	stop := job.stopMode
	job.stopMode = ""
	if job.Status == "done" {
		job.Error = ""
		doneCopy = *job
		finishedOK = true
		m.mu.Unlock()
		m.notify()
		if m.OnDone != nil {
			m.OnDone(doneCopy)
		}
		return
	}
	switch {
	case stop == "pause" || job.Status == "paused":
		job.Status = "paused"
		job.Message = "متوقف شد"
		job.Error = "توسط شما متوقف شد — قطعه‌های دانلودشده نگه داشته می‌شوند"
	case stop == "cancel" || job.Status == "canceled":
		job.Status = "canceled"
		job.Message = "لغو شد"
		job.Error = "توسط شما لغو شد"
	case err != nil:
		if err == context.Canceled || strings.Contains(err.Error(), "context canceled") {
			job.Status = "canceled"
			job.Message = "لغو شد"
			job.Error = "دانلود قطع شد"
		} else {
			job.Status = "error"
			job.Error = err.Error()
			job.Message = "خطا"
		}
	}
	m.mu.Unlock()
	m.notify()
	_ = finishedOK
}

func (m *Manager) downloadJob(ctx context.Context, job *Job) error {
	client, err := m.client()
	if err != nil {
		return err
	}
	watch, err := api.GetEpisodeWatch(client, job.ContentID)
	if err != nil {
		return err
	}
	meta := api.FetchContentMeta(client, job.ContentID, watch)
	title := strings.TrimSpace(meta.Title)
	if title == "" {
		title = strings.TrimSpace(watch.Data.Attributes.Name)
	}
	if title == "" {
		title = job.Title
	}
	m.updateJob(job, func() {
		job.Title = title
		job.Kind = "movie"
		if meta.IsSeries {
			job.Kind = "series"
		}
		if meta.SeriesTitle != "" {
			job.SeriesTitle = meta.SeriesTitle
		}
		if job.SeriesKey == "" {
			if meta.ParentID != "" {
				job.SeriesKey = meta.ParentID
			} else if meta.SeriesTitle != "" {
				job.SeriesKey = meta.SeriesTitle
			}
		}
		if meta.Season > 0 {
			job.Season = meta.Season
		}
		if meta.Episode > 0 {
			job.Episode = meta.Episode
		}
		if meta.SeasonText != "" {
			job.SeasonTitle = meta.SeasonText
		}
		coverURL := firstNonEmpty(meta.PosterURL, meta.CoverURL, meta.ThumbURL)
		// Keep episode thumbplay from enqueue when movie/one only has shared series poster.
		if coverURL != "" && job.Cover == "" {
			job.Cover = coverURL
		}
		job.Message = "خواندن کیفیت‌ها"
		job.Progress = 4
	})

	hls, err := stream.GetHlsErr(client, watch)
	if err != nil {
		return err
	}

	variant, err := pickVariant(hls.Variants, job.Quality)
	if err != nil {
		return err
	}
	audios := pickAudio(hls.Tracks, job.Audio)
	subs := pickSubs(watch.Data.Attributes.Subtitles, job.Subs)

	root := m.Settings().DownloadPath
	helper.MakeDirectories(root)
	downloadDir := stream.DownloadDirIn(root, title)

	m.updateJob(job, func() {
		job.Message = "ذخیره پوستر و اطلاعات"
		job.Progress = 6
	})
	coverURL := firstNonEmpty(meta.PosterURL, meta.CoverURL, meta.ThumbURL)
	coverPath := filepath.Join(downloadDir, "cover.jpg")
	if helper.IsFileExists(coverPath) {
		// resume: don't re-download poster
	} else if coverURL != "" {
		if p, err := stream.DownloadCover(coverURL, downloadDir); err == nil {
			coverPath = p
		} else {
			coverPath = ""
		}
	} else {
		coverPath = ""
	}
	infoPath := filepath.Join(downloadDir, "info.json")
	if !helper.IsFileExists(infoPath) {
		_ = stream.SaveSidecarMeta(downloadDir, meta, coverPath)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	job.Status = "running"
	job.Message = "دانلود ویدیو"
	job.Progress = 8
	m.mu.Unlock()
	m.notify()

	videoDir := filepath.Join(stream.VideoDir(downloadDir), variant.Quality)
	err = stream.DownloadSegments(ctx, client, variant.Link, videoDir, func(done, total int) {
		pct := 8 + (float64(done)/float64(max(total, 1)))*62
		m.updateJob(job, func() {
			job.Status = "running"
			job.Progress = pct
			job.Message = fmt.Sprintf("ویدیو %d از %d", done, total)
		})
	})
	if err != nil {
		return err
	}

	if len(audios) > 0 {
		m.updateJob(job, func() {
			job.Message = "دانلود صدا"
			job.Progress = 72
		})
		for i, track := range audios {
			if err := ctx.Err(); err != nil {
				return err
			}
			audioDir := filepath.Join(stream.AudioDir(downloadDir), track.Language)
			err = stream.DownloadSegments(ctx, client, track.Link, audioDir, func(done, total int) {
				base := 72 + float64(i)/float64(len(audios))*10
				pct := base + (float64(done)/float64(max(total, 1)))*(10/float64(len(audios)))
				m.updateJob(job, func() {
					job.Progress = pct
					job.Message = fmt.Sprintf("صدا %s  %d از %d", track.Language, done, total)
				})
			})
			if err != nil {
				return err
			}
		}
	}

	if len(subs) > 0 {
		m.updateJob(job, func() {
			job.Message = "دانلود زیرنویس"
			job.Progress = 84
		})
		for _, sub := range subs {
			if err := ctx.Err(); err != nil {
				return err
			}
			stream.DownloadSubtitle(client, sub, downloadDir)
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	m.updateJob(job, func() {
		job.Message = "ساخت فایل نهایی با FFmpeg"
		job.Progress = 90
	})
	ext := ".mp4"
	if job.Format == "mkv" {
		ext = ".mkv"
	}
	out, err := stream.BuildAuto(downloadDir, "", ext)
	if err != nil {
		return fmt.Errorf("ساخت فایل در ۹۰٪ ماند: %w", err)
	}
	m.updateJob(job, func() {
		job.Progress = 95
		job.Message = "نوشتن متادیتا و پوستر"
	})
	if tagged, err := stream.ApplyFileMetadata(out, meta, coverPath); err == nil {
		out = tagged
	} else {
		fmt.Printf("  metadata warning: %v\n", err)
	}
	_ = stream.SaveSidecarMeta(downloadDir, meta, coverPath)
	m.updateJob(job, func() {
		job.Progress = 98
		job.Message = "در حال نهایی‌سازی"
	})

	sizeMB := 0.0
	if info, statErr := os.Stat(out); statErr == nil {
		sizeMB = float64(info.Size()) / 1024 / 1024
	}

	m.mu.Lock()
	job.Status = "done"
	job.Progress = 100
	job.Message = "تمام شد"
	job.Error = ""
	job.FilePath = out
	m.hist.Add(history.DownloadRecord{
		ID:           job.ContentID,
		Title:        title,
		Quality:      variant.Quality,
		Format:       strings.TrimPrefix(ext, "."),
		SizeMB:       sizeMB,
		DownloadedAt: time.Now(),
		FilePath:     out,
	})
	openFolder := m.cfg.AutoOpenFolder
	notifyUser := m.cfg.Notifications
	m.mu.Unlock()

	if openFolder {
		helper.OpenFolder(filepath.Dir(out))
	}
	if notifyUser {
		helper.Notify("دانلود تمام شد", title)
	}
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func pickVariant(variants []stream.HlsVideoVariant, quality string) (stream.HlsVideoVariant, error) {
	if len(variants) == 0 {
		return stream.HlsVideoVariant{}, fmt.Errorf("کیفیتی برای دانلود نیست")
	}
	quality = strings.TrimSpace(quality)
	for _, v := range variants {
		if v.Quality == quality {
			return v, nil
		}
	}
	return variants[len(variants)-1], nil
}

func pickAudio(tracks []stream.HlsAudioTrack, wanted []string) []stream.HlsAudioTrack {
	if len(wanted) == 0 {
		return tracks
	}
	set := map[string]bool{}
	for _, w := range wanted {
		set[w] = true
	}
	var out []stream.HlsAudioTrack
	for _, t := range tracks {
		if set[t.Language] {
			out = append(out, t)
		}
	}
	return out
}

func pickSubs(subs []api.WatchSubtitle, wanted []string) []api.WatchSubtitle {
	if len(wanted) == 0 {
		return subs
	}
	set := map[string]bool{}
	for _, w := range wanted {
		set[w] = true
	}
	var out []api.WatchSubtitle
	for _, s := range subs {
		if set[s.Language] {
			out = append(out, s)
		}
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
