package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

func exempt_route(urls map[string]string) {
	for suburl, host := range urls {
		http.Handle(suburl+"exempt/", &httputil.ReverseProxy{
			Director: func(r *http.Request) {
				fmt.Println("exempt", r.URL.Path)
				r.URL.Scheme = "http"
				r.URL.Host = host
				r.URL.Path = strings.Replace(r.URL.Path, suburl, "/", 1)
				fmt.Println("after", r.URL.Path)
			}})
	}
}

func auth_route(urls map[string]string) {
	for suburl, host := range urls {
		http.Handle(suburl, &httputil.ReverseProxy{
			Director: func(r *http.Request) {
				fmt.Println("auth", r.URL.Path)
				r.URL.Scheme = "http"
				r.URL.Host = host
				r.URL.Path = strings.Replace(r.URL.Path, suburl, "/", 1)
				fmt.Println("after", r.URL.Path)
			}})
	}
}

func default_handler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, req.URL.Path)
}

func main() {
	urls := make(map[string]string)
	for _, env := range os.Environ() {
		variable := strings.Split(env, "=")
		if strings.Contains(variable[0], "SERVICE") {
			suburl := variable[1]
			host := os.Getenv(strings.Replace(variable[0], "SERVICE", "HOST", 1))
			fmt.Println(suburl, host)
			urls[suburl] = host
		}
	}
	fmt.Println(urls)
	exempt_route(urls)
	auth_route(urls)
	http.HandleFunc("/", default_handler)
	http.ListenAndServe(":8000", nil)
}
