package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestReadCache(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	p := jsonprinter{w: tmpfile}
	if err := p.Print(repo); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	repos, err := readCache(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("unexpected to read cache.")
	}
	if repos[0].Name != repo.Name || repos[0].Stars != repo.Stars || repos[0].CreatedAt.UnixNano() != repo.CreatedAt.UnixNano() {
		t.Fatalf("unexpected to read cache. expected %q actual: %q", repo, repos[0])
	}
}
