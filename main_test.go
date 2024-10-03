package main

import "testing"

func TestSync(t *testing.T) {
	bucket := "pratermade-gotest"
	p := "C:\\Users\\pratersm\\Videos\\Handbreak\\Sin City (Disc 1)" // Needs Splitting
	// p := "C:\\Users\\pratersm\\Videos\\Handbreak\\THE BIG BANG THEORY SEASON 1 DISC 3\\The Big Bang Theory S01E13.mkv"
	filters := []string{
		"mp4",
		"mkv",
	}
	err := sync(bucket, p, filters, false)
	if err != nil {
		t.Fatal(err)
	}

}
