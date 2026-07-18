package internal

type Blacklist struct {
	Keywords  []string `json:"keywords,omitempty"`
	Usernames []string `json:"usernames,omitempty"`
	Tripcodes []string `json:"tripcodes,omitempty"`
}

type RunConfig struct {
	DefaultOutDir        string        `json:"default_output_dir"`
	Blacklist            Blacklist     `json:"blacklist"`
	MaxFilesize          int           `json:"max_filesize,omitempty"`
	AcceptedRatios       []AspectRatio `json:"accepted_ratios,omitempty"`
	AcceptedExtensions   []string      `json:"accepted_extensions,omitempty"`
	OnlyImages           bool          `json:"only_images,omitempty"`
	MaxConcurrentThreads int           `json:"max_concurrent_threads,omitempty"`
	MaxConcurrentMedia   int           `json:"max_concurrent_media,omitempty"`
	MaxRetries           int           `json:"max_retries,omitempty"`
	RequestTimeout       string        `json:"request_timeout,omitempty"`
}

// https://a.4cdn.org/<board>/catalog.json
type Page struct {
	Threads []Thread `json:"threads"`
}

// https://a.4cdn.org/<board>/thread/<id>.json
type Thread struct {
	ID            int64        `json:"no"`
	OPTitle       string       `json:"sub"`
	OPDescription string       `json:"com,omitempty"`
	Posts         []ThreadItem `json:"posts"`
}

// https://i.4cdn.org/<board>/<id><ext>
type ThreadItem struct {
	ID         int64  `json:"no"`
	OPName     string `json:"name"`
	OPTripcode string `json:"trip,omitempty"`
	ImageID    int64  `json:"tim,omitempty"` // missing if no media attached (0)
	ImageExt   string `json:"ext,omitempty"` // including dot, e.g. ".webm"
	ImageW     int    `json:"w,omitempty"`
	ImageH     int    `json:"h,omitempty"`
	ImageMD5   string `json:"md5,omitempty"`
	ImageSize  int    `json:"fsize,omitempty"`
}
