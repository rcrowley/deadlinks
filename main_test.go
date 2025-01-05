package main

import (
	"slices"
	"testing"

	"github.com/rcrowley/mergician/files"
)

var (
	ignored = []string{"../dead.html", "/dead.html", "/dead/", "dead/", "https://rcrowley.org/dead.html"}
	verbose bool
)

func TestScan(t *testing.T) {
	lists, err := files.AllHTML([]string{"testdata"}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	deadlinks, err := scan(lists, []string{}, &verbose)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(deadlinks, ignored) {
		t.Fatal(deadlinks)
	}
}

func TestScanIgnore(t *testing.T) {
	lists, err := files.AllHTML([]string{"testdata"}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	deadlinks, err := scan(lists, ignored, &verbose)
	if err != nil {
		t.Fatal(err)
	}
	if len(deadlinks) != 0 {
		t.Fatal(deadlinks)
	}
}
