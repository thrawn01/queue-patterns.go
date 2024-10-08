package queue

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/duh-rpc/duh-go"
	v1 "github.com/duh-rpc/duh-go/proto/v1"
	"github.com/kapetan-io/tackle/clock"
	pb "github.com/thrawn01/queue-patterns.go/proto"
	"google.golang.org/protobuf/proto"
	"net/http"
)

type ClientConfig struct {
	// Users can provide their own http client with TLS config if needed
	Client *http.Client
	// The address of endpoint in the format `<scheme>://<host>:<port>`
	Endpoint string
}

type Client struct {
	client *duh.Client
	conf   ClientConfig
}

// NewClient creates a new instance of the Gubernator user client
func NewClient(conf ClientConfig) (*Client, error) {
	if conf.Client == nil {
		return nil, errors.New("conf.Client is nil; must provide an http.Client")
	}

	if len(conf.Endpoint) == 0 {
		return nil, errors.New("conf.Endpoint is empty; must provide an http endpoint")
	}

	return &Client{
		client: &duh.Client{
			Client: conf.Client,
		},
		conf: conf,
	}, nil
}

func (c *Client) ProduceItems(ctx context.Context, req *pb.ProduceRequest) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, "/produce"), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	var res v1.Reply
	return c.client.Do(r, &res)
}

// WithNoTLS returns ClientConfig suitable for use with NON-TLS clients
func WithNoTLS(address string) ClientConfig {
	return ClientConfig{
		Endpoint: fmt.Sprintf("http://%s", address),
		Client: &http.Client{
			Transport: &http.Transport{
				// NOTE: Set to '1' here to simulate single writer in benchmarks
				MaxConnsPerHost:     1,
				MaxIdleConns:        1,
				MaxIdleConnsPerHost: 1,
				IdleConnTimeout:     60 * clock.Second,
			},
		},
	}
}

// WithTLS returns ClientConfig suitable for use with TLS clients
func WithTLS(tls *tls.Config, address string) ClientConfig {
	return ClientConfig{
		Endpoint: fmt.Sprintf("https://%s", address),
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tls,
				// NOTE: Set to '1' here to simulate single writer in benchmarks
				MaxConnsPerHost:     1,
				MaxIdleConns:        1,
				MaxIdleConnsPerHost: 1,
				IdleConnTimeout:     60 * clock.Second,
			},
		},
	}
}
