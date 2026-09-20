package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (m *Manager) queuePath() string {
	return filepath.Join(m.basePath, "data", "queue.json")
}

func (m *Manager) loadQueue() {
	path := m.queuePath()
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return
	}
	var jobs []*Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		fmt.Printf("Warning: could not load queue: %v\n", err)
		return
	}
	for _, job := range jobs {
		if job == nil {
			continue
		}
		// دانلود نیمه‌کاره بعد از ری‌استارت باید متوقف بماند تا کاربر ادامه بدهد
		if job.Status == "running" || job.Status == "preparing" {
			job.Status = "paused"
			job.Message = "بعد از بستن برنامه متوقف شد"
			job.Error = "اپ بسته شد — برای ادامه، ادامه را بزنید"
		}
		job.stopMode = ""
		if job.Audio == nil {
			job.Audio = []string{}
		}
		if job.Subs == nil {
			job.Subs = []string{}
		}
	}
	m.jobs = jobs
}

func (m *Manager) saveQueueLocked() {
	path := m.queuePath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(m.jobs, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0644)
}

func (m *Manager) persistQueue() {
	m.mu.Lock()
	m.saveQueueLocked()
	m.mu.Unlock()
}
