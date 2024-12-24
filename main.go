package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rcrowley/mergician/html"
)

func Main(args []string, stdin io.Reader, stdout io.Writer) {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	ignore := flags.String("i", "", "file containing links to ignore")
	verbose := flags.Bool("v", false, "print the name of each scanned file to standard error")
	flags.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: deadlinks [-i <ignore>] [-v] [<docroot>[...]]
  -i <ignore>  file containing links to ignore
  <docroot>    document root directory to scan for dead links (defaults to the current working directory)

Synopsis: deadlinks scans one or more directories for <a>, <img>, <link rel="stylesheet">, <script>, and <style> elements with HTTP(S) URLs and reports any which do not respond with an HTTP status less than 400.
`)
	}
	flags.Parse(args[1:])

	var ignored []string
	if *ignore != "" {
		f := must2(os.Open(*ignore))
		defer f.Close()
		r := bufio.NewReader(f)
		for {
			s, err := r.ReadString('\n')
			if err == io.EOF {
				break
			}
			must(err)
			ignored = append(ignored, strings.TrimSpace(s))
		}
	}
	sort.Strings(ignored)

	var docroots []string
	if flags.NArg() == 0 {
		docroots = []string{""}
	} else {
		docroots = flags.Args()
	}

	deadlinks := must2(scan(docroots, ignored, verbose)) // TODO use files.All
	if *verbose {
		fmt.Fprintf(os.Stderr, "\nfound %d dead links", len(deadlinks))
		if len(deadlinks) > 0 {
			fmt.Fprintln(os.Stderr, ":")
		} else {
			fmt.Fprintln(os.Stderr, "!")
		}
	}
	for _, href := range deadlinks {
		fmt.Println(href)
	}
	if len(deadlinks) > 0 {
		os.Exit(1)
	}
}

func init() {
	log.SetFlags(0)
}

func main() {
	Main(os.Args, os.Stdin, os.Stdout)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func must2[T any](v T, err error) T {
	must(err)
	return v
}

func scan(dirs, ignored []string, verbose *bool) (deadlinks []string, err error) {
	cache := make(map[string]bool)
	for _, dir := range dirs {
		if dir, err = filepath.Abs(dir); err != nil {
			return
		}
		walk := walker(dir, cache, ignored, verbose)
		var fi fs.FileInfo
		if fi, err = os.Stat(dir); err != nil {
			return
		}
		if fi.IsDir() {
			err = fs.WalkDir(os.DirFS(dir), ".", walk)
		} else {
			err = walk(dir, fs.FileInfoToDirEntry(fi), nil)
		}
		if err != nil {
			return
		}
	}

	for href, ok := range cache {
		if !ok {
			deadlinks = append(deadlinks, href)
		}
	}
	sort.Strings(deadlinks)
	return
}

func walker(dir string, cache map[string]bool, ignored []string, verbose *bool) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		must(err)

		if !d.Type().IsRegular() || filepath.Ext(path) != ".html" {
			return nil
		}
		if *verbose {
			log.Printf("scanning %s\n", filepath.Join(dir, path))
		}

		in := must2(html.ParseFile(filepath.Join(dir, path)))
		for _, out := range html.FindAll(in, html.Any(
			html.Match(must2(html.ParseString(`<a>`))),
			html.Match(must2(html.ParseString(`<img>`))),
			html.Match(must2(html.ParseString(`<link rel="stylesheet">`))),
			html.Match(must2(html.ParseString(`<script>`))),
			html.Match(must2(html.ParseString(`<style>`))),
		)) {
			href := html.Attr(out, "href")
			if href == "" {
				href = html.Attr(out, "src") // different name but we treat it the same
			}
			if href == "" || href == "#" {
				continue
			}

			if _, ok := cache[href]; ok {
				continue
			}

			if i := sort.SearchStrings(ignored, href); i < len(ignored) && ignored[i] == href {
				if *verbose {
					log.Printf("ignoring %s", href)
				}
				continue
			}

			u, err := url.Parse(href)
			if err != nil {
				log.Print(err)
				cache[href] = false
			}

			if u.Scheme == "http" || u.Scheme == "https" {
				resp, err := http.Head(u.String())
				cache[href] = err == nil && resp.StatusCode < http.StatusBadRequest

			} else if u.Scheme == "mailto" {
				cache[href] = true // TODO test this mailbox by actually connecting to the SMTP server

			} else if u.Scheme == "tel" {
				cache[href] = true // we're not going to try to verify phone numbers, come on

			} else if u.Path != "" {
				hrefPath := u.Path
				if !strings.HasPrefix(u.Path, "/") {
					hrefPath = filepath.Join(filepath.Dir(path), u.Path)
				}
				if fi, err := os.Stat(filepath.Join(dir, hrefPath)); err == nil && !fi.IsDir() {
					cache[href] = true
				} else if fi, err := os.Stat(filepath.Join(dir, hrefPath, "index.html")); err == nil && !fi.IsDir() {
					cache[href] = true
				} else {
					cache[href] = false
				}

			} else if fragment := u.EscapedFragment(); fragment != "" {
				matcher := html.HasAttr("id", fragment)
				if id, err := url.QueryUnescape(fragment); err == nil {
					matcher = html.Any(matcher, html.HasAttr("id", id))
				}
				cache[href] = html.Find(in, matcher) != nil

			} else {
				cache[href] = false
				log.Printf("unclear how to test %s", href)

			}
		}

		return nil
	}
}
