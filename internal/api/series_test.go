package api

import (
	"encoding/json"
	"testing"
)

func TestListEpisodesFromWatch(t *testing.T) {
	raw := `{"data":{"attributes":{"pre_title":"سریال","movie_title":"نمونه: جلسه 1","seriesData":[{"title":"فصل اول","cover":"https://img/season1.jpg","episode":[{"title":"قسمت ۱","uid":"o4fza","image":"https://img/ep1-poster.jpg","thumbplay":"https://img/ep1-frame.jpg"},{"title":"قسمت ۲","uid":"q716x","image":"https://img/ep2-poster.jpg","thumbplay":"https://img/ep2-frame.jpg"}]}]}}}`
	var watch Watch
	if err := json.Unmarshal([]byte(raw), &watch); err != nil {
		t.Fatal(err)
	}
	if !IsSeriesWatch(watch) {
		t.Fatal("expected series")
	}
	if got := SeriesTitle(watch); got != "نمونه" {
		t.Fatalf("title %q", got)
	}
	eps := ListEpisodesFromWatch(watch)
	if len(eps) != 2 {
		t.Fatalf("episodes %d", len(eps))
	}
	if eps[0].ID != "o4fza" || eps[0].Number != 1 || eps[0].Season != 1 {
		t.Fatalf("first %#v", eps[0])
	}
	if eps[0].Cover != "https://img/ep1-frame.jpg" {
		t.Fatalf("ep1 should prefer thumbplay, got %q", eps[0].Cover)
	}
	if eps[0].Poster != "https://img/ep1-poster.jpg" {
		t.Fatalf("ep1 poster %q", eps[0].Poster)
	}
	if eps[1].Cover != "https://img/ep2-frame.jpg" {
		t.Fatalf("ep2 cover %q", eps[1].Cover)
	}
}
