# dlim

4chan mass media downloader, improved.

## Usage

```
# dlim -h
usage: dlim [-config path] [-out dir] <command> [args]

commands:
  thread <url>         download a single thread
  board <abbrev>       download an entire board
  threads <file>       download threads listed one-per-line in a file
```

By default, config path is `$CWD/config.json` and output directory `$CWD/downloads/`. See [config.example.json](./config.example.json) for reference config. Filters are skipped if they're not configured (i.e. "accepted values" list is empty or the configured max. file size is zero or negative).

If you get errors regarding exceeded context deadlines or canceled requests, you can simply repeat the same command after the initial round finishes. This'll skip the already downloaded files, and re-attempt to download the failed ones.

