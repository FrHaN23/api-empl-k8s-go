package types

type Res struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Total   int    `json:"total,omitempty"`
}
