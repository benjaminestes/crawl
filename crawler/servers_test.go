package crawler

import (
	"fmt"
	"html/template"
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

	c := &Crawler{
		From:            []string{ts.URL},
		MaxDepth:        10,
		RobotsUserAgent: "Crawler",
		Connections:     2,
		WaitTime:        "100ms",
	}

	err := c.Start()
	if err != nil {
		t.Errorf("%v", err)
	}

	c.Next()
}

func pow(x int, n int) int {
	switch {
	case n < 0:
		return 0
	case n == 0:
		return 1
	case n == 1:
		return x
	default:
		return x * pow(x, n-1)
	}
}

func expectedCount(d int) int {
	var total int
	for d > 0 {
		total += pow(5, d)
		d--
	}
	return total + 1
}

func TestTestTwo(t *testing.T) {
	var children = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	type Page struct {
		ID       string
		Children []int
	}

	funcMap := template.FuncMap{
		"odd": func(i int) bool {
			return i%2 == 1
		},
	}

	tmpl := template.Must(template.New("nice_page.html").Funcs(funcMap).ParseFiles("testdata/nice_page.html"))

	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "user-agent: *\nallow: /\n")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		p := Page{
			ID:       req.URL.Path[1:],
			Children: children,
		}
		w.Header().Set("Content-Type", "text/html")
		tmpl.Execute(w, p)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := &Crawler{
		From:            []string{ts.URL},
		MaxDepth:        3,
		RobotsUserAgent: "Crawler",
		Connections:     20,
		RespectNofollow: true,
		WaitTime:        "1ms",
	}

	err := c.Start()
	if err != nil {
		t.Errorf("%v", err)
	}

	var count int
	for n := c.Next(); n != nil; n = c.Next() {
		count++
	}

	wantCount := expectedCount(c.MaxDepth)
	if count != wantCount {
		t.Errorf("expected %d URLs, returned %d", wantCount, count)
	}
}
