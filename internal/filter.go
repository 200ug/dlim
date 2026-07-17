package internal

import (
	"strings"
)

// returns true if OP's title or description contains a blacklisted keyword
func (b Blacklist) MatchesThread(t Thread) bool {
	text := strings.ToLower(t.OPTitle + " " + t.OPDescription)
	for _, kw := range b.Keywords {
		if kw != "" && strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// returns true if the media is authored by blacklisted username or tripcode
func (b Blacklist) MatchesPost(post ThreadItem) bool {
	name := strings.ToLower(post.OPName)
	trip := strings.ToLower(post.OPTripcode)
	for _, u := range b.Usernames {
		if u != "" && name == strings.ToLower(u) {
			return true
		}
	}
	for _, t := range b.Tripcodes {
		if t != "" && trip == strings.ToLower(t) {
			return true
		}
	}
	return false
}

// returns true if the post's media passes all configured filters (expects
// post.ImageID to be non-empty)
func (rc RunConfig) AcceptsMedia(post ThreadItem) bool {
	if rc.MaxFilesize > 0 && post.ImageSize > rc.MaxFilesize {
		return false
	}
	if rc.OnlyImages && isVideo(post.ImageExt) {
		return false
	}
	if len(rc.AcceptedExtensions) > 0 && !matchesAnyExtension(post.ImageExt, rc.AcceptedExtensions) {
		return false
	}
	if len(rc.AcceptedRatios) > 0 && !matchesAnyRatio(post.ImageW, post.ImageH, rc.AcceptedRatios) {
		return false
	}
	return true
}

func isVideo(ext string) bool {
	switch strings.ToLower(ext) {
	case ".webm", ".mp4":
		return true
	}
	return false
}

func matchesAnyExtension(ext string, accepted []string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	for _, a := range accepted {
		if ext == strings.ToLower(a) {
			return true
		}
	}
	return false
}

func matchesAnyRatio(w, h int, accepted []AspectRatio) bool {
	if w == 0 || h == 0 {
		return false
	}
	rw, rh := aspectRatio(w, h)
	for _, a := range accepted {
		if rw == a.W && rh == a.H {
			return true
		}
	}
	return false
}
