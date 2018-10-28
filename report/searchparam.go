package report

type SearchParam struct {
	Field      string  `json:"field,omitempty"`
	Type       string  `json:"type,omitempty"`
	Equal      string  `json:"equal,omitempty"`
	UpperLimit float64 `json:"upper_limit,omitempty"`
	LowerLimit float64 `json:"lower_limit,omitempty"`
}
