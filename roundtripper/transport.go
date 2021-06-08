package transport

import (
	"net/http"
	"net/http/httputil"
)

type DebugLogger interface {
	Debug(args ...interface{})
}

type DebugLoggerFunc func(args ...interface{})

func (fn DebugLoggerFunc) Debug(args ...interface{}) {
	fn(args...)
}

type DebugTransport struct {
	rt     http.RoundTripper
	logger DebugLogger
}

func NewDebugTransport(logger DebugLogger, rt http.RoundTripper) *DebugTransport {
	if rt == nil {
		rt = http.DefaultTransport
	}

	if logger == nil {
		logger = DebugLoggerFunc(func(args ...interface{}) {})
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

	d.logger.Debug(string(requestBody))

	resp, err := d.rt.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	respBody, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	d.logger.Debug(string(respBody))

	return resp, nil
}
