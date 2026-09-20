package stream

import (
	"os"
	"testing"
)

func TestLookFFmpegIgnoresEmptyPATH(t *testing.T) {
	old := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", old) })
	os.Setenv("PATH", "/usr/bin:/bin")
	p, err := lookFFmpeg()
	if err != nil {
		t.Fatal(err)
	}
	if p == "" {
		t.Fatal("empty path")
	}
	t.Log("ffmpeg=", p)
}
