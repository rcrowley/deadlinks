package main

import (
	"slices"
	"testing"
)

var (
	ignored = []string{"../dead.html", "/dead.html", "/dead/", "dead/", "https://rcrowley.org/dead.html"}
	verbose bool
)

func TestScan(t *testing.T) {
	deadlinks, err := scan([]string{"testdata"}, []string{}, &verbose)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(deadlinks, ignored) {
		t.Fatal(deadlinks)
	}
}

func TestScanIgnore(t *testing.T) {
	deadlinks, err := scan([]string{"testdata"}, ignored, &verbose)
	if err != nil {
		t.Fatal(err)
	}
	if len(deadlinks) != 0 {
		t.Fatal(deadlinks)
	}
}
