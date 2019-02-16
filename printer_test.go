package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

var repo = repository{
	Name:        "testrepo",
	FullName:    "testuser/testrepo",
	Description: "this is test repository",
	Owner:       "testuser",
	Stars:       7777,
	URL:         "http://www.google.com/",
	PushedAt:    time.Now(),
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
}

func TestIOPrinter(t *testing.T) {
	buf := new(bytes.Buffer)
	p := &ioprinter{w: buf, params: "name,stars"}
	if err := p.Print(repo); err != nil {
		t.Fatal(err)
	}

	exp := fmt.Sprintf("%s %d\n", repo.Name, repo.Stars)
	if buf.String() != exp {
		t.Fatalf("unxpected value. expected: %q actual: %q", exp, buf.String())
	}
}

func TestJsonPrinter(t *testing.T) {
	buf := new(bytes.Buffer)
	p := &jsonprinter{w: buf}
	if err := p.Print(repo); err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(repo)
	if err != nil {
		t.Fatal(err)
	}

	exp := string(b) + "\n" // printer adds `\n` to end of line.
	if buf.String() != exp {
		t.Fatalf("unxpected value. expected: %q actual: %q", exp, buf.String())
	}
}

func TestMultiPrinter(t *testing.T) {
	b1 := new(bytes.Buffer)
	b2 := new(bytes.Buffer)
	bp1 := &ioprinter{w: b1, params: "name"}
	bp2 := &ioprinter{w: b2, params: "stars"}
	mp := multiprinter{[]printer{bp1, bp2}}

	if err := mp.Print(repo); err != nil {
		t.Fatal(err)
	}

	exp := fmt.Sprintf("%s\n", repo.Name)
	if b1.String() != exp {
		t.Fatalf("unexpected value, expected: %q actual: %q", exp, b1.String())
	}
	exp = fmt.Sprintf("%d\n", repo.Stars)
	if b2.String() != exp {
		t.Fatalf("unexpected value. expected %q, actual: %q", exp, b2.String())
	}
}
