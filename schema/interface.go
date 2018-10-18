// Package scheme is an internal package of the tool Crawl,
// responsible for automatically generating schema definitions
// for output files.
package schema

import "encoding/json"

func BigQueryJSON() []byte {
	// FIXME: Check error
	j, _ := json.MarshalIndent(bq, "", "\t")
	return j
}
