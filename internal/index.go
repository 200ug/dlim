package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// metadata of a single downloaded file
type IndexEntry struct {
	MD5      string `json:"md5"`
	Filename string `json:"filename"`
	Board    string `json:"board"`
	Size     int    `json:"size"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	URL      string `json:"url"`
}

// gloabl, md5-keyed deduplication index. keeps track of files already on disk
// & of downloads currently in-flight (dedup for concurrent workers)
type Index struct {
	mu       sync.Mutex
	path     string
	entries  map[string]IndexEntry
	inFlight map[string]struct{}
}

// reads an existing index from disk or creates an empty one
func LoadIndex(path string) (*Index, error) {
	idx := &Index{
		path:     path,
		entries:  make(map[string]IndexEntry),
		inFlight: make(map[string]struct{}),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return idx, nil
		}
		return nil, err
	}

	var entries map[string]IndexEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse index %s: %w", path, err)
	}
	if entries == nil {
		entries = make(map[string]IndexEntry)
	}
	idx.entries = entries
	return idx, nil
}

// returns true if the caller should download the file with the given md5,
// returns false if the md5 is already in the index or in-flight
func (idx *Index) Acquire(md5 string) bool {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if _, exists := idx.entries[md5]; exists {
		return false
	}
	if _, exists := idx.inFlight[md5]; exists {
		return false
	}
	idx.inFlight[md5] = struct{}{}
	return true
}

// removes in-flight marker, also adds the file to index if success=true
func (idx *Index) Release(md5 string, success bool, e IndexEntry) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.inFlight, md5)
	if !success {
		return nil
	}
	idx.entries[md5] = e
	return idx.saveUnsafe()
}

func (idx *Index) saveUnsafe() error {
	dir := filepath.Dir(idx.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp := idx.path + ".tmp"
	data, err := json.MarshalIndent(idx.entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, idx.path)
}
