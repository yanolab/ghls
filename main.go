package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-github/github"
	home "github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

const (
	tokenName     = "GITHUB_TOKEN"
	cacheFileName = ".ghls_cache"
)

var (
	errTokenNotFound  = errors.New("github token is not found")
	errCacheIsExpired = errors.New("cache is expired")
)

var homedir string
var update bool

func init() {
	dir, err := home.Dir()
	if err != nil {
		panic("cannot find homedir")
	}
	homedir = dir

	flag.BoolVar(&update, "u", false, "get and update repositories")

	flag.Parse()
}

func getToken() (string, error) {
	tkn := os.Getenv(tokenName)
	if tkn != "" {
		return tkn, nil
	}

	// try to read hub config file.
	cfg := filepath.Join(homedir, ".config", "hub")
	f, err := os.Open(cfg)
	if err != nil {
		return "", err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	m := new(map[string][]map[string]string)
	if err := d.Decode(m); err != nil {
		return "", err
	}

	items, ok := (*m)["github.com"]
	if !ok || len(items) == 0 {
		return "", errTokenNotFound
	}

	for _, v := range items {
		if tkn, ok := v["oauth_token"]; ok {
			return tkn, nil
		}
	}

	return "", errTokenNotFound
}

func readCache() ([]string, error) {
	// try to read cache file.
	cache := filepath.Join(homedir, cacheFileName)
	fi, err := os.Stat(cache)
	if err != nil {
		return nil, err
	}
	// check cache file is flesh.
	if time.Now().Sub(fi.ModTime()).Hours() >= 24.0 {
		return nil, errCacheIsExpired
	}

	f, err := os.Open(cache)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lines := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func printRepositories(out io.Writer, repos []string) {
	for _, v := range repos {
		fmt.Fprintln(out, v)
	}
}

func listGithubRepositories(token string) ([]string, error) {
	ctx := context.Background()

	tkn, err := getToken()
	if err != nil {
		return nil, err
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tkn})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opt := new(github.RepositoryListOptions)
	opt.Page = 1
	opt.PerPage = 100

	names := make([]string, 0)
	for {
		repos, res, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			panic(err)
		}

		for _, repo := range repos {
			names = append(names, *repo.FullName)
		}

		if res.NextPage == 0 {
			break
		}

		opt.Page = res.NextPage
	}

	return names, nil
}

func main() {
	if !update {
		if caches, err := readCache(); err == nil {
			printRepositories(os.Stdout, caches)
			os.Exit(0)
		}
	}

	tkn, err := getToken()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	var w io.Writer = os.Stdout
	repos, err := listGithubRepositories(tkn)
	cache := filepath.Join(homedir, cacheFileName)
	f, err := os.OpenFile(cache, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err == nil {
		w = io.MultiWriter(os.Stdout, f)
		defer f.Close()
	} else {
		fmt.Println(err)
	}

	printRepositories(w, repos)
}
