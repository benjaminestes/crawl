package sitemap

import (
	"fmt"
	"testing"
)

func TestFromURL(t *testing.T) {
	urls, err := FromURL("https://www.payscale.com/sitemap-index.xml")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(len(urls))
}
