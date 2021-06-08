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

func NewDebugTransport(logger Logger, rt http.RoundTripper) *DebugTransport {
	if rt == nil {
		rt = http.DefaultTransport
	}

	if logger == nil {
		logger = func(args ...interface{}) {}
	}

	return &DebugTransport{
		rt:     rt,
		logger: logger,
	}
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
