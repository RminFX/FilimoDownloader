package main

import (
	"context"

	"FilimoDownloader-GholamTaksir/engine"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
	mgr *engine.Manager
}

func NewApp() *App {
	return &App{
		mgr: engine.NewManager(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.mgr.OnChange = func() {
		runtime.EventsEmit(a.ctx, "queue:update", a.mgr.Jobs())
		runtime.EventsEmit(a.ctx, "history:update", a.mgr.History())
	}
	a.mgr.OnDone = func(job engine.Job) {
		runtime.EventsEmit(a.ctx, "job:done", job)
	}
	a.mgr.RestoreSession()
}

func (a *App) Session() engine.Session {
	return a.mgr.RestoreSession()
}

func (a *App) Login(token string) (engine.Session, error) {
	return a.mgr.Login(token)
}

func (a *App) Logout() engine.Session {
	return a.mgr.Logout()
}

func (a *App) Inspect(query string) (*engine.InspectResult, error) {
	return a.mgr.Inspect(query)
}

func (a *App) Enqueue(req engine.EnqueueRequest) ([]engine.Job, error) {
	return a.mgr.Enqueue(req)
}

func (a *App) Jobs() []engine.Job {
	return a.mgr.Jobs()
}

func (a *App) History() []engine.HistoryItem {
	return a.mgr.History()
}

func (a *App) Cancel(id string) error {
	return a.mgr.Cancel(id)
}

func (a *App) Pause(id string) error {
	return a.mgr.Pause(id)
}

func (a *App) Resume(id string) error {
	return a.mgr.Resume(id)
}

func (a *App) Retry(id string) error {
	return a.mgr.Retry(id)
}

func (a *App) RemoveJob(id string) error {
	return a.mgr.RemoveJob(id)
}

func (a *App) RemoveSeries(seriesKey string) error {
	return a.mgr.RemoveSeries(seriesKey)
}

func (a *App) OpenPath(p string) error {
	return a.mgr.OpenPath(p)
}

func (a *App) RevealPath(p string) error {
	return a.mgr.RevealPath(p)
}

func (a *App) PreviewPath(p string) error {
	return a.mgr.PreviewPath(p)
}

func (a *App) OpenSeriesFolder(seriesKey string) error {
	return a.mgr.OpenSeriesFolder(seriesKey)
}

func (a *App) Settings() engine.Settings {
	return a.mgr.Settings()
}

func (a *App) SaveSettings(s engine.Settings) engine.Settings {
	return a.mgr.SaveSettings(s)
}

func (a *App) LibrarySeries() []engine.LibrarySeries {
	return a.mgr.LibrarySeries()
}

func (a *App) LibrarySeasons(seriesKey string) []engine.LibrarySeason {
	return a.mgr.LibrarySeasons(seriesKey)
}

func (a *App) LibraryEpisodes(seriesKey string, season int) []engine.Job {
	return a.mgr.LibraryEpisodes(seriesKey, season)
}

func (a *App) PickFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "پوشه ذخیره فایل‌ها",
	})
}

func (a *App) AppVersion() string {
	return engine.Version()
}

func (a *App) CheckForUpdates() engine.UpdateInfo {
	return engine.CheckForUpdates()
}

func (a *App) OpenURL(url string) {
	if url == "" {
		return
	}
	runtime.BrowserOpenURL(a.ctx, url)
}
