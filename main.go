package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

// func exempt_route(urls map[string]string) {
// 	for suburl, _ := range urls {
// 		http.Handle(suburl+"exempt/", &httputil.ReverseProxy{
// 			Director: func(r *http.Request) {
// 				fmt.Println("exempt", r.URL.Path, suburl)
// 				r.URL.Scheme = "http"
// 				host := urls[suburl]
// 				r.URL.Host = host
// 				r.URL.Path = strings.Replace(r.URL.Path, suburl, "/", 1)
// 				fmt.Println("after", r.URL.Path, host)
// 			}})
// 	}
// }

// func auth_route(urls map[string]string) {
// 	for suburl, host := range urls {
// 		fmt.Println("URL", suburl, host)
// 		http.Handle(suburl, &httputil.ReverseProxy{
// 			Director: func(r *http.Request) {
// 				fmt.Println("auth", r.URL.Path, suburl)
// 				r.URL.Scheme = "http"
// 				host := urls[suburl]
// 				r.URL.Host = host
// 				r.URL.Path = strings.Replace(r.URL.Path, suburl, "/", 1)
// 				fmt.Println("after", r.URL.Path, host)
// 			}})
// 	}
// }

func default_handler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, req.URL.Path)
}

func handler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, req.URL.Path)
}

func main() {
	urls := make(map[string]string)
	for _, env := range os.Environ() {
		variable := strings.Split(env, "=")
		if strings.Contains(variable[0], "SERVICE_TASKIT") {
			suburl := variable[1]
			host := os.Getenv(strings.Replace(variable[0], "SERVICE_TASKIT", "HOST_TASKIT", 1))
			fmt.Println(suburl, host)
			urls[suburl] = host
		}
	}
	fmt.Println(urls)
	// exempt_route(urls)
	// auth_route(urls)
	http.Handle("/api", &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			for suburl, host := range urls {
				if strings.HasPrefix(r.URL.Path, suburl) {
					fmt.Println("auth", r.URL.Path, suburl)
					r.URL.Host = host
					r.URL.Path = strings.Replace(r.URL.Path, suburl, "/", 1)
					fmt.Println("after", r.URL.Path, host)
					break
				}
			}
		}})
	http.HandleFunc("/", default_handler)
	http.ListenAndServe(":8000", nil)
}
