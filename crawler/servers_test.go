package crawler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTest(t *testing.T) {
	mux := http.NewServeMux()
	// mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})

	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "user-agent: *\ndisallow: /\n")
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := Crawl(&Config{
		Start:           []string{ts.URL},
		MaxDepth:        10,
		RobotsUserAgent: "Crawler",
	})

	c.Next()
}
