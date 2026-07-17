package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/200ug/dlim/internal"
	"github.com/charmbracelet/log"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: dlim [-config path] [-out dir] <command> [args]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  thread <url>         download a single thread")
	fmt.Fprintln(os.Stderr, "  board <abbrev>       download an entire board")
	fmt.Fprintln(os.Stderr, "  threads <file>       download threads listed one-per-line in a file")
}

func main() {
	fs := flag.NewFlagSet("dlim", flag.ExitOnError)
	fs.Usage = usage
	configPath := fs.String("config", "config.json", "path to config file")
	outDir := fs.String("out", "", "output directory (overrides config)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	args := fs.Args()
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	rc, err := loadConfig(*configPath, *outDir)
	if err != nil {
		log.Fatal("failed to load config", "err", err)
	}

	dl, err := internal.NewDownloader(rc)
	if err != nil {
		log.Fatal("failed to create downloader", "err", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch args[0] {
	case "thread":
		if len(args) != 2 {
			usage()
			os.Exit(1)
		}
		if err := runThread(ctx, dl, args[1]); err != nil {
			exitOnError(err)
		}
	case "board":
		if len(args) != 2 {
			usage()
			os.Exit(1)
		}
		if err := dl.FetchBoard(ctx, args[1]); err != nil {
			exitOnError(err)
		}
	case "threads":
		if len(args) != 2 {
			usage()
			os.Exit(1)
		}
		if err := runThreadsFromFile(ctx, dl, args[1]); err != nil {
			exitOnError(err)
		}
	default:
		usage()
		os.Exit(1)
	}
}

func exitOnError(err error) {
	if errors.Is(err, context.Canceled) {
		os.Exit(0)
	}
	log.Fatal("download failed", "err", err)
}

func loadConfig(path, outOverride string) (internal.RunConfig, error) {
	rc := internal.RunConfig{
		DefaultOutDir: "./downloads",
	}

	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &rc); err != nil {
			return rc, fmt.Errorf("invalid config %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return rc, err
	}

	if outOverride != "" {
		rc.DefaultOutDir = outOverride
	}

	return rc, nil
}

func runThread(ctx context.Context, dl *internal.Downloader, url string) error {
	board, threadID, err := internal.ParseThreadURL(url)
	if err != nil {
		return err
	}
	return dl.FetchSingleThread(ctx, board, threadID)
}

func runThreadsFromFile(ctx context.Context, dl *internal.Downloader, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	urlsByBoard := make(map[string][]int64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url == "" || strings.HasPrefix(url, "#") {
			continue
		}
		board, threadID, err := internal.ParseThreadURL(url)
		if err != nil {
			log.Error("skipping invalid thread URL", "url", url, "err", err)
			continue
		}
		urlsByBoard[board] = append(urlsByBoard[board], threadID)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for board, ids := range urlsByBoard {
		if err := dl.FetchThreads(ctx, board, ids); err != nil {
			return err
		}
	}
	return nil
}
