package version

import (
	"fmt"
	"runtime"
)

// Version contains the crawl version number
var Version = "0.1.1"

// UserAgent contains a sane user agent per RFC7231 to use a a default
var UserAgent = getUserAgent()

func getUserAgent() string {
	return fmt.Sprintf("Crawl/%s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH)
}
