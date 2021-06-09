package transport

import (
	"net/http"
	"net/http/httputil"
)

type Logger func(args ...interface{})

type DebugTransport struct {
	rt     http.RoundTripper
	logger Logger
}

type DebugTransportOption func(*DebugTransport)

func WithLogger(loggor Logger) DebugTransportOption {
	return func(dt *DebugTransport) {
		dt.logger = loggor
	}
}

func WithRoundTripper(rt http.RoundTripper) DebugTransportOption {
	return func(dt *DebugTransport) {
		dt.rt = rt
	}
}

func NewDebugTransport(opts ...DebugTransportOption) *DebugTransport {
	dt := new(DebugTransport)

	for _, opt := range opts {
		opt(dt)
	}

	if dt.rt == nil {
		dt.rt = http.DefaultTransport
	}

	if dt.logger == nil {
		dt.logger = func(args ...interface{}) {}
	}

	return dt
}

func (d *DebugTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	requestBody, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		return nil, err
	}

	d.logger(string(requestBody))

	resp, err := d.rt.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	respBody, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	d.logger(string(respBody))

	return resp, nil
}
