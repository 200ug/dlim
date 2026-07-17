package internal

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
)

func TestParseTestURLs(t *testing.T) {
	f, err := os.Open("../test.txt")
	if err != nil {
		t.Fatalf("open test.txt: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		url := scanner.Text()
		if url == "" {
			continue
		}
		board, id, err := ParseThreadURL(url)
		if err != nil {
			t.Errorf("ParseThreadURL(%q): %v", url, err)
			continue
		}
		if board == "" || id == 0 {
			t.Errorf("ParseThreadURL(%q) returned empty board or id", url)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan test.txt: %v", err)
	}
	if count == 0 {
		t.Fatal("no URLs parsed from test.txt")
	}
}

func TestParseThreadURL(t *testing.T) {
	cases := []struct {
		url       string
		wantBoard string
		wantID    int64
	}{
		{
			url:       "https://boards.4chan.org/an/thread/5140385/ursidae-bears-general",
			wantBoard: "an",
			wantID:    5140385,
		},
		{
			url:       "https://boards.4chan.org/g/thread/109244018",
			wantBoard: "g",
			wantID:    109244018,
		},
		{
			url:       "https://boards.4chan.org/g/thread/109244018#bottom",
			wantBoard: "g",
			wantID:    109244018,
		},
		{
			url:       "https://boards.4chan.org/g/thread/109244018?sort=old",
			wantBoard: "g",
			wantID:    109244018,
		},
		{
			url:       "https://a.4cdn.org/g/thread/109244018.json",
			wantBoard: "g",
			wantID:    109244018,
		},
		{
			url:       "http://boards.4chan.org/g/thread/109244018",
			wantBoard: "g",
			wantID:    109244018,
		},
	}

	for _, c := range cases {
		board, id, err := ParseThreadURL(c.url)
		if err != nil {
			t.Errorf("ParseThreadURL(%q): unexpected error: %v", c.url, err)
			continue
		}
		if board != c.wantBoard || id != c.wantID {
			t.Errorf("ParseThreadURL(%q) = (%q, %d), want (%q, %d)", c.url, board, id, c.wantBoard, c.wantID)
		}
	}
}

func TestParseThreadURLInvalid(t *testing.T) {
	invalid := []string{
		"https://boards.4chan.org/g/",
		"not a url",
		"",
	}
	for _, url := range invalid {
		if _, _, err := ParseThreadURL(url); err == nil {
			t.Errorf("ParseThreadURL(%q) should have returned an error", url)
		}
	}
}

func TestAspectRatioJSON(t *testing.T) {
	var rc RunConfig
	if err := json.Unmarshal([]byte(`{"accepted_ratios":[[16,9],[4,3]]}`), &rc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(rc.AcceptedRatios) != 2 {
		t.Fatalf("expected 2 ratios, got %d", len(rc.AcceptedRatios))
	}
	if rc.AcceptedRatios[0].W != 16 || rc.AcceptedRatios[0].H != 9 {
		t.Errorf("expected 16:9, got %d:%d", rc.AcceptedRatios[0].W, rc.AcceptedRatios[0].H)
	}
	if rc.AcceptedRatios[1].W != 4 || rc.AcceptedRatios[1].H != 3 {
		t.Errorf("expected 4:3, got %d:%d", rc.AcceptedRatios[1].W, rc.AcceptedRatios[1].H)
	}

	data, err := json.Marshal(rc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != `{"default_output_dir":"","blacklist":{},"accepted_ratios":[[16,9],[4,3]]}` {
		t.Errorf("unexpected marshaled output: %s", string(data))
	}
}
