package client

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aaronland/go-aws-auth"
	"github.com/cenkalti/backoff/v4"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchtransport"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/awsv2"
)

// ClientOptions is a struct definining properties used to create a new `opensearch.Client` instance
// using the `NewClient` method.
type ClientOptions struct {
	// A list of valid Opensearch endpoint URIs
	Addresses []string
	Index     string
	// Disable TLS verification checks
	Insecure bool
	// A valid Opensearch username
	Username string
	// A valid Opensearch password
	Password string
	// AWSCredentialsURI is a valid `aaronland/go-aws-auth` URI
	AWSCredentialsURI string
	// Enable debugging for Opensearch requests
	Debug bool
}

func ClientOptionsFromURI(ctx context.Context, uri string) (*ClientOptions, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	var opensearch_endpoint string

	port := u.Port()

	// TO DO: update to support multiple addresses and/or the fact that TLS may be
	// enabled on a non-443 port

	switch port {
	case "443":
		opensearch_endpoint = fmt.Sprintf("https://%s", u.Host)
	default:
		opensearch_endpoint = fmt.Sprintf("http://%s", u.Host)
	}

	opensearch_index := strings.TrimLeft(u.Path, "/")

	debug := false

	q := u.Query()

	q_debug := q.Get("debug")
	q_insecure := q.Get("insecure")
	q_tls := q.Get("require-tls")
	q_username := q.Get("username")
	q_password := q.Get("password") // update to use go-runtime
	q_aws_credentials_uri := q.Get("aws-credentials-uri")

	if q_tls != "" {

		v, err := strconv.ParseBool(q_tls)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?require-tls= parameter")
		}

		if v {
			opensearch_endpoint = fmt.Sprintf("https://%s", u.Host)
		}
	}

	os_client_opts := &ClientOptions{
		Addresses:         []string{opensearch_endpoint},
		Index:             opensearch_index,
		Username:          q_username,
		Password:          q_password,
		AWSCredentialsURI: q_aws_credentials_uri,
	}

	if q_debug != "" {

		v, err := strconv.ParseBool(q_debug)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?debug= parameter, %w", err)
		}

		debug = v
		os_client_opts.Debug = debug
	}

	if q_insecure != "" {

		insecure, err := strconv.ParseBool(q_insecure)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?insecure= parameter, %w", err)
		}

		os_client_opts.Insecure = insecure
	}

	return os_client_opts, nil
}

func NewClient(ctx context.Context, uri string) (*opensearch.Client, error) {

	os_client_opts, err := ClientOptionsFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create client options, %w", err)
	}

	return NewClientFromOptions(ctx, os_client_opts)
}

func NewClientFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*opensearch.Client, error) {

	os_client_opts := &ClientOptions{
		Addresses: []string{
			os_endpoint,
		},
		Insecure:          os_insecure,
		Username:          os_username,
		Password:          os_password,
		AWSCredentialsURI: os_aws_uri,
	}

	return NewClientFromOptions(ctx, os_client_opts)
}

// NewClient is an opinionated method for returning a new `opensearch.Client` instance using a `ClientOptions`
// for configuring basic settings for common Opensearch clients. If this method doesn't do what you need it to
// it may make more to create a new client from scratch.
func NewClientFromOptions(ctx context.Context, opts *ClientOptions) (*opensearch.Client, error) {

	retry := backoff.NewExponentialBackOff()

	os_cfg := opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: opts.Insecure,
			},
		},
		Addresses:     opts.Addresses,
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retry.Reset()
			}
			return retry.NextBackOff()
		},
		MaxRetries: 5,
	}

	if opts.Debug {

		opensearch_logger := &opensearchtransport.ColorLogger{
			Output:             os.Stdout,
			EnableRequestBody:  true,
			EnableResponseBody: true,
		}

		os_cfg.Logger = opensearch_logger
	}

	if opts.AWSCredentialsURI != "" {

		aws_cfg, err := auth.NewConfig(ctx, opts.AWSCredentialsURI)

		if err != nil {
			return nil, fmt.Errorf("Failed to create new AWS config, %w", err)
		}

		signer, err := requestsigner.NewSignerWithService(aws_cfg, "es")

		if err != nil {
			return nil, fmt.Errorf("Failed to create request signer, %w", err)
		}

		os_cfg.Signer = signer

	} else {

		os_cfg.Username = opts.Username
		os_cfg.Password = opts.Password
	}

	client, err := opensearch.NewClient(os_cfg)

	if err != nil {
		return nil, fmt.Errorf("Failed to create client, %w", err)
	}

	return client, nil
}
