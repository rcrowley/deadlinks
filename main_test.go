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
	timeout int = 10
	verbose bool
)

func TestScan(t *testing.T) {
	lists, err := files.AllHTML([]string{"testdata"}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	deadlinks, err := scan(lists, []string{}, &timeout, &verbose)
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
	deadlinks, err := scan(lists, ignored, &timeout, &verbose)
	if err != nil {
		t.Fatal(err)
	}
	if len(deadlinks) != 0 {
		t.Fatal(deadlinks)
	}
}
