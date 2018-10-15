package crawler

import (
	"encoding/json"

	"github.com/benjaminestes/crawl/crawler/schema"
)

func SchemaBigQueryJSON() []byte {
	// FIXME: Check error
	j, _ := json.MarshalIndent(schema.BQ, "", "\t")
	return j
}
