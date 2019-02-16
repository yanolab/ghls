package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	home "github.com/mitchellh/go-homedir"
)

var homedir string
var update bool
var params string

func init() {
	dir, err := home.Dir()
	if err != nil {
		panic("cannot find homedir")
	}
	homedir = dir

	flag.BoolVar(&update, "u", false, "get and update repositories")
	flag.StringVar(&params, "p", "", "print value by paramters")

	flag.Parse()
}

func printRepositories(p printer, repos []repository) {
	for _, v := range repos {
		p.Print(v)
	}
}

func main() {
	stdp := &ioprinter{w: os.Stdout, params: params}
	if !update {
		// try to read cache file.
		p := filepath.Join(homedir, cacheFileName)
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

	var mp *multiprinter
	repos, err := listGithubRepositories(tkn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get repositories due to %s\n", err)
		os.Exit(1)
	}
	cache := filepath.Join(homedir, cacheFileName)
	f, err := os.OpenFile(cache, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err == nil {
		defer f.Close()
		mp = &multiprinter{printers: []printer{stdp, &jsonprinter{f}}}
	} else {
		fmt.Fprintf(os.Stderr, "cannot open cache file due to %s", err)
	}

	printRepositories(mp, repos)
}
