package database

type QueryResult struct {
	ID         string
	Metadata   map[string]string
	Embedding  []float32
	Content    string
	Similarity float32
}
