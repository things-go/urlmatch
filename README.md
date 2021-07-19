# urlmatch

urlmatch is a trie match url. Copy from [httprouter](https://github.com/julienschmidt/httprouter) but just for match url.

[![GoDoc](https://godoc.org/github.com/things-go/urlmatch?status.svg)](https://godoc.org/github.com/things-go/urlmatch)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/things-go/urlmatch?tab=doc)
[![Build Status](https://www.travis-ci.com/things-go/urlmatch.svg?branch=main)](https://www.travis-ci.com/things-go/urlmatch)
[![codecov](https://codecov.io/gh/things-go/urlmatch/branch/main/graph/badge.svg)](https://codecov.io/gh/things-go/urlmatch)
![Action Status](https://github.com/things-go/urlmatch/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/things-go/urlmatch)](https://goreportcard.com/report/github.com/things-go/urlmatch)
[![License](https://img.shields.io/github/license/things-go/urlmatch)](https://github.com/things-go/urlmatch/raw/main/LICENSE)
[![Tag](https://img.shields.io/github/v/tag/things-go/urlmatch)](https://github.com/things-go/urlmatch/tags)

## Features

**Only explicit matches: a requested URL path could match multiple patterns. Therefore they have some awkward pattern priority rules, like *longest match* or *first registered, first matched*. By design of this router, a request can only match exactly one or no route. As a result, there are also no unintended matches, which makes it great for SEO and improves the user experience.

**Stop caring about trailing slashes:** Choose the URL style you like, the router automatically redirects the client if a trailing slash is missing or if there is one extra. Of course it only does so, if the new path has a handler. If you don't like it, you can [turn off this behavior](https://pkg.go.dev/github.com/things-go/urlmatch#Router.RedirectTrailingSlash).

**Path auto-correction:** Besides detecting the missing or additional trailing slash at no extra cost, the router can also fix wrong cases and remove superfluous path elements (like `../` or `//`). Is [CAPTAIN CAPS LOCK](http://www.urbandictionary.com/define.php?term=Captain+Caps+Lock) one of your users? HttpRouter can help him by making a case-insensitive look-up and redirecting him to the correct URL.

**Parameters in your routing pattern:** Stop parsing the requested URL path, just give the path segment a name and the router delivers the dynamic value to you. Because of the design of the router, path parameters are very cheap.


## Usage

This is just a quick introduction, view the [Go.Dev](https://pkg.go.dev/github.com/things-go/urlmatch?tab=doc) for details.

Let's start with a trivial example:

```go
package main

import (
	"log"
	"net/http"

	"github.com/things-go/urlmatch"
)

func main() {
	router := urlmatch.New()
	router.GET("/", "/")
	router.GET("/hello/:name", "Hello")

	v, _, matched := router.Match(http.MethodGet, "/")
	if matched {
		log.Println(v)
	}
	v, ps, matched := router.Match(http.MethodGet, "/hello/myname")
	if matched {
		log.Println(v)
		log.Println(ps.Param("name"))
	}
}

```

### Named parameters

As you can see, `:name` is a *named parameter*. The values are accessible via `httprouter.Params`, which is just a slice of `httprouter.Param`s. You can get the value of a parameter either by its index in the slice, or by using the `Param(name)` method: `:name` can be retrieved by `Param("name")`.

Named parameters only match a single path segment:

```
Pattern: /user/:user

 /user/gordon              match
 /user/you                 match
 /user/gordon/profile      no match
 /user/                    no match
```

**Note:** Since this router has only explicit matches, you can not register static routes and parameters for the same path segment. For example you can not register the patterns `/user/new` and `/user/:user` for the same request method at the same time. The routing of different request methods is independent from each other.

### Catch-All parameters

The second type are *catch-all* parameters and have the form `*name`. Like the name suggests, they match everything. Therefore they must always be at the **end** of the pattern:

```
Pattern: /src/*filepath

 /src/                     match
 /src/somefile.go          match
 /src/subdir/somefile.go   match
```

## How does it work?

The router relies on a tree structure which makes heavy use of *common prefixes*, it is basically a *compact* [*prefix tree*](https://en.wikipedia.org/wiki/Trie) (or just [*Radix tree*](https://en.wikipedia.org/wiki/Radix_tree)). Nodes with a common prefix also share a common parent. Here is a short example what the routing tree for the `GET` request method could look like:

```
Priority   Path             Value
9          \                *<1>
3          ├s               nil
2          |├earch\         *<2>
1          |└upport\        *<3>
2          ├blog\           *<4>
1          |    └:post      nil
1          |         └\     *<5>
2          ├about-us\       *<6>
1          |        └team\  *<7>
1          └contact\        *<8>
```

Every `*<num>` represents the memory address of a handler function (a pointer). If you follow a path trough the tree from the root to the leaf, you get the complete route path, e.g `\blog\:post\`, where `:post` is just a placeholder ([*parameter*](#named-parameters)) for an actual post name. Unlike hash-maps, a tree structure also allows us to use dynamic parts like the `:post` parameter, since we actually match against the routing patterns instead of just comparing hashes. 

Since URL paths have a hierarchical structure and make use only of a limited set of characters (byte values), it is very likely that there are a lot of common prefixes. This allows us to easily reduce the routing into ever smaller problems. Moreover the router manages a separate tree for every request method. For one thing it is more space efficient than holding a method->value map in every single node, it also allows us to greatly reduce the routing problem before even starting the look-up in the prefix-tree.

For even better scalability, the child nodes on each tree level are ordered by priority, where the priority is just the number of handles registered in sub nodes (children, grandchildren, and so on..). This helps in two ways:

1. Nodes which are part of the most routing paths are evaluated first. This helps to make as much routes as possible to be reachable as fast as possible.
2. It is some sort of cost compensation. The longest reachable path (highest cost) can always be evaluated first. The following scheme visualizes the tree structure. Nodes are evaluated from top to bottom and from left to right.

```
├------------
├---------
├-----
├----
├--
├--
└-
```