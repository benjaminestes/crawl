package sitemap

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

type url struct {
	Loc string `xml:"loc"`
}

type urlset struct {
	URLs []url `xml:"url"`
}

type sitemapURL struct {
	Loc string `xml:"loc"`
}

type index struct {
	URLs []sitemapURL `xml:"sitemap"`
}

func loadURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Parse(in io.Reader) ([]string, error) {
	res := &urlset{}

	err := xml.Unmarshal(in, res)
	if err != nil {
		return nil, err
	}

	var urls []string
	for _, u := range res.URLs {
		urls = append(urls, u.Loc)
	}

	return urls, nil
}

func ParseIndex(in io.Reader) ([]string, error) {
	res := &index{}

	err := xml.Unmarshal(in, res)
	if err != nil {
		return nil, err
	}

	var sitemaps []string
	for _, s := range res.URLs {
		sitemaps = append(sitemaps, s.Loc)
	}

	// var urls []string
	// for _, s := range sitemaps {
	// 	newurls, err := FromURL(s)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	urls = append(urls, newurls...)
	//}
	// return urls, nil
}

func Fetch(url string) ([]string, error) {
}

func FetchIndex(url string) ([]string, error) {
}

func FetchAll(url string) ([]string, error) {
	data, err := loadURL(url)
	if err != nil {
		return nil, err
	}

	var urls []string

	urls, err = Parse(data)
	if err != nil {
		return nil, err
	}

	if len(urls) > 0 {
		return urls, nil
	}

	sitemaps, err = ParseIndex(data)
	if err != nil {
		return nil, err
	}

	for _, s := range sitemaps {
		newurls, err := FetchAll(s)
		if err != nil {
			return nil, err
		}
		urls = append(urls, 

	return urls, nil
}
