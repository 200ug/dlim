# dlim

4chan mass media downloader, improved.

## Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `default_output_dir` | string | `"./downloads"` | Directory where downloaded media is saved. |
| `blacklist` | object | `{}` | Filters posts by `keywords`, `usernames`, or `tripcodes`. See `config.example.json`. |
| `max_filesize` | integer | `0` | Maximum accepted file size in bytes. `0` disables the filter. |
| `accepted_ratios` | array of `[w,h]` | `[]` | Whitelist of accepted aspect ratios, e.g. `[[16,9],[4,3]]`. Empty disables the filter. |
| `accepted_extensions` | array of strings | `[]` | Whitelist of accepted file extensions, e.g. `["png","jpg"]`. Empty disables the filter. |
| `only_images` | boolean | `false` | When `true`, skip non-image media (e.g. `.webm`). |
| `max_concurrent_threads` | integer | `32` | Number of concurrent thread fetch workers. |
| `max_concurrent_media` | integer | `64` | Number of concurrent media download workers. |
| `max_retries` | integer | `3` | Maximum number of retries per HTTP request. |
| `request_timeout` | string | `"120s"` | Per-request timeout, including reading the full response body. Accepts Go `time.Duration` values such as `"120s"`, `"2m"`, or `"5m"`. |

By default, the config path is `$CWD/config.json`. See [config.example.json](./config.example.json) for a full reference config.

## Usage

```
# dlim -h
usage: dlim [-config path] [-out dir] <command> [args]

commands:
  thread <url>         download a single thread
  board <abbrev>       download an entire board
  threads <file>       download threads listed one-per-line in a file
```

If you get a lot of errors regarding exceeded context deadlines, you can simply repeat the same command (potentially with higher `request_timeout` and lower `max_concurrent_media` limits) after the initial round finishes to re-attempt downloading the failed ones (while skipping the already downloaded files). The error is usually caused by bottlenecks in the network i/o with bigger files.

