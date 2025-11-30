package main

import (
	"slices"
	"testing"

	"github.com/rcrowley/mergician/files"
)

var (
	ignored = []string{ // or the expected result when not ignoring anything
		"../dead.html",
		"/dead.css",
		"/dead.html",
		"/dead.js",
		"/dead.php",
		"/dead/",
		"dead/",
		"https://rcrowley.org/dead.html",
	}
	retries int = 0
	timeout int = 1 // seconds
	verbose bool
)

func TestScan(t *testing.T) {
	lists, err := files.AllHTML([]string{"testdata"}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	deadlinks, err := scan(lists, []string{}, &retries, &timeout, &verbose)
	if err != nil {
		t.Fatal(err)
	}
	hrefs := make([]string, len(deadlinks))
	for i, d := range deadlinks {
		hrefs[i] = d.href
	}
	if !slices.Equal(hrefs, ignored) {
		t.Fatal(deadlinks)
	}
}

func TestScanIgnore(t *testing.T) {
	lists, err := files.AllHTML([]string{"testdata"}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	deadlinks, err := scan(lists, ignored, &retries, &timeout, &verbose)
	if err != nil {
		t.Fatal(err)
	}
	if len(deadlinks) != 0 {
		t.Fatal(deadlinks)
	}
}
