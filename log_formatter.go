package logzelasticsearch

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7/estransport"
)

// LogFormatter is ...
type LogFormatter interface {
	// Format makes a new log from elasticsearch log informations.
	Format(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error)

	// RequestBodyEnabled makes the client pass a copy of request body to the logger.
	RequestBodyEnabled() bool

	// ResponseBodyEnabled makes the client pass a copy of response body to the logger.
	ResponseBodyEnabled() bool
}

// DefaultLogFormatter is a estransport.TextLogger based log formatter. it doesn't include request body and response body to the new log.
var DefaultLogFormatter LogFormatter = NewLoggerBasedLogFormatter(func(w io.Writer) estransport.Logger {
	return &estransport.TextLogger{Output: w}
})

// LoggerBasedLogFormatter is log formatter using estransport.Logger.
type LoggerBasedLogFormatter struct {
	logger  estransport.Logger
	builder strings.Builder
}

// NewLoggerBasedLogFormatter creates a new LoggerBasedLogFormatter.
func NewLoggerBasedLogFormatter(newLogger func(w io.Writer) estransport.Logger) *LoggerBasedLogFormatter {
	formatter := &LoggerBasedLogFormatter{}
	formatter.logger = newLogger(&formatter.builder)
	return formatter
}

// Format a new log from elasticsearch log informations.
func (f *LoggerBasedLogFormatter) Format(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
	defer f.builder.Reset()
	if err := f.logger.LogRoundTrip(req, res, err, start, dur); err != nil {
		return "", err
	}

	return f.builder.String(), nil
}

// RequestBodyEnabled makes the client pass a copy of request body to the formatter.
func (f *LoggerBasedLogFormatter) RequestBodyEnabled() bool {
	return f.logger.RequestBodyEnabled()
}

// ResponseBodyEnabled makes the client pass a copy of response body to the formatter.
func (f *LoggerBasedLogFormatter) ResponseBodyEnabled() bool {
	return f.logger.ResponseBodyEnabled()
}

// PrefixedLogFormatter is ...
type PrefixedLogFormatter struct {
	Prefix string
	LogFormatter
}

// Format a new log by LogFormatter and add Prefix for the new log.
func (f *PrefixedLogFormatter) Format(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
	if s, err := f.LogFormatter.Format(req, res, err, start, dur); err != nil {
		return "", nil
	} else {
		return f.Prefix + s, nil
	}
}
