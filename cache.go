package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"time"
)

const (
	cacheFileName = ".ghls_cache"
)

var (
	errCacheIsExpired = errors.New("cache is expired")
)

func readCache(p string) ([]repository, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	// check cache file is flesh.
	if time.Since(fi.ModTime()) >= 24*time.Hour {
		return nil, errCacheIsExpired
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bufs := make([]repository, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var r repository
		b := scanner.Text()
		if err := json.Unmarshal([]byte(b), &r); err != nil {
			return nil, err
		}
		bufs = append(bufs, r)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return bufs, nil
}
