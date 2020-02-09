package version

import (
	"fmt"
	"runtime"
)

const Version = "v0.2.0"

// UserAgent contains a sane user agent per RFC7231 to use as a default
func UserAgent() string {
	return fmt.Sprintf("Crawl/%s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH)
}
