package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"FilimoDownloader-GholamTaksir/internal/config"
	"FilimoDownloader-GholamTaksir/internal/helper"
	"FilimoDownloader-GholamTaksir/internal/history"
)

type Manager struct {
	mu       sync.Mutex
	basePath string
	cfg      *config.Settings
	hist     *history.History
	token    helper.AuthToken
	username string
	jobs     []*Job
	cancels  map[string]context.CancelFunc
	active   map[string]bool
	OnChange func()
	OnDone   func(job Job)

	wake      chan struct{}
	lastUI    time.Time
	pendingUI bool
}

func AppSupportDir() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "FilimoDownloader")
	case "windows":
		base := os.Getenv("APPDATA")
		if base == "" {
			base = home
		}
		return filepath.Join(base, "FilimoDownloader")
	default:
		return filepath.Join(home, ".local", "share", "FilimoDownloader")
	}
}

func DefaultDownloadDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Downloads", "Filimo")
}

func NewManager() *Manager {
	base := AppSupportDir()
	helper.MakeDirectories(helper.DataDirPath(base))
	cfg := config.Load(base)
	if cfg.DownloadPath == "" || cfg.DownloadPath == "Downloads" {
		cfg.DownloadPath = DefaultDownloadDir()
		config.Save(base, cfg)
	}
	hist := history.NewHistory(base)
	hist.Load()
	m := &Manager{
		basePath: base,
		cfg:      cfg,
		hist:     hist,
		token: helper.AuthToken{
			Path: helper.TokenPath(base),
		},
		cancels: map[string]context.CancelFunc{},
		active:  map[string]bool{},
		wake:    make(chan struct{}, 1),
	}
	m.loadQueue()
	go m.worker()
	return m
}

func (m *Manager) kick() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

func (m *Manager) emitUI() {
	if m.OnChange != nil {
		m.OnChange()
	}
}

// notify persists queue and pushes UI — use for status changes.
func (m *Manager) notify() {
	m.mu.Lock()
	m.pendingUI = false
	m.lastUI = time.Now()
	m.mu.Unlock()
	m.persistQueue()
	m.emitUI()
}

// notifyUI pushes progress to the UI without writing queue.json every tick.
func (m *Manager) notifyUI() {
	m.mu.Lock()
	now := time.Now()
	elapsed := now.Sub(m.lastUI)
	if elapsed < 250*time.Millisecond {
		if !m.pendingUI {
			m.pendingUI = true
			delay := 250*time.Millisecond - elapsed
			m.mu.Unlock()
			time.AfterFunc(delay, func() {
				m.mu.Lock()
				if !m.pendingUI {
					m.mu.Unlock()
					return
				}
				m.pendingUI = false
				m.lastUI = time.Now()
				m.mu.Unlock()
				m.emitUI()
			})
			return
		}
		m.mu.Unlock()
		return
	}
	m.lastUI = now
	m.pendingUI = false
	m.mu.Unlock()
	m.emitUI()
}

func (m *Manager) client() (helper.HttpClient, error) {
	tok := strings.TrimSpace(m.token.Get())
	if tok == "" {
		return helper.HttpClient{}, fmt.Errorf("اول توکن AuthV1 را وارد کنید")
	}
	return helper.HttpClient{Token: tok, UserAgent: helper.GetUserAgent()}, nil
}

func (m *Manager) sessionLocked() Session {
	return Session{
		LoggedIn: m.username != "",
		Username: m.username,
		HasToken: strings.TrimSpace(m.token.Get()) != "",
	}
}

func (m *Manager) Session() Session {
	m.mu.Lock()
	s := m.sessionLocked()
	needRestore := s.HasToken && !s.LoggedIn
	m.mu.Unlock()
	if needRestore {
		return m.RestoreSession()
	}
	return s
}

func (m *Manager) RestoreSession() Session {
	tok := strings.TrimSpace(m.token.Get())
	if tok == "" {
		m.mu.Lock()
		defer m.mu.Unlock()
		return m.sessionLocked()
	}
	client := helper.HttpClient{Token: tok, UserAgent: helper.GetUserAgent()}
	name, err := getUser(client)
	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil {
		// فقط وقتی فیلیمو توکن را رد کند پاک کن؛ خطای شبکه توکن ذخیره‌شده را نگه دار
		if strings.Contains(err.Error(), "توکن نامعتبر") || strings.Contains(err.Error(), "نامعتبر است") {
			m.token.Delete()
			m.username = ""
		}
		return m.sessionLocked()
	}
	m.username = name
	return m.sessionLocked()
}

func (m *Manager) Login(token string) (Session, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Session{}, fmt.Errorf("توکن خالی است")
	}
	client := helper.HttpClient{Token: token, UserAgent: helper.GetUserAgent()}
	name, err := getUser(client)
	if err != nil {
		return Session{}, err
	}
	m.token.Set(token)
	m.mu.Lock()
	m.username = name
	m.mu.Unlock()
	return m.Session(), nil
}

func (m *Manager) Logout() Session {
	m.token.Delete()
	m.mu.Lock()
	m.username = ""
	m.mu.Unlock()
	return m.Session()
}

func (m *Manager) Settings() Settings {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Settings{
		DefaultQuality:       m.cfg.DefaultQuality,
		DefaultFormat:        m.cfg.DefaultFormat,
		DownloadPath:         m.cfg.DownloadPath,
		AutoOpenFolder:       m.cfg.AutoOpenFolder,
		ShowInfoBeforeDL:     m.cfg.ShowInfoBeforeDL,
		ConcurrentDownloads:  m.cfg.MaxThreads,
		Theme:                m.cfg.Theme,
		Language:             m.cfg.Language,
		Notifications:        m.cfg.Notifications,
		CheckUpdatesOnLaunch: m.cfg.CheckUpdatesOnLaunch,
	}
}

func (m *Manager) SaveSettings(s Settings) Settings {
	m.mu.Lock()
	if strings.TrimSpace(s.DefaultQuality) != "" {
		m.cfg.DefaultQuality = s.DefaultQuality
	}
	if s.DefaultFormat == "mkv" || s.DefaultFormat == "mp4" {
		m.cfg.DefaultFormat = s.DefaultFormat
	}
	if strings.TrimSpace(s.DownloadPath) != "" {
		m.cfg.DownloadPath = strings.TrimSpace(s.DownloadPath)
	}
	m.cfg.AutoOpenFolder = s.AutoOpenFolder
	m.cfg.ShowInfoBeforeDL = s.ShowInfoBeforeDL
	if s.ConcurrentDownloads > 0 {
		m.cfg.MaxThreads = s.ConcurrentDownloads
	}
	if s.Theme != "" {
		m.cfg.Theme = s.Theme
	}
	if s.Language != "" {
		m.cfg.Language = s.Language
	}
	m.cfg.Notifications = s.Notifications
	m.cfg.CheckUpdatesOnLaunch = s.CheckUpdatesOnLaunch
	config.Save(m.basePath, m.cfg)
	m.mu.Unlock()
	return m.Settings()
}

func (m *Manager) Jobs() []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		out = append(out, *job)
	}
	return out
}

func (m *Manager) History() []HistoryItem {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]HistoryItem, 0, len(m.hist.Downloads))
	for _, rec := range m.hist.Downloads {
		out = append(out, HistoryItem{
			ID:           rec.ID,
			Title:        rec.Title,
			Quality:      rec.Quality,
			Format:       rec.Format,
			SizeMB:       rec.SizeMB,
			DownloadedAt: rec.DownloadedAt.Format("2006-01-02 15:04"),
			FilePath:     rec.FilePath,
		})
	}
	return out
}

func (m *Manager) Enqueue(req EnqueueRequest) ([]Job, error) {
	if _, err := m.client(); err != nil {
		return nil, err
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("چیزی برای صف انتخاب نشده")
	}
	if strings.TrimSpace(req.Quality) == "" {
		req.Quality = m.Settings().DefaultQuality
	}
	if req.Format != "mkv" {
		req.Format = "mp4"
	}
	m.mu.Lock()
	for i, item := range req.Items {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = id
		}
		kind := item.Kind
		if kind == "" {
			kind = "movie"
		}
		seriesKey := strings.TrimSpace(item.SeriesKey)
		seriesTitle := strings.TrimSpace(item.SeriesTitle)
		if kind == "series" {
			if seriesKey == "" {
				seriesKey = seriesTitle
			}
			if seriesKey == "" {
				seriesKey = id
			}
			if seriesTitle == "" {
				seriesTitle = title
			}
		}
		job := &Job{
			ID:          fmt.Sprintf("%d-%d-%s", time.Now().UnixNano(), i, id),
			ContentID:   id,
			Title:       title,
			Quality:     req.Quality,
			Format:      req.Format,
			Status:      "queued",
			Message:     "در صف",
			Kind:        kind,
			SeriesKey:   seriesKey,
			SeriesTitle: seriesTitle,
			Season:      item.Season,
			Episode:     item.Episode,
			SeasonTitle: item.SeasonTitle,
			Cover:       item.Cover,
			Audio:       append([]string{}, req.Audio...),
			Subs:        append([]string{}, req.Subtitles...),
		}
		m.jobs = append(m.jobs, job)
	}
	m.mu.Unlock()
	m.notify()
	m.kick()
	return m.Jobs(), nil
}

func (m *Manager) Cancel(id string) error {
	m.mu.Lock()
	for _, job := range m.jobs {
		if job.ID != id {
			continue
		}
		switch job.Status {
		case "queued":
			job.Status = "canceled"
			job.Message = "لغو شد"
			job.Error = "توسط شما از صف حذف شد"
			job.stopMode = ""
		case "paused":
			job.Status = "canceled"
			job.Message = "لغو شد"
			job.Error = "توسط شما لغو شد"
			job.stopMode = ""
		case "preparing", "running":
			job.stopMode = "cancel"
			if cancel, ok := m.cancels[id]; ok {
				cancel()
				job.Message = "در حال لغو…"
			} else {
				job.Status = "canceled"
				job.Message = "لغو شد"
				job.Error = "توسط شما از صف حذف شد"
				job.stopMode = ""
			}
		}
	}
	m.mu.Unlock()
	m.notify()
	return nil
}

func (m *Manager) Pause(id string) error {
	m.mu.Lock()
	found := false
	for _, job := range m.jobs {
		if job.ID != id {
			continue
		}
		found = true
		switch job.Status {
		case "queued":
			job.Status = "paused"
			job.Message = "متوقف شد"
			job.Error = "توسط شما متوقف شد — با ادامه، دوباره در صف قرار می‌گیرد"
			job.stopMode = ""
		case "preparing", "running":
			job.stopMode = "pause"
			job.Message = "در حال توقف…"
			if cancel, ok := m.cancels[id]; ok {
				cancel()
			} else if job.Status == "preparing" {
				job.Status = "paused"
				job.Message = "متوقف شد"
				job.Error = "توسط شما متوقف شد — با ادامه، دوباره در صف قرار می‌گیرد"
				job.stopMode = ""
			}
		default:
			m.mu.Unlock()
			return fmt.Errorf("این دانلود را نمی‌شود متوقف کرد")
		}
	}
	m.mu.Unlock()
	if !found {
		return fmt.Errorf("دانلود پیدا نشد")
	}
	m.notify()
	return nil
}

func (m *Manager) Resume(id string) error {
	return m.requeue(id, false)
}

func (m *Manager) Retry(id string) error {
	return m.requeue(id, true)
}

func (m *Manager) requeue(id string, fresh bool) error {
	m.mu.Lock()
	found := false
	for _, job := range m.jobs {
		if job.ID != id {
			continue
		}
		found = true
		switch job.Status {
		case "paused", "canceled", "error", "done":
			job.Error = ""
			job.stopMode = ""
			job.FilePath = ""
			if fresh {
				job.Status = "queued"
				job.Progress = 0
				job.Message = "در صف (دانلود مجدد)"
			} else {
				// Play → show preparing immediately until real download starts
				job.Status = "preparing"
				job.Message = "آماده‌سازی"
			}
		case "queued", "preparing", "running":
			m.mu.Unlock()
			return fmt.Errorf("این دانلود همین حالا فعال است")
		default:
			m.mu.Unlock()
			return fmt.Errorf("وضعیت این دانلود برای ادامه مناسب نیست")
		}
	}
	m.mu.Unlock()
	if !found {
		return fmt.Errorf("دانلود پیدا نشد")
	}
	m.notify()
	m.kick()
	return nil
}

func (m *Manager) OpenPath(p string) error {
	if strings.TrimSpace(p) == "" {
		return fmt.Errorf("مسیر خالی است")
	}
	info, err := os.Stat(p)
	if err != nil {
		return err
	}
	target := p
	if !info.IsDir() {
		target = filepath.Dir(p)
	}
	helper.OpenFolder(target)
	return nil
}

func (m *Manager) RevealPath(p string) error {
	if strings.TrimSpace(p) == "" {
		return fmt.Errorf("مسیر خالی است")
	}
	if _, err := os.Stat(p); err != nil {
		return err
	}
	helper.RevealInFinder(p)
	return nil
}

func (m *Manager) PreviewPath(p string) error {
	if strings.TrimSpace(p) == "" {
		return fmt.Errorf("مسیر خالی است")
	}
	info, err := os.Stat(p)
	if err != nil {
		return err
	}
	if info.IsDir() {
		helper.OpenFolder(p)
		return nil
	}
	helper.PreviewFile(p)
	return nil
}

func (m *Manager) OpenSeriesFolder(seriesKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, job := range m.jobs {
		key := job.SeriesKey
		if key == "" {
			key = "movie:" + job.ContentID
		}
		if key != seriesKey {
			continue
		}
		if strings.TrimSpace(job.FilePath) == "" {
			continue
		}
		helper.OpenFolder(filepath.Dir(job.FilePath))
		return nil
	}
	root := strings.TrimSpace(m.cfg.DownloadPath)
	if root == "" {
		return fmt.Errorf("هنوز فایلی برای این مورد ذخیره نشده")
	}
	helper.OpenFolder(root)
	return nil
}

func (m *Manager) RemoveJob(id string) error {
	m.mu.Lock()
	idx := -1
	for i, job := range m.jobs {
		if job.ID != id {
			continue
		}
		idx = i
		if job.Status == "running" {
			job.stopMode = "cancel"
			if cancel, ok := m.cancels[id]; ok {
				cancel()
			}
		}
		break
	}
	if idx < 0 {
		m.mu.Unlock()
		return fmt.Errorf("آیتم پیدا نشد")
	}
	m.jobs = append(m.jobs[:idx], m.jobs[idx+1:]...)
	m.mu.Unlock()
	m.notify()
	return nil
}

func (m *Manager) RemoveSeries(seriesKey string) error {
	m.mu.Lock()
	kept := make([]*Job, 0, len(m.jobs))
	removed := 0
	for _, job := range m.jobs {
		key := job.SeriesKey
		if key == "" {
			key = "movie:" + job.ContentID
		}
		if key == seriesKey {
			if job.Status == "running" {
				job.stopMode = "cancel"
				if cancel, ok := m.cancels[job.ID]; ok {
					cancel()
				}
			}
			removed++
			continue
		}
		kept = append(kept, job)
	}
	m.jobs = kept
	m.mu.Unlock()
	if removed == 0 {
		return fmt.Errorf("موردی برای حذف نبود")
	}
	m.notify()
	return nil
}

func (m *Manager) worker() {
	for {
		job := m.claimNextQueued()
		if job == nil {
			select {
			case <-m.wake:
			case <-time.After(350 * time.Millisecond):
			}
			continue
		}
		go m.runJob(job)
	}
}

func (m *Manager) claimNextQueued() *Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	limit := m.cfg.MaxThreads
	if limit < 1 {
		limit = 1
	}
	if limit > 5 {
		limit = 5
	}
	running := 0
	for _, job := range m.jobs {
		if m.active[job.ID] || job.Status == "running" {
			running++
		}
	}
	if running >= limit {
		return nil
	}
	for _, job := range m.jobs {
		if m.active[job.ID] {
			continue
		}
		if job.Status == "queued" || job.Status == "preparing" {
			m.active[job.ID] = true
			job.Status = "preparing"
			job.Message = "آماده‌سازی"
			job.Error = ""
			return job
		}
	}
	return nil
}

func (m *Manager) updateJob(job *Job, fn func()) {
	m.mu.Lock()
	fn()
	m.mu.Unlock()
	m.notifyUI()
}

func (m *Manager) LibrarySeries() []LibrarySeries {
	m.mu.Lock()
	defer m.mu.Unlock()
	type agg struct {
		item     LibrarySeries
		seasons  map[int]bool
		progress float64
		n        int
	}
	groups := map[string]*agg{}
	order := []string{}
	for _, job := range m.jobs {
		key := job.SeriesKey
		title := job.SeriesTitle
		kind := job.Kind
		if key == "" || kind != "series" {
			key = "movie:" + job.ContentID
			title = job.Title
			kind = "movie"
		}
		g, ok := groups[key]
		if !ok {
			g = &agg{item: LibrarySeries{Key: key, Title: title, Cover: job.Cover, Kind: kind}, seasons: map[int]bool{}}
			groups[key] = g
			order = append(order, key)
		}
		if g.item.Cover == "" && job.Cover != "" {
			g.item.Cover = job.Cover
		}
		if job.Season > 0 {
			g.seasons[job.Season] = true
		}
		g.item.EpisodeCount++
		if job.Status == "done" {
			g.item.DoneCount++
		}
		if job.Status == "queued" || job.Status == "preparing" || job.Status == "running" || job.Status == "paused" {
			g.item.ActiveCount++
		}
		g.progress += job.Progress
		g.n++
	}
	out := make([]LibrarySeries, 0, len(order))
	for _, key := range order {
		g := groups[key]
		g.item.SeasonCount = len(g.seasons)
		if g.n > 0 {
			g.item.Progress = g.progress / float64(g.n)
		}
		out = append(out, g.item)
	}
	return out
}

func (m *Manager) LibrarySeasons(seriesKey string) []LibrarySeason {
	m.mu.Lock()
	defer m.mu.Unlock()
	type agg struct {
		item     LibrarySeason
		progress float64
		n        int
	}
	groups := map[int]*agg{}
	order := []int{}
	for _, job := range m.jobs {
		key := job.SeriesKey
		if key == "" {
			key = "movie:" + job.ContentID
		}
		if key != seriesKey {
			continue
		}
		season := job.Season
		g, ok := groups[season]
		if !ok {
			title := job.SeasonTitle
			if title == "" {
				if season > 0 {
					title = fmt.Sprintf("فصل %d", season)
				} else {
					title = "قسمت‌ها"
				}
			}
			g = &agg{item: LibrarySeason{SeriesKey: seriesKey, Season: season, Title: title, Cover: job.Cover}}
			groups[season] = g
			order = append(order, season)
		}
		if g.item.Cover == "" && job.Cover != "" {
			g.item.Cover = job.Cover
		}
		g.item.EpisodeCount++
		if job.Status == "done" {
			g.item.DoneCount++
		}
		g.progress += job.Progress
		g.n++
	}
	out := make([]LibrarySeason, 0, len(order))
	for _, season := range order {
		g := groups[season]
		if g.n > 0 {
			g.item.Progress = g.progress / float64(g.n)
		}
		out = append(out, g.item)
	}
	return out
}

func (m *Manager) LibraryEpisodes(seriesKey string, season int) []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Job, 0)
	for _, job := range m.jobs {
		key := job.SeriesKey
		if key == "" {
			key = "movie:" + job.ContentID
		}
		if key != seriesKey {
			continue
		}
		if job.Kind == "series" && job.Season != season {
			continue
		}
		out = append(out, *job)
	}
	return out
}
