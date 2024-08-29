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

#### OpenCLIPEmbedder

The `OpenCLIPEmbedder` implementation derives embeddings from an HTTP server that processes `POST` requests. This package assumes that the server has been installed and is already running. A sample HTTP server (Flask) implementation is included below:

The syntax for creating a new `OpenCLIPEmbedder` is:

```
import (
	"context"
	
	"github.com/whosonfirst/go-dedupe/embeddings"
)

ctx := context.Background()
e, _ := embeddings.NewEmbedder(ctx, "openclip://{HOST}:{PORT}?{PARAMETERS")
```

Where `{HOST}` and `{PORT}` are the hostname and port the llamafile API is listening for requests on. Defaults, if omitted, are "127.0.0.1" and "5000", respectively.

Valid parameters for the `OpenCLIPEmbedder` implemetation are:

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| tls | boolean| no | A boolean flag signaling that requests to the llamafile API should be made using a secure connection. Default is false. |

Use of the `OpenCLIPEmbedder` implementation requires tools be built with the `-openclip` tag.

##### OpenCLIP server

Here is an example of a very simple Flask server to process embedding requests locally. Imagine this is a file called `openclip_server.py`:

```
from flask import Flask, request, jsonify
from langchain_experimental.open_clip import OpenCLIPEmbeddings
from PIL import Image
import tempfile
import base64
import os

model="ViT-g-14"
checkpoint="laion2b_s34b_b88k"

# For smaller, memory-constrained devices try something like:
# model="ViT-B-32"
# checkpoint="laion2b_s34b_b79k"

clip_embd = OpenCLIPEmbeddings(model_name=model, checkpoint=checkpoint)

app = Flask(__name__)

@app.route("/embeddings", methods=['POST'])
def embeddings():
    req = request.json
    embeddings = clip_embd.embed_documents([ req["data"] ])
    return jsonify({"embedding": embeddings[0]})

@app.route("/embeddings/image", methods=['POST'])
def embeddings_image():

    req = request.json
    body = base64.b64decode(req["image_data"][0]["data"])

    # Note that `delete_on_close=False` is a Python 3.12-ism
    # Use `delete=False` for Python 3.11
    
    with tempfile.NamedTemporaryFile(delete_on_close=False, mode="wb") as wr:

        wr.write(body)
        wr.close()

        embeddings = clip_embd.embed_image([wr.name])
        os.remove(wr.name)

        return jsonify({"embedding": embeddings[0]})
```

To start the server:

```
$> ./bin/flask --app openclip_server run
/usr/local/src/lancedb/lib/python3.12/site-packages/open_clip/factory.py:129: FutureWarning: You are using `torch.load` with `weights_only=False` (the current default value), which uses the default pickle module implicitly. It is possible to construct malicious pickle data which will execute arbitrary code during unpickling (See https://github.com/pytorch/pytorch/blob/main/SECURITY.md#untrusted-models for more details). In a future release, the default value for `weights_only` will be flipped to `True`. This limits the functions that could be executed during unpickling. Arbitrary objects will no longer be allowed to be loaded via this mode unless they are explicitly allowlisted by the user via `torch.serialization.add_safe_globals`. We recommend you start setting `weights_only=True` for any use case where you don't have full control of the loaded file. Please open an issue on GitHub for any issues related to this experimental feature.
  checkpoint = torch.load(checkpoint_path, map_location=map_location)
 * Serving Flask app 'openclip_server'
 * Debug mode: off
INFO:werkzeug:WARNING: This is a development server. Do not use it in a production deployment. Use a production WSGI server instead.
 * Running on http://127.0.0.1:5000
```