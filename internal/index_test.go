package internal

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestIndexAcquireRelease(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.json")

	idx, err := LoadIndex(path)
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}

	md5 := "abc123"
	if !idx.Acquire(md5) {
		t.Fatal("first Acquire should succeed")
	}
	if idx.Acquire(md5) {
		t.Fatal("second Acquire should fail while in flight")
	}

	entry := IndexEntry{MD5: md5, Filename: "a.jpg", Board: "g", Size: 1234}
	if err := idx.Release(md5, true, entry); err != nil {
		t.Fatalf("Release: %v", err)
	}

	if idx.Acquire(md5) {
		t.Fatal("Acquire should fail after successful Release")
	}

	// reload and verify persistence
	idx2, err := LoadIndex(path)
	if err != nil {
		t.Fatalf("LoadIndex second: %v", err)
	}
	if idx2.Acquire(md5) {
		t.Fatal("Acquire should fail after reloading persisted index")
	}
}

func TestIndexReleaseFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.json")

	idx, err := LoadIndex(path)
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}

	md5 := "failmd5"
	if !idx.Acquire(md5) {
		t.Fatal("first Acquire should succeed")
	}
	if err := idx.Release(md5, false, IndexEntry{}); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if !idx.Acquire(md5) {
		t.Fatal("Acquire should succeed again after failed download")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("index file should not exist when no entries have been saved")
	}
}

func TestIndexConcurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.json")

	idx, err := LoadIndex(path)
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}

	md5 := "concurrent"
	done := make(chan bool, 10)
	acquired := 0
	var mu sync.Mutex
	for i := 0; i < 10; i++ {
		go func() {
			if idx.Acquire(md5) {
				mu.Lock()
				acquired++
				mu.Unlock()
			}
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	if acquired != 1 {
		t.Fatalf("expected exactly 1 Acquire, got %d", acquired)
	}
}
