package internal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// extract board abbreviation and thread ID (supports both legacy and modern slug formats)
func ParseThreadURL(url string) (board string, threadID int64, err error) {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, ".json")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	parts := strings.Split(url, "/")
	for i, p := range parts {
		if p == "thread" && i+1 < len(parts) && i > 0 {
			board = parts[i-1]
			idStr := parts[i+1]
			idStr = strings.Split(idStr, "?")[0]
			idStr = strings.Split(idStr, "#")[0]
			if board == "" || idStr == "" {
				return "", 0, fmt.Errorf("invalid thread URL")
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return "", 0, fmt.Errorf("invalid thread URL: %s", url)
			}
			return board, id, nil
		}
	}
	return "", 0, fmt.Errorf("invalid thread URL: %s", url)
}

type AspectRatio struct {
	W int
	H int
}

func (a AspectRatio) MarshalJSON() ([]byte, error) {
	return json.Marshal([]int{a.W, a.H})
}

func (a *AspectRatio) UnmarshalJSON(data []byte) error {
	var v []int
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if len(v) != 2 {
		return fmt.Errorf("aspect ratio must have exactly 2 values")
	}
	a.W, a.H = v[0], v[1]
	return nil
}

func aspectRatio(w, h int) (int, int) {
	g := gcd(w, h)
	return w / g, h / g
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}
