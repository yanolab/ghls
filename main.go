package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type _params struct {
	homedir      string
	update       bool
	params       string
	user         string
	disableCache bool
	cleanCache   bool
}

var params _params

func parseFlag() {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic("cannot find homedir")
	}
	params.homedir = dir

	flag.BoolVar(&params.update, "u", false, "get and update repositories cache")
	flag.StringVar(&params.params, "p", "", "print value by paramters")
	flag.StringVar(&params.user, "user", "", "target user")
	flag.BoolVar(&params.disableCache, "disablecache", false, "caching is disable")
	flag.BoolVar(&params.cleanCache, "cleancache", false, "clean cache")

	flag.Parse()
}

func printRepositories(p printer, repos []repository) {
	for _, v := range repos {
		p.Print(v)
	}
}

func main() {
	parseFlag()

	if params.cleanCache {
		p := filepath.Join(params.homedir, cacheFileName)
		if err := os.Remove(p); err != nil {
			fmt.Fprintf(os.Stderr, "failed to clean cache due to %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	stdp := &ioprinter{w: os.Stdout, params: params.params}
	if !params.update && !params.disableCache {
		// try to read cache file.
		p := filepath.Join(params.homedir, cacheFileName)
		if caches, err := readCache(p); err == nil {
			printRepositories(stdp, caches)
			os.Exit(0)
		}
	}

	tkn, err := getToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get github token due to %s\n", err)
		os.Exit(1)
	}

	opts := []listOption{
		withUser(params.user),
		withToken(tkn),
	}

	var mp *multiprinter
	repos, err := listGithubRepositories(opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get repositories due to %s\n", err)
		os.Exit(1)
	}
	printers := make([]printer, 0, 1)
	printers = append(printers, stdp)
	if !params.disableCache {
		cache := filepath.Join(params.homedir, cacheFileName)
		f, err := os.OpenFile(cache, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err == nil {
			defer f.Close()
			printers = append(printers, &jsonprinter{f})
		} else {
			fmt.Fprintf(os.Stderr, "cannot open cache file due to %s", err)
		}
	}
	mp = &multiprinter{printers: printers}

	printRepositories(mp, repos)
}
