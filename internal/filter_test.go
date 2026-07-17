package internal

import (
	"testing"
)

func TestBlacklistMatchesThread(t *testing.T) {
	b := Blacklist{Keywords: []string{"bad", "WORST"}}

	if !b.MatchesThread(Thread{OPTitle: "this is a bad thread"}) {
		t.Error("expected match on title keyword")
	}
	if !b.MatchesThread(Thread{OPDescription: "the WORST thing"}) {
		t.Error("expected match on description keyword")
	}
	if b.MatchesThread(Thread{OPTitle: "completely fine", OPDescription: "nothing wrong"}) {
		t.Error("expected no match")
	}
}

func TestBlacklistMatchesPost(t *testing.T) {
	b := Blacklist{
		Usernames: []string{"Troll"},
		Tripcodes: []string{"!trip123"},
	}

	if !b.MatchesPost(ThreadItem{OPName: "Troll", ImageID: 1}) {
		t.Error("expected match on username")
	}
	if !b.MatchesPost(ThreadItem{OPTripcode: "!trip123", ImageID: 1}) {
		t.Error("expected match on tripcode")
	}
	if b.MatchesPost(ThreadItem{OPName: "Anonymous", OPTripcode: "!other", ImageID: 1}) {
		t.Error("expected no match")
	}
}

func TestAcceptsMediaMaxFilesize(t *testing.T) {
	rc := RunConfig{MaxFilesize: 1000}
	if !rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageSize: 500}) {
		t.Error("expected media under limit to be accepted")
	}
	if rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageSize: 1500}) {
		t.Error("expected media over limit to be rejected")
	}
}

func TestAcceptsMediaOnlyImages(t *testing.T) {
	rc := RunConfig{OnlyImages: true}
	if !rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageExt: ".png"}) {
		t.Error("expected png to be accepted")
	}
	if rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageExt: ".webm"}) {
		t.Error("expected webm to be rejected")
	}
	if rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageExt: ".mp4"}) {
		t.Error("expected mp4 to be rejected")
	}
}

func TestAcceptsMediaExtensions(t *testing.T) {
	rc := RunConfig{AcceptedExtensions: []string{"png", "jpg"}}
	if !rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageExt: ".png"}) {
		t.Error("expected .png to match")
	}
	if !rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageExt: ".jpg"}) {
		t.Error("expected .jpg to match")
	}
	if rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageExt: ".webm"}) {
		t.Error("expected .webm to be rejected")
	}
}

func TestAcceptsMediaRatios(t *testing.T) {
	rc := RunConfig{AcceptedRatios: []AspectRatio{{16, 9}, {4, 3}}}
	if !rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageW: 1920, ImageH: 1080}) {
		t.Error("expected 16:9 to match")
	}
	if !rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageW: 800, ImageH: 600}) {
		t.Error("expected 4:3 to match")
	}
	if rc.AcceptsMedia(ThreadItem{ImageID: 1, ImageW: 1000, ImageH: 1000}) {
		t.Error("expected 1:1 to be rejected")
	}
}
