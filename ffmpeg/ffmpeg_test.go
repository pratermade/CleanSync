package ffmpeg

import (
	"fmt"
	"testing"
)

func TestGetChapterInfo(t *testing.T) {
	video, err := NewVideo("C:\\artifacts\\original.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if video.Chapters[1].end != 20.967 {
		t.Fatalf("Expected the second chapter to end on 20967 but it ended on %f", video.Chapters[1].end)
	}
}

func TestReassembleChapters(t *testing.T) {
	video, err := NewVideo("C:\\artifacts\\original.mp4")
	if err != nil {
		t.Fatal(err)
	}
	_, err = video.Recut([]int{2})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetNonAdIndexes(t *testing.T) {
	video, err := NewVideo("C:\\artifacts\\original.mp4")
	if err != nil {
		t.Fatal(err)
	}
	nonads := video.GetNonAdIndexes(false)
	fmt.Println(nonads)
}
