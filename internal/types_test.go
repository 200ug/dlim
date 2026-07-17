package internal

import (
	"encoding/json"
	"testing"
)

func TestThreadItemJSONNumericFields(t *testing.T) {
	data := []byte(`{
		"no": 5140385,
		"name": "Anonymous",
		"tim": 1781910543334700,
		"ext": ".png",
		"w": 1200,
		"h": 675,
		"md5": "ZuO/EIV/YPB3zRS74Q5KCQ==",
		"fsize": 998037
	}`)

	var post ThreadItem
	if err := json.Unmarshal(data, &post); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if post.ID != 5140385 {
		t.Errorf("expected ID 5140385, got %d", post.ID)
	}
	if post.ImageID != 1781910543334700 {
		t.Errorf("expected ImageID 1781910543334700, got %d", post.ImageID)
	}
}

func TestThreadJSONNumericID(t *testing.T) {
	data := []byte(`{
		"no": 5140385,
		"sub": "test thread",
		"com": "description",
		"posts": [
			{"no": 5140386, "name": "Anonymous", "tim": 123, "ext": ".jpg"}
		]
	}`)

	var thread Thread
	if err := json.Unmarshal(data, &thread); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if thread.ID != 5140385 {
		t.Errorf("expected thread ID 5140385, got %d", thread.ID)
	}
	if len(thread.Posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(thread.Posts))
	}
	if thread.Posts[0].ID != 5140386 {
		t.Errorf("expected post ID 5140386, got %d", thread.Posts[0].ID)
	}
}
