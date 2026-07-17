package internal

import (
	"sync/atomic"

	"github.com/charmbracelet/log"
)

// atomic counters used to report download progress
type Stats struct {
	ThreadsFetched  atomic.Int64
	ThreadsFailed   atomic.Int64
	MediaFound      atomic.Int64
	MediaFiltered   atomic.Int64
	MediaEnqueued   atomic.Int64
	MediaDownloaded atomic.Int64
	MediaSkipped    atomic.Int64
	MediaFailed     atomic.Int64
}

func (s *Stats) Log() {
	log.Info("progress",
		"threads_fetched", s.ThreadsFetched.Load(),
		"threads_failed", s.ThreadsFailed.Load(),
		"media_found", s.MediaFound.Load(),
		"media_filtered", s.MediaFiltered.Load(),
		"media_enqueued", s.MediaEnqueued.Load(),
		"media_downloaded", s.MediaDownloaded.Load(),
		"media_skipped", s.MediaSkipped.Load(),
		"media_failed", s.MediaFailed.Load(),
	)
}
