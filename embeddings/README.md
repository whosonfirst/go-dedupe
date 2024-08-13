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

The `ChromemOllamaEmbedder` implementation uses the [philippgille/chromem-go](https://github.com/philippgille/chromem-go) package to generate embeddings. In turn `chromem-go` uses the [Ollama application's REST API](https://github.com/ollama/ollama?tab=readme-ov-file#rest-api) to generate embeddings for a text. This package assumes that the Ollama application has already installed, is running and set up to use the models necessary to generate embeddings. Please consult the [Ollama documentation](https://github.com/ollama/ollama) for details.

The syntax for creating a new `ChromemOllamaEmbedder` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/embeddings"
)

ctx := context.Background()
, _ := embeddings.NewEmbedder(ctx, "chromemollama://?{PARAMETERS")
```

Valid parameters for the `ChromemOllamaEmbedder` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| model | string| yes | The name of the model you want to Ollama API to use when generating embeddings. |

#### OllamaEmbedder

The `OllamaEmbedder` implementation uses the [Ollama application's REST API](https://github.com/ollama/ollama?tab=readme-ov-file#rest-api) to generate embeddings for a text. This package assumes that the Ollama application has already installed, is running and set up to use the models necessary to generate embeddings. Please consult the [Ollama documentation](https://github.com/ollama/ollama) for details.

The syntax for creating a new `OllamaEmbedder` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/embeddings"
)

ctx := context.Background()
, _ := embeddings.NewEmbedder(ctx, "ollama://?{PARAMETERS")
```

Valid parameters for the `OllamaEmbedder` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| model | string| yes | The name of the model you want to Ollama API to use when generating embeddings. |