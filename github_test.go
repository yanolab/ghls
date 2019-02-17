package main

import (
	"os"
	"testing"
)

func TestGetToken(t *testing.T) {
	const token = "DUMMYTOKEN"
	if err := os.Setenv(tokenName, token); err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv(tokenName); got != token {
		t.Fatalf("unexpected token. expected: %s, actual: %s", token, got)
	}

	got, err := getToken()
	if err != nil {
		t.Fatal(err)
	}
	if got != token {
		t.Fatalf("unexpected token. expected: %s, actual: %s", token, got)
	}
}
