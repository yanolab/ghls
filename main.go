package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
var params string

type repository struct {
	Name        string    `json:"name"`
	FullName    string    `json:"fullname"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	StartCount  int       `json:"start_count"`
	URL         string    `json:"url"`
	PushedAt    time.Time `json:"pushed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func init() {
	dir, err := home.Dir()
	if err != nil {
		panic("cannot find homedir")
	}
	homedir = dir

	flag.BoolVar(&update, "u", false, "get and update repositories")
	flag.StringVar(&params, "p", "", "print paramters")

	flag.Parse()
}

func getTokenFromHub() (string, error) {
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

func getToken() (string, error) {
	tkn := os.Getenv(tokenName)
	if tkn != "" {
		return tkn, nil
	}

	return getTokenFromHub()
}

func readCache() ([]repository, error) {
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

type printer interface {
	Print(repository) error
}

type multiprinter struct {
	printers []printer
}

func (p *multiprinter) Print(repo repository) error {
	for _, p := range p.printers {
		p.Print(repo)
	}
	return nil
}

type stdprinter struct {
	io.Writer
	params string
}

func (p *stdprinter) Print(repo repository) error {
	parts := strings.Split(p.params, ",")

	buf := make([]string, 0)
	for _, v := range parts {
		switch strings.ToLower(v) {
		case "name":
			buf = append(buf, repo.Name)
		case "fullname":
			buf = append(buf, repo.FullName)
		case "owner":
			buf = append(buf, repo.Owner)
		case "star_count":
			buf = append(buf, strconv.Itoa(repo.StartCount))
		case "pushed_at":
			buf = append(buf, repo.PushedAt.String())
		case "created_at":
			buf = append(buf, repo.CreatedAt.String())
		case "updated_at":
			buf = append(buf, repo.UpdatedAt.String())
		case "description":
			buf = append(buf, repo.Description)
		case "url":
			buf = append(buf, repo.URL)
		}
	}

	if len(buf) > 0 {
		_, err := fmt.Fprintln(p.Writer, strings.Join(buf, " "))
		return err
	}

	_, err := fmt.Fprintln(p.Writer, repo.FullName)
	return err
}

type jsonprinter struct {
	io.Writer
}

func (p *jsonprinter) Print(repo repository) error {
	buf, err := json.Marshal(repo)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(p.Writer, string(buf))
	return err
}

func printRepositories(p printer, repos []repository) {
	for _, v := range repos {
		p.Print(v)
	}
}

func listGithubRepositories(token string) ([]repository, error) {
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

	repositories := make([]repository, 0)
	for {
		repos, res, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			panic(err)
		}

		for _, repo := range repos {
			r := repository{
				Name:        repo.GetName(),
				FullName:    repo.GetFullName(),
				Description: repo.GetDescription(),
				Owner:       repo.GetOwner().GetLogin(),
				StartCount:  repo.GetStargazersCount(),
				PushedAt:    repo.GetPushedAt().UTC(),
				URL:         repo.GetHTMLURL(),
				CreatedAt:   repo.GetCreatedAt().UTC(),
				UpdatedAt:   repo.GetUpdatedAt().UTC(),
			}
			repositories = append(repositories, r)
		}

		if res.NextPage == 0 {
			break
		}

		opt.Page = res.NextPage
	}

	return repositories, nil
}

func main() {
	stdp := &stdprinter{Writer: os.Stdout, params: params}
	if !update {
		if caches, err := readCache(); err == nil {
			printRepositories(stdp, caches)
			os.Exit(0)
		}
	}

	tkn, err := getToken()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	var mp *multiprinter
	repos, err := listGithubRepositories(tkn)
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
