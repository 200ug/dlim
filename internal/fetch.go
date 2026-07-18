package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

const (
	defaultUserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.3"
	defaultThreadWorkers  = 32
	defaultMediaWorkers   = 64
	defaultMaxRetries     = 3
	defaultRequestTimeout = 120 * time.Second
)

type downloadTask struct {
	board string
	post  ThreadItem
	url   string
}

type userAgentTransport struct {
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", defaultUserAgent)
	return t.base.RoundTrip(req)
}

type Downloader struct {
	client http.Client
	config RunConfig
	index  *Index
	stats  Stats
}

func NewDownloader(rc RunConfig) (*Downloader, error) {
	if rc.DefaultOutDir == "" {
		return nil, fmt.Errorf("default_output_dir is required")
	}

	indexPath := filepath.Join(rc.DefaultOutDir, ".dlim-index.json")
	idx, err := LoadIndex(indexPath)
	if err != nil {
		return nil, err
	}

	timeout := defaultRequestTimeout
	if rc.RequestTimeout != "" {
		parsed, err := time.ParseDuration(rc.RequestTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid request_timeout %q: %w", rc.RequestTimeout, err)
		}
		if parsed <= 0 {
			return nil, fmt.Errorf("invalid request_timeout %q: must be positive", rc.RequestTimeout)
		}
		timeout = parsed
	}

	return &Downloader{
		client: http.Client{
			Timeout: timeout,
			Transport: &userAgentTransport{
				base: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 100,
					IdleConnTimeout:     90 * time.Second,
				},
			},
		},
		config: rc,
		index:  idx,
	}, nil
}

func (d *Downloader) maxRetries() int {
	if d.config.MaxRetries > 0 {
		return d.config.MaxRetries
	}
	return defaultMaxRetries
}

func (d *Downloader) getWithRetry(ctx context.Context, url string) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= d.maxRetries(); attempt++ {
		if attempt > 0 {
			delay := 500 * time.Millisecond * (1 << (attempt - 1))
			if delay > 10*time.Second {
				delay = 10 * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		lastResp, lastErr = d.client.Do(req)
		if lastErr != nil {
			continue
		}

		if lastResp.StatusCode < 400 {
			return lastResp, nil
		}
		if lastResp.StatusCode < 500 && lastResp.StatusCode != http.StatusTooManyRequests {
			lastErr = fmt.Errorf("unexpected status %d for %s", lastResp.StatusCode, url)
			return lastResp, lastErr
		}
		if attempt < d.maxRetries() {
			lastResp.Body.Close()
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("unexpected status %d for %s", lastResp.StatusCode, url)
	}
	return lastResp, lastErr
}

func (d *Downloader) FetchBoard(ctx context.Context, boardShort string) error {
	pages, err := d.fetchCatalog(ctx, boardShort)
	if err != nil {
		return err
	}

	threadWorkers := d.config.MaxConcurrentThreads
	if threadWorkers <= 0 {
		threadWorkers = defaultThreadWorkers
	}
	mediaWorkers := d.config.MaxConcurrentMedia
	if mediaWorkers <= 0 {
		mediaWorkers = defaultMediaWorkers
	}

	threadCh := make(chan Thread, threadWorkers*2)
	mediaCh := make(chan downloadTask, mediaWorkers*4)

	var threadWG sync.WaitGroup
	for i := 0; i < threadWorkers; i++ {
		threadWG.Add(1)
		go func() {
			defer threadWG.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-threadCh:
					if !ok {
						return
					}
					if err := d.fetchThreadAndEnqueue(ctx, boardShort, t, mediaCh); err != nil {
						log.Error("failed fetching thread", "board", boardShort, "thread_id", t.ID, "err", err)
					}
				}
			}
		}()
	}

	var mediaWG sync.WaitGroup
	for i := 0; i < mediaWorkers; i++ {
		mediaWG.Add(1)
		go func() {
			defer mediaWG.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-mediaCh:
					if !ok {
						return
					}
					if err := d.downloadMedia(ctx, task); err != nil {
						log.Error("failed downloading media", "url", task.url, "err", err)
					}
				}
			}
		}()
	}

	go func() {
		for _, p := range pages {
			for _, t := range p.Threads {
				if d.config.Blacklist.MatchesThread(t) {
					continue
				}
				select {
				case threadCh <- t:
				case <-ctx.Done():
					close(threadCh)
					return
				}
			}
		}
		close(threadCh)
	}()

	go func() {
		threadWG.Wait()
		close(mediaCh)
	}()

	progressDone := make(chan struct{})
	go d.logProgress(ctx, progressDone)

	mediaWG.Wait()
	close(progressDone)
	d.stats.Log()
	return nil
}

func (d *Downloader) logProgress(ctx context.Context, done <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.stats.Log()
		}
	}
}

func (d *Downloader) fetchCatalog(ctx context.Context, boardShort string) ([]Page, error) {
	url := fmt.Sprintf("https://a.4cdn.org/%s/catalog.json", boardShort)
	resp, err := d.getWithRetry(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pages []Page
	if err := json.Unmarshal(body, &pages); err != nil {
		return nil, err
	}
	return pages, nil
}

func (d *Downloader) FetchSingleThread(ctx context.Context, boardShort string, threadID int64) error {
	return d.FetchThreads(ctx, boardShort, []int64{threadID})
}

func (d *Downloader) FetchThreads(ctx context.Context, boardShort string, threadIDs []int64) error {
	mediaWorkers := d.config.MaxConcurrentMedia
	if mediaWorkers <= 0 {
		mediaWorkers = defaultMediaWorkers
	}

	mediaCh := make(chan downloadTask, mediaWorkers*4)
	var mediaWG sync.WaitGroup
	for i := 0; i < mediaWorkers; i++ {
		mediaWG.Add(1)
		go func() {
			defer mediaWG.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-mediaCh:
					if !ok {
						return
					}
					if err := d.downloadMedia(ctx, task); err != nil {
						log.Error("failed downloading media", "url", task.url, "err", err)
					}
				}
			}
		}()
	}

	progressDone := make(chan struct{})
	go d.logProgress(ctx, progressDone)

	for _, id := range threadIDs {
		if err := d.fetchThreadAndEnqueue(ctx, boardShort, Thread{ID: id}, mediaCh); err != nil {
			log.Error("failed fetching thread", "board", boardShort, "thread_id", id, "err", err)
		}
		select {
		case <-ctx.Done():
			close(mediaCh)
			mediaWG.Wait()
			close(progressDone)
			return ctx.Err()
		default:
		}
	}
	close(mediaCh)
	mediaWG.Wait()
	close(progressDone)
	d.stats.Log()
	return nil
}

func (d *Downloader) fetchThreadAndEnqueue(ctx context.Context, boardShort string, t Thread, mediaCh chan<- downloadTask) error {
	url := fmt.Sprintf("https://a.4cdn.org/%s/thread/%d.json", boardShort, t.ID)
	resp, err := d.getWithRetry(ctx, url)
	if err != nil {
		d.stats.ThreadsFailed.Add(1)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.stats.ThreadsFailed.Add(1)
		return err
	}

	var thread Thread
	if err := json.Unmarshal(body, &thread); err != nil {
		d.stats.ThreadsFailed.Add(1)
		return err
	}

	d.stats.ThreadsFetched.Add(1)
	return thread.fetchMedia(ctx, d, boardShort, mediaCh)
}

func (t *Thread) fetchMedia(ctx context.Context, d *Downloader, boardShort string, mediaCh chan<- downloadTask) error {
	for _, post := range t.Posts {
		if post.ImageID == 0 {
			continue
		}
		d.stats.MediaFound.Add(1)
		if d.config.Blacklist.MatchesPost(post) {
			d.stats.MediaFiltered.Add(1)
			continue
		}
		if !d.config.AcceptsMedia(post) {
			d.stats.MediaFiltered.Add(1)
			continue
		}
		select {
		case mediaCh <- downloadTask{
			board: boardShort,
			post:  post,
			url:   fmt.Sprintf("https://i.4cdn.org/%s/%d%s", boardShort, post.ImageID, post.ImageExt),
		}:
			d.stats.MediaEnqueued.Add(1)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (d *Downloader) downloadMedia(ctx context.Context, task downloadTask) error {
	if task.post.ImageMD5 == "" {
		d.stats.MediaFailed.Add(1)
		return fmt.Errorf("post %d has no md5", task.post.ID)
	}

	if !d.index.Acquire(task.post.ImageMD5) {
		d.stats.MediaSkipped.Add(1)
		return nil // already downloaded or in flight
	}

	outDir := filepath.Join(d.config.DefaultOutDir, task.board)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		d.index.Release(task.post.ImageMD5, false, IndexEntry{})
		d.stats.MediaFailed.Add(1)
		return err
	}

	filename := fmt.Sprintf("%d%s", task.post.ImageID, task.post.ImageExt)
	outPath := filepath.Join(outDir, filename)

	resp, err := d.getWithRetry(ctx, task.url)
	if err != nil {
		d.index.Release(task.post.ImageMD5, false, IndexEntry{})
		d.stats.MediaFailed.Add(1)
		return err
	}
	defer resp.Body.Close()

	tmp, err := os.CreateTemp(outDir, fmt.Sprintf("%d.*.tmp", task.post.ImageID))
	if err != nil {
		d.index.Release(task.post.ImageMD5, false, IndexEntry{})
		d.stats.MediaFailed.Add(1)
		return err
	}
	tmpPath := tmp.Name()

	size, err := io.Copy(tmp, resp.Body)
	tmp.Close()
	if err != nil {
		os.Remove(tmpPath)
		d.index.Release(task.post.ImageMD5, false, IndexEntry{})
		d.stats.MediaFailed.Add(1)
		return err
	}

	if err := os.Rename(tmpPath, outPath); err != nil {
		os.Remove(tmpPath)
		d.index.Release(task.post.ImageMD5, false, IndexEntry{})
		d.stats.MediaFailed.Add(1)
		return err
	}

	d.stats.MediaDownloaded.Add(1)
	return d.index.Release(task.post.ImageMD5, true, IndexEntry{
		MD5:      task.post.ImageMD5,
		Filename: filename,
		Board:    task.board,
		Size:     int(size),
		Width:    task.post.ImageW,
		Height:   task.post.ImageH,
		URL:      task.url,
	})
}
