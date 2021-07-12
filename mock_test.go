package logzelasticsearch_test

import (
	"io"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7/estransport"
	logzelasticsearch "github.com/karupanerura/logz-elasticsearch"
)

type mockLogFormatter struct {
	format              func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error)
	requestBodyEnabled  bool
	responseBodyEnabled bool
}

var _ logzelasticsearch.LogFormatter = mockLogFormatter{}

func (f mockLogFormatter) Format(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
	return f.format(req, res, err, start, dur)
}

func (f mockLogFormatter) RequestBodyEnabled() bool {
	return f.requestBodyEnabled
}

func (f mockLogFormatter) ResponseBodyEnabled() bool {
	return f.responseBodyEnabled
}

type mockElasticsearchLogger struct {
	w                   io.Writer
	format              func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error)
	requestBodyEnabled  bool
	responseBodyEnabled bool
}

var _ estransport.Logger = mockElasticsearchLogger{}

func (f mockElasticsearchLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	if s, err := f.format(req, res, err, start, dur); err != nil {
		return err
	} else {
		_, err = io.WriteString(f.w, s)
		return err
	}
}

func (f mockElasticsearchLogger) RequestBodyEnabled() bool {
	return f.requestBodyEnabled
}

func (f mockElasticsearchLogger) ResponseBodyEnabled() bool {
	return f.responseBodyEnabled
}
