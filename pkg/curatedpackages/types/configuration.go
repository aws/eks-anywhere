package types

type Schema struct {
	ID         string               `json:"$id,omitempty"`
	Schema     string               `json:"$schema,omitempty"`
	Title      string               `json:"title,omitempty"`
	Type       string               `json:"type,omitempty"`
	Required   []string             `json:"required,omitempty"`
	Properties map[string]*Property `json:"properties,omitempty"`
}

type Property struct {
	Type        string `json:"type,omitempty"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}
