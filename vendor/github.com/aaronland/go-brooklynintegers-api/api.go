package api

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/aaronland/go-artisanal-integers/client"
	"github.com/cenkalti/backoff/v4"
	"github.com/tidwall/gjson"
	"go.uber.org/ratelimit"
)

// In principle this could also be done with a sync.OnceFunc call but that will
// require that everyone uses Go 1.21 (whose package import changes broke everything)
// which is literally days old as I write this. So maybe a few releases after 1.21.

var register_mu = new(sync.RWMutex)
var register_map = map[string]bool{}

func init() {

	ctx := context.Background()
	err := RegisterClientSchemes(ctx)

	if err != nil {
		panic(err)
	}
}

// RegisterClientSchemes will explicitly register all the schemes associated with the `client.Client` interface.
func RegisterClientSchemes(ctx context.Context) error {

	roster := map[string]client.ClientInitializeFunc{
		"brooklynintegers": NewAPIClient,
	}

	register_mu.Lock()
	defer register_mu.Unlock()

	for scheme, fn := range roster {

		_, exists := register_map[scheme]

		if exists {
			continue
		}

		err := client.RegisterClient(ctx, scheme, fn)

		if err != nil {
			return fmt.Errorf("Failed to register client for '%s', %w", scheme, err)
		}

		register_map[scheme] = true
	}

	return nil
}

type APIClient struct {
	client.Client
	isa          string
	http_client  *http.Client
	Scheme       string
	Host         string
	Endpoint     string
	rate_limiter ratelimit.Limiter
}

type APIError struct {
	Code    int64
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

type APIResponse struct {
	raw []byte
}

func (rsp *APIResponse) Int() (int64, error) {

	ints := gjson.GetBytes(rsp.raw, "integers.0.integer")

	if !ints.Exists() {
		return -1, fmt.Errorf("Failed to generate any integers")
	}

	i := ints.Int()
	return i, nil
}

func (rsp *APIResponse) Stat() string {

	r := gjson.GetBytes(rsp.raw, "stat")

	if !r.Exists() {
		return ""
	}

	return r.String()
}

func (rsp *APIResponse) Ok() (bool, error) {

	stat := rsp.Stat()

	if stat == "ok" {
		return true, nil
	}

	return false, rsp.Error()
}

func (rsp *APIResponse) Error() error {

	c := gjson.GetBytes(rsp.raw, "error.code")
	m := gjson.GetBytes(rsp.raw, "error.message")

	if !c.Exists() {
		return fmt.Errorf("Failed to parse error code")
	}

	if !m.Exists() {
		return fmt.Errorf("Failed to parse error message")
	}

	err := APIError{
		Code:    c.Int(),
		Message: m.String(),
	}

	return &err
}

func NewAPIClient(ctx context.Context, uri string) (client.Client, error) {

	http_client := &http.Client{}
	rl := ratelimit.New(10) // please make this configurable

	cl := &APIClient{
		Scheme:       "https",
		Host:         "api.brooklynintegers.com",
		Endpoint:     "rest/",
		http_client:  http_client,
		rate_limiter: rl,
	}

	return cl, nil
}

func (client *APIClient) NextInt(ctx context.Context) (int64, error) {

	params := url.Values{}
	method := "brooklyn.integers.create"

	var next_id int64

	cb := func() error {

		rsp, err := client.executeMethod(ctx, method, &params)

		if err != nil {
			return err
		}

		i, err := rsp.Int()

		if err != nil {
			log.Println(err)
			return err
		}

		next_id = i
		return nil
	}

	bo := backoff.NewExponentialBackOff()

	err := backoff.Retry(cb, bo)

	if err != nil {
		return -1, fmt.Errorf("Failed to execute method (%s), %w", method, err)
	}

	return next_id, nil
}

func (client *APIClient) executeMethod(ctx context.Context, method string, params *url.Values) (*APIResponse, error) {

	client.rate_limiter.Take()

	url := client.Scheme + "://" + client.Host + "/" + client.Endpoint

	params.Set("method", method)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)

	if err != nil {
		return nil, fmt.Errorf("Failed to create request (%s), %w", url, err)
	}

	req.URL.RawQuery = (*params).Encode()

	req.Header.Add("Accept-Encoding", "gzip")

	rsp, err := client.http_client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to create request (%s), %w", url, err)
	}

	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)

	if err != nil {
		return nil, fmt.Errorf("Failed to read response, %w", err)
	}

	r := APIResponse{
		raw: body,
	}

	return &r, nil
}
