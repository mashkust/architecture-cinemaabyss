package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func newProxy(rawURL string) *httputil.ReverseProxy {
	target, err := url.Parse(rawURL)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director

	proxy.Director = func(r *http.Request) {
		originalDirector(r)
		r.Host = target.Host
	}

	return proxy
}

func main() {
	rand.Seed(time.Now().UnixNano())

	port := os.Getenv("PORT")
	monolithURL := os.Getenv("MONOLITH_URL")
	moviesURL := os.Getenv("MOVIES_SERVICE_URL")
	eventsURL := os.Getenv("EVENTS_SERVICE_URL")

	gradualMigration := os.Getenv("GRADUAL_MIGRATION") == "true"
	migrationPercent, _ := strconv.Atoi(os.Getenv("MOVIES_MIGRATION_PERCENT"))

	monolithProxy := newProxy(monolithURL)
	moviesProxy := newProxy(moviesURL)
	eventsProxy := newProxy(eventsURL)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":true}`))
			return
		}

		if strings.HasPrefix(path, "/api/events") {
			eventsProxy.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(path, "/api/movies") {
			if !gradualMigration {
				moviesProxy.ServeHTTP(w, r)
				return
			}

			if rand.Intn(100) < migrationPercent {
				moviesProxy.ServeHTTP(w, r)
				return
			}

			monolithProxy.ServeHTTP(w, r)
			return
		}

		monolithProxy.ServeHTTP(w, r)
	})

	log.Printf("proxy started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
