package transport_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jacoelho/transport"
)

var ErrHealthCheck = errors.New("health check failed")

type Client struct {
	client *http.Client
	url    string
}

func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrHealthCheck
	}

	return nil
}

func TestTransport_DebugLoggerFunc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := &Client{
		client: http.DefaultClient,
		url:    ts.URL,
	}

	sb := new(strings.Builder)
	stringLogger := func(args ...interface{}) {
		fmt.Fprint(sb, args...)
	}

	c.client.Transport = transport.NewDebugTransport(transport.WithLogger(stringLogger), transport.WithRoundTripper(c.client.Transport))

	err := c.HealthCheck(context.Background())
	if !errors.Is(err, ErrHealthCheck) {
		t.Error("expected error")
	}

	got := sb.String()

	if !strings.Contains(got, `GET / HTTP/1.1`) {
		t.Errorf("expected http request, got %s", got)
	}

	if !strings.Contains(got, `HTTP/1.1 500 Internal Server Error`) {
		t.Errorf("expected 500 response, got %s", got)
	}
}

func TestTransport_DebugLoggerNotSet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := &Client{
		client: http.DefaultClient,
		url:    ts.URL,
	}

	c.client.Transport = transport.NewDebugTransport()

	err := c.HealthCheck(context.Background())
	if !errors.Is(err, ErrHealthCheck) {
		t.Error("expected error")
	}
}
