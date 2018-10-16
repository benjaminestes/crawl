package crawler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"
)

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

var defaultConfig = Config{
	Connections:     1,
	MaxDepth:        0,
	RobotsUserAgent: "Crawler",
	WaitTime:        "100ms",
}

func ConfigFromJSON(in io.Reader) (*Config, error) {
	config := defaultConfig

	configJSON, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configJSON, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (conf *Config) connections() int {
	if conf.Connections < 1 {
		return 1
	}
	return conf.Connections
}

func (conf *Config) initialQueue() ([]resolvedURL, error) {
	var result []resolvedURL
	for _, s := range conf.Start {
		u, err := url.Parse(s)
		if err != nil {
			return nil, err
		}
		// Per RFC 1945, a request without a path part must
		// send a "/".
		if u.Path == "" {
			u.Path = "/"
		}
		result = append(result, resolvedURL(u.String()))
	}
	return result, nil
}
