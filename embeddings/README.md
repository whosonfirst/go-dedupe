# Embeddings

## embeddings.Embedder

```
// Embedder defines an interface for generating (vector) embeddings
type Embedder interface {
	// Embeddings returns the embeddings for a string as a list of float64 values.
	Embeddings(context.Context, string) ([]float64, error)
	// Embeddings32 returns the embeddings for a string as a list of float32 values.	
	Embeddings32(context.Context, string) ([]float32, error)
}
```

### Implementations

#### ChromemOllamaEmbedder

#### OllamaEmbedder
