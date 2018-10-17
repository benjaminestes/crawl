package schema

type schemaItem struct {
	Description string       `json:"description,omitempty"`
	Mode        string       `json:"mode"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Fields      []schemaItem `json:"fields,omitempty"`
}
