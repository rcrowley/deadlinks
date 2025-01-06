Deadlinks
=========

Deadlinks scans all the HTML documents in a document root directory for dead links. A link is considered dead if it refers to a local URL path, relative or absolute, that doesn't exist or responds with an HTTP status less than 400. It scans all `<a href="...">`, `<form action="...">`, `<img src="...">`, `<link href="..." rel="stylesheet">`, `<script src="...">`, or `<style src="...">` elements.

It's intended to be used almost like spellcheck in a CI process or pre-commit hook.

Installation
------------

```sh
go install github.com/rcrowley/deadlinks@latest
```

Usage
-----

```sh
deadlinks [-i <ignore>] [-v] [-x <exclude>[...]] [<docroot>[...]]
```

* `-i <ignore>`: file containing links to ignore
* `-v`: print the name of each scanned file to standard error
* `-x <exclude>`: subdirectory of `<docroot>` to exclude (may be repeated)
* `<docroot>`: document root directory to scan for dead links (defaults to the current working directory; may be repeated)

See also
--------

Deadlinks is part of the [Mergician](https://github.com/rcrowley/mergician) suite of tools that manipulate HTML documents:

* [Electrostatic](https://github.com/rcrowley/electrostatic): Mergician-powered, pure-HTML CMS
* [Feed](https://github.com/rcrowley/feed): Scan a document root directory to construct an Atom feed
* [Frag](https://github.com/rcrowley/frag): Extract fragments of HTML documents
* [Sitesearch](https://github.com/rcrowley/sitesearch): Index a document root directory and serve queries to it in AWS Lambda
