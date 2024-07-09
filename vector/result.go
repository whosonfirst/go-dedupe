package vector

type QueryResult struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	// Name    string `json:"name"`
	// Address string `json:"address"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Embedding  []float32         `json:"embeddings,omitempty"`
	Similarity float32           `json:"similarity"`
}

func (r *QueryResult) String() string {
	return r.Content
}
