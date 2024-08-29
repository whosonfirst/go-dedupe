//go:build llamafile

package embeddings

// https://github.com/Mozilla-Ocho/llamafile/blob/main/llama.cpp/server/README.md#api-endpoints
// https://github.com/Mozilla-Ocho/llamafile?tab=readme-ov-file#other-example-llamafiles
//
// curl --request POST --url http://localhost:8080/embedding --header "Content-Type: application/json" --data '{"content": "Hello world" }'

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	_ "io"
	"net/http"
	"net/url"
	"strings"
	"strconv"
	"time"
)

type LlamafileImageDataEmbeddingRequest struct {
	Id   int64  `json:"id"`
	Data string `json:"data"`
}

type LlamafileEmbeddingRequest struct {
	Content   string                                `json:"content,omitempty"`
	ImageData []*LlamafileImageDataEmbeddingRequest `json:"image_data,omitempty"`
}

type LlamafileEmbeddingResponse struct {
	Embeddings []float64 `json:"embedding,omitempty"`
}

// LlamafileEmbedder implements the `Embedder` interface using an Llamafile API endpoint to derive embeddings.
type LlamafileEmbedder struct {
	Embedder
	client *http.Client
	host   string
	port   string
	tls    bool
}

func init() {
	ctx := context.Background()
	err := RegisterEmbedder(ctx, "llamafile", NewLlamafileEmbedder)

	if err != nil {
		panic(err)
	}
}

func NewLlamafileEmbedder(ctx context.Context, uri string) (Embedder, error) {

	host := "localhost"
	port := "8080"
	tls := false

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	if u.Host != "" {
		host = u.Host

		parts := strings.Split(host, ":")

		if len(parts) < 1 {
			return nil, fmt.Errorf("Failed to parse host component of URI")
		}

		host = parts[0]
	}

	if u.Port() != "" {
		port = u.Port()
	}

	q := u.Query()

	if q.Has("tls") {

		v, err := strconv.ParseBool("tls")

		if err != nil {
			return nil, fmt.Errorf("Invalid ?tls= parameter, %w", err)
		}

		tls = v
	}

	cl := &http.Client{}

	e := &LlamafileEmbedder{
		client: cl,
		host:   host,
		port:   port,
		tls:    tls,
	}

	return e, nil
}

func (e *LlamafileEmbedder) Embeddings(ctx context.Context, content string) ([]float64, error) {

	req := &LlamafileEmbeddingRequest{
		Content: content,
	}

	rsp, err := e.embeddings(ctx, req)

	if err != nil {
		return nil, err
	}

	return rsp.Embeddings, nil
}

func (e *LlamafileEmbedder) Embeddings32(ctx context.Context, content string) ([]float32, error) {

	e64, err := e.Embeddings(ctx, content)

	if err != nil {
		return nil, err
	}

	return asFloat32(e64), nil
}

func (e *LlamafileEmbedder) ImageEmbeddings(ctx context.Context, data []byte) ([]float64, error) {

	data_b64 := base64.StdEncoding.EncodeToString(data)

	now := time.Now()
	ts := now.Unix()

	image_req := &LlamafileImageDataEmbeddingRequest{
		Data: data_b64,
		Id:   ts,
	}

	req := &LlamafileEmbeddingRequest{
		ImageData: []*LlamafileImageDataEmbeddingRequest{
			image_req,
		},
	}

	rsp, err := e.embeddings(ctx, req)

	if err != nil {
		return nil, err
	}

	return rsp.Embeddings, nil
}

func (e *LlamafileEmbedder) ImageEmbeddings32(ctx context.Context, data []byte) ([]float32, error) {

	e64, err := e.ImageEmbeddings(ctx, data)

	if err != nil {
		return nil, err
	}

	return asFloat32(e64), nil
}

func (e *LlamafileEmbedder) embeddings(ctx context.Context, llamafile_req *LlamafileEmbeddingRequest) (*LlamafileEmbeddingResponse, error) {

	u := url.URL{}
	u.Scheme = "http"
	u.Host = fmt.Sprintf("%s:%s", e.host, e.port)
	u.Path = "/embedding"

	if e.tls {
		u.Scheme = "https"
	}

	endpoint := u.String()

	enc_msg, err := json.Marshal(llamafile_req)

	if err != nil {
		return nil, fmt.Errorf("Failed to encode message, %w", err)
	}

	br := bytes.NewReader(enc_msg)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, br)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new request, %w", err)
	}

	req.Header.Set("Content-type", "application/json")

	rsp, err := e.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute request, %w", err)
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Embeddings request failed %d: %s", rsp.StatusCode, rsp.Status)
	}

	// body, _ := io.ReadAll(rsp.Body)
	// fmt.Println("WUT", string(body))

	var llamafile_rsp *LlamafileEmbeddingResponse

	dec := json.NewDecoder(rsp.Body)
	err = dec.Decode(&llamafile_rsp)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal embeddings, %w", err)
	}

	return llamafile_rsp, nil
}
