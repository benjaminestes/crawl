package sitemap

import (
	"fmt"
	"testing"
)

func TestFromURL(t *testing.T) {
	urls, err := Fetch("https://www.distilled.net/sitemap.xml")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(len(urls))
}
