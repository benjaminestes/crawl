package sitemap

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFetch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("testdata/sitemap.xml")
		if err != nil {
			t.Errorf("couldn't open sitemap.xml for reading")
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("couldn't read sitemap.xml")
		}
		fmt.Fprintf(w, "%s", data)
	}))

	urls, err := Fetch(ts.URL)
	if err != nil {
		t.Errorf("couldn't retrieve test sitemap")
	}

	if len(urls) != 1 {
		t.Errorf("expected 1 URL, got %d", len(urls))
	}
}

func TestFetchIndex(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("testdata/sitemap-index.xml")
		if err != nil {
			t.Errorf("couldn't open sitemap-index.xml for reading")
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("couldn't read sitemap-index.xml")
		}
		fmt.Fprintf(w, "%s", data)
	}))

	urls, err := FetchIndex(ts.URL)
	if err != nil {
		t.Errorf("couldn't retrieve test sitemap index")
	}

	if len(urls) != 2 {
		t.Errorf("expected 2 URLs, got %d", len(urls))
	}
}

func TestInvalidData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "This is not a sitemap(index)?!")
	}))

	_, err := Fetch(ts.URL)
	if err == nil {
		t.Errorf("Fetch should've reported an error")
	}

	_, err = FetchIndex(ts.URL)
	if err == nil {
		t.Errorf("FetchIndex should've reported an error")
	}
}
