package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

func main() {
	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opt := new(github.RepositoryListOptions)
	opt.Page = 1
	opt.PerPage = 100

	for {
		repos, res, err := client.Repositories.List(ctx, "", opt)

		if err != nil {
			panic(err)
		}

		for _, repo := range repos {
			fmt.Printf("%#v\n", *repo.FullName)
		}

		if res.NextPage == 0 {
			break
		}

		opt.Page = res.NextPage
	}
}
