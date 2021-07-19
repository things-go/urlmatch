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
