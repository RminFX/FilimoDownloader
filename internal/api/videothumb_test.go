package api

import (
	"strings"
	"testing"
	"time"
)

func TestPickVTTCue(t *testing.T) {
	vtt := "WEBVTT\n\n00:00.000 --> 00:10.000\n196217-thumb-t01.jpg#xywh=0,0,106,60\n\n00:20.000 --> 00:30.000\n196217-thumb-t01.jpg#xywh=212,0,106,60\n"
	file, x, y, w, h, ok := pickVTTCue(vtt, 0)
	if !ok || file != "196217-thumb-t01.jpg" || x != 0 || y != 0 || w != 106 || h != 60 {
		t.Fatalf("first cue file=%s x=%d y=%d w=%d h=%d ok=%v", file, x, y, w, h, ok)
	}
	file, x, _, _, _, ok = pickVTTCue(vtt, 25*time.Second)
	if !ok || x != 212 {
		t.Fatalf("25s cue file=%s x=%d ok=%v", file, x, ok)
	}
}

func TestResolveSiblingURL(t *testing.T) {
	got := resolveSiblingURL("https://static.cdn.asset.filimo.com//filimo-video/196217-thumb-t.vtt", "196217-thumb-t01.jpg")
	if !strings.HasSuffix(got, "/filimo-video/196217-thumb-t01.jpg") {
		t.Fatalf("got %s", got)
	}
}
