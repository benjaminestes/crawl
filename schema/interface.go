package schema

import "encoding/json"

func BigQueryJSON() []byte {
	// FIXME: Check error
	j, _ := json.MarshalIndent(bq, "", "\t")
	return j
}
