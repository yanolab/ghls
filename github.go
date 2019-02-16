package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	tokenName = "GITHUB_TOKEN"
)

var (
	errTokenNotFound = errors.New("github token is not found")
)

type repository struct {
	Name        string    `json:"name"`
	FullName    string    `json:"fullname"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	Stars       int       `json:"stars"`
	URL         string    `json:"url"`
	PushedAt    time.Time `json:"pushed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
				Stars:       repo.GetStargazersCount(),
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
