package gowebtable

//Field defines the struct for a Field element which stores values relating to a table field
type Field struct {
	Name   string       `json:"name"`
	Header string       `json:"header"`
	Type   string       `json:"type"`
	Hide   bool         `json:"-"`
	Unique []FieldValue `json:"-"`
}

//FieldValue defines the struct for a FieldValue element which stores values used in the field filter boxes
type FieldValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}
