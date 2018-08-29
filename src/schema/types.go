package schema

type SchemaItem struct {
	Description string       `json:"description,omitempty"`
	Mode        string       `json:"mode"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Fields      []SchemaItem `json:"fields,omitempty"`
}
