package crawler

import (
	"encoding/json"

	"github.com/benjaminestes/crawl/crawler/schema"
)

func Schema() []byte {
	// FIXME: Check error
	j, _ := json.MarshalIndent(schema.BQ, "", "\t")
	return j
}
