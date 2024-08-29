//go:build openclip

package embeddings

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type OpenCLIPImageDataEmbeddingRequest struct {
	Id   int64  `json:"id"`
	Data string `json:"data"`
}

type OpenCLIPEmbeddingRequest struct {
	Content   string                               `json:"content,omitempty"`
	ImageData []*OpenCLIPImageDataEmbeddingRequest `json:"image_data,omitempty"`
}

type OpenCLIPEmbeddingResponse struct {
	Embeddings []float64 `json:"embedding,omitempty"`
}

// OpenCLIPEmbedder implements the `Embedder` interface using an OpenCLIP API endpoint to derive embeddings.
type OpenCLIPEmbedder struct {
	Embedder
	client *http.Client
	host   string
	port   string
	tls    bool
}

func init() {
	ctx := context.Background()
	err := RegisterEmbedder(ctx, "openclip", NewOpenCLIPEmbedder)

	if err != nil {
		panic(err)
	}
}

func NewOpenCLIPEmbedder(ctx context.Context, uri string) (Embedder, error) {

	host := "127.0.0.1"
	port := "5000"
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

	slog.Debug("URL", "host", host, "port", port)

	q := u.Query()

	if q.Has("tls") {

		v, err := strconv.ParseBool("tls")

		if err != nil {
			return nil, fmt.Errorf("Invalid ?tls= parameter, %w", err)
		}

		tls = v
	}

	cl := &http.Client{}

	e := &OpenCLIPEmbedder{
		client: cl,
		host:   host,
		port:   port,
		tls:    tls,
	}

	return e, nil
}

func (e *OpenCLIPEmbedder) Embeddings(ctx context.Context, content string) ([]float64, error) {

	req := &OpenCLIPEmbeddingRequest{
		Content: content,
	}

	rsp, err := e.embeddings(ctx, req)

	if err != nil {
		return nil, err
	}

	return rsp.Embeddings, nil
}

func (e *OpenCLIPEmbedder) Embeddings32(ctx context.Context, content string) ([]float32, error) {

	e64, err := e.Embeddings(ctx, content)

	if err != nil {
		return nil, err
	}

	return asFloat32(e64), nil
}

func (e *OpenCLIPEmbedder) ImageEmbeddings(ctx context.Context, data []byte) ([]float64, error) {

	data_b64 := base64.StdEncoding.EncodeToString(data)

	now := time.Now()
	ts := now.Unix()

	image_req := &OpenCLIPImageDataEmbeddingRequest{
		Data: data_b64,
		Id:   ts,
	}

	req := &OpenCLIPEmbeddingRequest{
		ImageData: []*OpenCLIPImageDataEmbeddingRequest{
			image_req,
		},
	}

	rsp, err := e.embeddings(ctx, req)

	if err != nil {
		return nil, err
	}

	return rsp.Embeddings, nil
}

func (e *OpenCLIPEmbedder) ImageEmbeddings32(ctx context.Context, data []byte) ([]float32, error) {

	e64, err := e.ImageEmbeddings(ctx, data)

	if err != nil {
		return nil, err
	}

	return asFloat32(e64), nil
}

func (e *OpenCLIPEmbedder) embeddings(ctx context.Context, openclip_req *OpenCLIPEmbeddingRequest) (*OpenCLIPEmbeddingResponse, error) {

	u := url.URL{}
	u.Scheme = "http"
	u.Host = fmt.Sprintf("%s:%s", e.host, e.port)

	if len(openclip_req.ImageData) > 0 {
		u.Path = "/embeddings/image"
	} else {
		u.Path = "/embeddings"
	}

	if e.tls {
		u.Scheme = "https"
	}

	endpoint := u.String()

	enc_msg, err := json.Marshal(openclip_req)

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

	var openclip_rsp *OpenCLIPEmbeddingResponse

	dec := json.NewDecoder(rsp.Body)
	err = dec.Decode(&openclip_rsp)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal embeddings, %w", err)
	}

	return openclip_rsp, nil
}

/*

#  https://python.langchain.com/v0.2/docs/integrations/text_embedding/open_clip/
from flask import Flask, request, jsonify
from langchain_experimental.open_clip import OpenCLIPEmbeddings
from PIL import Image
import tempfile
import base64
import os

model="ViT-g-14"
checkpoint="laion2b_s34b_b88k"

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

    with tempfile.NamedTemporaryFile(delete_on_close=False, mode="wb") as wr:

        wr.write(body)
        wr.close()

        embeddings = clip_embd.embed_image([wr.name])
        os.remove(wr.name)

        return jsonify({"embedding": embeddings[0]})

*/
