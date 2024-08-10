package export

type Feature struct {
	Type       string      `json:"type"`
	Id         int64       `json:"id"`
	Properties interface{} `json:"properties"`
	Bbox       []float64   `json:"bbox,omitempty"`
	Geometry   interface{} `json:"geometry"`
}
