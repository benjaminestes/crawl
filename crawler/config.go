package crawler

import "net/url"

type Config struct {
	Connections     int
	UserAgent       string
	RobotsUserAgent string
	Include         []string
	Exclude         []string
	Start           []string
	RespectNofollow bool
	MaxDepth        int
	WaitTime        string
}

func (conf *Config) initialQueue() ([]resolvedURL, error) {
	var result []resolvedURL
	for _, s := range conf.Start {
		u, err := url.Parse(s)
		if err != nil {
			return nil, err
		}
		result = append(result, resolvedURL(u.String()))
	}
	return result, nil
}
