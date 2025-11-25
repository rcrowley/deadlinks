Deadlinks
=========

Deadlinks scans all the HTML documents in a document root directory for dead links. It's intended to be used almost like spellcheck in a CI process or pre-commit hook.

Installation
------------

```sh
go install github.com/rcrowley/deadlinks
```

Usage
-----

```sh
deadlinks [-i <ignore>] [-v] [<docroot>[...]]
```

* `-i <ignore>`: file containing links to ignore
* `<docroot>`: document root directory to scan for dead links (defaults to the current working directory)

See also
--------

Deadlinks is part of the [Mergician](https://github.com/rcrowley/mergician) suite of tools that manipulate HTML documents:

* [Electrostatic](https://github.com/rcrowley/electrostatic): Mergician-powered, pure-HTML CMS
* [Frag](https://github.com/rcrowley/frag): Extract fragments of HTML documents
* [Sitesearch](https://github.com/rcrowley/sitesearch): Index a document root directory and serve queries to it in AWS Lambda
