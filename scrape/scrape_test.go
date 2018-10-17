package scrape

import (
	"os"
	"testing"

	"golang.org/x/net/html"
)

func TestQuery(t *testing.T) {
	f, err := os.Open("testdata/simple.html")
	if err != nil {
		t.Errorf("couldn't open test data")
	}

	doc, err := html.Parse(f)
	if err != nil {
		t.Errorf("couldn't parse test data")
	}

	n := Query("p", map[string]string{
		"name": "best-paragraph",
	}, doc)

	if txt := Text(n); txt != "Match this." {
		t.Errorf(`expected string "Match this.", got %s`, txt)
	}
}
