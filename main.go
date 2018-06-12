package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

func authenticate(r *http.Request) {

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
	http.Handle("/", &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			fmt.Println("base url", r.URL.Path)
			r.URL.Scheme = "http"
			for suburl, host := range urls {
				if strings.HasPrefix(r.URL.Path, suburl) {
					fmt.Println("auth", r.URL.Path, suburl)
					r.URL.Host = host
					if strings.HasPrefix(r.URL.Path, suburl+"exempt/") {
						fmt.Println("Exempting request")
					} else {
						fmt.Println("Auth request")
					}
					fmt.Println("after", r.URL.Path, host)
					break
				}
			}
		}})
	http.ListenAndServe(":8000", nil)
}
