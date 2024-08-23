# Embeddings

## embeddings.Embedder

```
// Embedder defines an interface for generating (vector) embeddings
type Embedder interface {
	// Embeddings returns the embeddings for a string as a list of float64 values.
	Embeddings(context.Context, string) ([]float64, error)
	// Embeddings32 returns the embeddings for a string as a list of float32 values.
	Embeddings32(context.Context, string) ([]float32, error)
	// ImageEmbeddings returns the embeddings for a base64-encoded image as a list of float64 values.	
	ImageEmbeddings(context.Context, string) ([]float64, error)
	// ImageEmbeddings32 returns the embeddings for a base64-encoded image as a list of float32 values.		
	ImageEmbeddings32(context.Context, string) ([]float32, error)
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
e, _ := embeddings.NewEmbedder(ctx, "chromemollama://?{PARAMETERS")
```

Valid parameters for the `ChromemOllamaEmbedder` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| model | string| yes | The name of the model you want to Ollama API to use when generating embeddings. |

Use of the `ChromemOllamaEmbedder` implementation requires tools be built with the `-chromem` tag.

#### LlamafileEmbedder

The `LlamafileEmbedder` implementation uses the [llamafile application's REST API](https://github.com/Mozilla-Ocho/llamafile/blob/main/llama.cpp/server/README.md#api-endpoints) to generate embeddings for a text. This package assumes that the llamafile application has already installed, is running and set up to use the models necessary to generate embeddings. Please consult the [llamafile documentation](https://github.com/Mozilla-Ocho/llamafile/tree/main) for details.

The syntax for creating a new `LlamafileEmbedder` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/embeddings"
)

ctx := context.Background()
e, _ := embeddings.NewEmbedder(ctx, "llamafile://{HOST}:{PORT}?{PARAMETERS")
```

Where `{HOST}` and `{PORT}` are the hostname and port the llamafile API is listening for requests on. Defaults, if omitted, are "localhost" and "8080", respectively.

Valid parameters for the `LlamafileEmbedder` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| tls | boolean| no | A boolean flag signaling that requests to the llamafile API should be made using a secure connection. Default is false. |

Use of the `LlamafileEmbedder` implementation requires tools be built with the `-llamafile` tag.

#### OllamaEmbedder

The `OllamaEmbedder` implementation uses the [Ollama application's REST API](https://github.com/ollama/ollama?tab=readme-ov-file#rest-api) to generate embeddings for a text. This package assumes that the Ollama application has already installed, is running and set up to use the models necessary to generate embeddings. Please consult the [Ollama documentation](https://github.com/ollama/ollama) for details.

The syntax for creating a new `OllamaEmbedder` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/embeddings"
)

ctx := context.Background()
e, _ := embeddings.NewEmbedder(ctx, "ollama://?{PARAMETERS")
```

Valid parameters for the `OllamaEmbedder` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| model | string| yes | The name of the model you want to Ollama API to use when generating embeddings. |

Use of the `OllamaEmbedder` implementation requires tools be built with the `-ollama` tag.