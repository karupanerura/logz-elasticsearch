package logzelasticsearch

import (
	"io"
	"net/http"
	"strings"
	"sync"
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
	pool                sync.Pool
	requestBodyEnabled  onceDecisiveFlag
	responseBodyEnabled onceDecisiveFlag
}

// NewLoggerBasedLogFormatter creates a new LoggerBasedLogFormatter.
func NewLoggerBasedLogFormatter(newLogger func(w io.Writer) estransport.Logger) *LoggerBasedLogFormatter {
	formatter := &LoggerBasedLogFormatter{
		pool: sync.Pool{
			New: func() interface{} {
				core := &loggerBasedLogFormatterCore{}
				core.logger = newLogger(&core.builder)
				return core
			},
		},
	}
	return formatter
}

// Format a new log from elasticsearch log informations.
func (f *LoggerBasedLogFormatter) Format(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
	core := f.pool.Get().(*loggerBasedLogFormatterCore)
	defer f.pool.Put(core)
	return core.format(req, res, err, start, dur)
}

// RequestBodyEnabled makes the client pass a copy of request body to the formatter.
func (f *LoggerBasedLogFormatter) RequestBodyEnabled() bool {
	return f.requestBodyEnabled.GetOnce(func() bool {
		core := f.pool.Get().(*loggerBasedLogFormatterCore)
		defer f.pool.Put(core)
		return core.logger.RequestBodyEnabled()
	})
}

// ResponseBodyEnabled makes the client pass a copy of response body to the formatter.
func (f *LoggerBasedLogFormatter) ResponseBodyEnabled() bool {
	return f.responseBodyEnabled.GetOnce(func() bool {
		core := f.pool.Get().(*loggerBasedLogFormatterCore)
		defer f.pool.Put(core)
		return core.logger.ResponseBodyEnabled()
	})
}

type loggerBasedLogFormatterCore struct {
	logger  estransport.Logger
	builder strings.Builder
}

func (f *loggerBasedLogFormatterCore) format(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
	defer f.builder.Reset()
	if err := f.logger.LogRoundTrip(req, res, err, start, dur); err != nil {
		return "", err
	}

	return f.builder.String(), nil
}

type onceDecisiveFlag struct {
	once  sync.Once
	value bool
}

func (f *onceDecisiveFlag) GetOnce(get func() bool) bool {
	f.once.Do(func() {
		f.value = get()
	})
	return f.value
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
