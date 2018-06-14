package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

var urls map[string]string

func authenticate(r *http.Request) (string, float64, error) {
	// fmt.Println(r.Cookies())
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
			if token != nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					cookie := http.Cookie{Name: "uid", Value: claims["user_id"].(string)}
					r.AddCookie(&cookie)
					exp := claims["exp"].(float64)
					now := float64(time.Now().Unix())
					return jwtToken, exp - now, nil
				} else {
					return "", 0, err
				}
			}
		}
	}
	return "", 0, errors.New("Unable to find JWT Token.")
}

func refresh(tokenString string) (string, error) {
	path := os.Getenv("REFRESH_JWT")
	// "http://localhost:8009/api/v1/userservice/refresh_jwt/"
	resp, err := http.PostForm(path, url.Values{"token": {tokenString}})
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var f map[string]string
	err = json.Unmarshal(body, &f)
	return f["token"], err
}

func direct(r *http.Request) {
	r.URL.Scheme = "http"
}

func modify(r *http.Response) error {
	fmt.Println(r.Request.Cookies())
	fmt.Println(r.Cookies())
	return nil
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("ENV") == "sandbox" {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, POST, PATCH, DELETE")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
			}
		}

		token, exp, auth_err := authenticate(r)
		fmt.Println("Cookies", r.Cookies())
		for suburl, host := range urls {
			if strings.HasPrefix(r.URL.Path, suburl) {
				r.URL.Host = host
				if strings.HasPrefix(r.URL.Path, suburl+"exempt/") {
					next.ServeHTTP(w, r)
				} else {
					if auth_err != nil {
						http.Error(w, "403 Access Forbiddent. Authentication Error."+auth_err.Error(), http.StatusForbidden)
					} else {
						// fmt.Println("Exp", exp)
						if exp >= 1 && exp < 5*60 {
							newToken, token_err := refresh(token)
							// fmt.Println("Refresh JWT", newToken)
							if token_err == nil {
								cookie := &http.Cookie{Name: "JWT", Value: newToken, HttpOnly: false}
								http.SetCookie(w, cookie)
							}
						}
						next.ServeHTTP(w, r)
					}
				}
				return
			}
		}

		if r.URL.Path == "/" {
			if auth_err == nil {
				http.Redirect(w, r, "/web/main/", http.StatusFound)
			} else {
				http.Redirect(w, r, "/web/basic/login/", http.StatusFound)
			}
		} else {
			http.Error(w, "404 URL Not Found: "+r.URL.Path, http.StatusNotFound)
		}
	})
}

func healthcheck(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(rw, "HEALTHY")
}

func main() {
	urls = make(map[string]string)
	for _, env := range os.Environ() {
		variable := strings.Split(env, "=")
		if strings.Contains(variable[0], "SERVICE_TASKIT") {
			suburl := variable[1]
			host := os.Getenv(strings.Replace(variable[0], "SERVICE_TASKIT", "HOST_TASKIT", 1))
			// fmt.Println(suburl, host)
			urls[suburl] = host
		}
	}
	// fmt.Println(urls)
	http.HandleFunc("/healthcheck", healthcheck)
	http.Handle("/", authMiddleware(&httputil.ReverseProxy{Director: direct, ModifyResponse: modify}))
	http.ListenAndServe(":8000", nil)
}
