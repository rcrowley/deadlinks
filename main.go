package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rcrowley/mergician/files"
	"github.com/rcrowley/mergician/html"
)

func Main(args []string, stdin io.Reader, stdout io.Writer) {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	ignore := flags.String("i", "", "file containing links to ignore")
	verbose := flags.Bool("v", false, "print the name of each scanned file to standard error")
	exclude := files.NewStringSliceFlag(flags, "x", "subdirectory of <input> to exclude (may be repeated)")
	flags.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: deadlinks [-i <ignore>] [-v] [-x <exclude>[...]] [<docroot>[...]]
  -i <ignore>   file containing links to ignore
  -v            print the name of each scanned file to standard error
  -x <exclude>  subdirectory of <docroot> to exclude (may be repeated)
  <docroot>     document root directory to scan for dead links (defaults to the current working directory; may be repeated)

Synopsis: deadlinks scans all the HTML documents in <docroot> for dead links (in <a>, <form>, <img>, <link rel="stylesheet">, <script>, and <style> elements).
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
		docroots = []string{"."}
	} else {
		docroots = flags.Args()
	}
	lists := must2(files.AllHTML(docroots, *exclude))

	deadlinks := must2(scan(lists, ignored, verbose))
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
		log.Print(err)
	}
}

func must2[T any](v T, err error) T {
	must(err)
	return v
}

func scan(lists []files.List, ignored []string, verbose *bool) (deadlinks []string, err error) {
	cache := make(map[string]bool)
	for _, list := range lists {
		for _, path := range list.RelativePaths() {
			if *verbose {
				log.Printf("scanning %s", filepath.Join(list.Root(), path))
			}
			dir := filepath.Dir(path)

			in := must2(html.ParseFile(filepath.Join(list.Root(), path)))
			for _, out := range html.FindAll(in, html.Any(
				html.Match(must2(html.ParseString(`<a>`))),
				html.Match(must2(html.ParseString(`<form>`))),
				html.Match(must2(html.ParseString(`<img>`))),
				html.Match(must2(html.ParseString(`<link rel="stylesheet">`))),
				html.Match(must2(html.ParseString(`<script>`))),
				html.Match(must2(html.ParseString(`<style>`))),
			)) {
				href := html.Attr(out, "href")
				if href == "" {
					href = html.Attr(out, "src") // different name but we treat it the same
				}
				if href == "" {
					href = html.Attr(out, "action") // another; treat it the same, too
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
					fragment := u.Fragment
					u.Fragment = ""
					ctx, _ := context.WithDeadline(context.TODO(), time.Now().Add(10*time.Second))
					req, err := http.NewRequestWithContext(ctx, "HEAD", u.String(), nil)
					if err != nil {
						log.Print(err)
						cache[href] = false
					}
					req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:134.0) Gecko/20100101 Firefox/134.0") // pretend a bit to be real
					resp, err := http.DefaultClient.Do(req)
					if err == nil {
						if resp.StatusCode < http.StatusBadRequest {
							cache[href] = true
						} else if resp.StatusCode == http.StatusForbidden { // often used to ward off scrapers
							cache[href] = true
						} else {
							if *verbose {
								log.Printf("<%s>: %s", u, resp.Status)
							}
							cache[href] = false
						}
					} else {
						if *verbose {
							log.Printf("<%s>: %v", u, err)
						}
						cache[href] = false
					}
					u.Fragment = fragment

				} else if u.Scheme == "mailto" {
					cache[href] = true // TODO test this mailbox by actually connecting to the SMTP server

				} else if u.Scheme == "tel" {
					cache[href] = true // we're not going to try to verify phone numbers, come on

				} else if u.Path != "" {
					hrefPath := u.Path
					if !strings.HasPrefix(u.Path, "/") {
						hrefPath = filepath.Join(dir, u.Path)
					}
					if fi, err := os.Stat(filepath.Join(list.Root(), hrefPath)); err == nil && !fi.IsDir() {
						cache[href] = true
					} else if fi, err := os.Stat(filepath.Join(list.Root(), hrefPath, "index.html")); err == nil && !fi.IsDir() {
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
