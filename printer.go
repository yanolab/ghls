package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

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
