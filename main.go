package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"

	"errors"
	"strings"
)

func authenticate(r *http.Request) error {
	fmt.Println(r.Cookies())
	for _, cookie := range r.Cookies() {
		if cookie.Name == "JWT" {
			jwtToken := cookie.Value
			token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
				pubkey, _ := ioutil.ReadFile("secrets/jwtRS256.key.pub")

				block, _ := pem.Decode(pubkey)
				key, _ := x509.ParsePKIXPublicKey(block.Bytes)
				return key, nil
			})
			// fmt.Println("token", token, err)
			if token != nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					// fmt.Println()
					cookie := http.Cookie{Name: "uid", Value: claims["user_id"].(string)}
					r.AddCookie(&cookie)
					fmt.Println(r.Cookies())
					return nil
				} else {
					fmt.Println(err)
					return err
				}
			}
		}
	}
	return errors.New("Unable to find JWT Token.")
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
						err := authenticate(r)
						if err != nil {
							fmt.Println(err)
						}
					}
					fmt.Println("after", r.URL.Path, host)
					break
				}
			}
		}})
	http.ListenAndServe(":7999", nil)
}
