package logzelasticsearch

import (
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7/estransport"
	"github.com/glassonion1/logz"
)

// LogzLogger is ...
type LogzLogger struct {
	LogFormatter
}

var _ estransport.Logger = &LogzLogger{}

// DefaultLogzLogger is a Logger using DefaultLogFormatter.
var DefaultLogzLogger = &LogzLogger{DefaultLogFormatter}

// LogRoundTrip is ...
func (l *LogzLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	s, err := l.Format(req, res, err, start, dur)
	if err != nil {
		return err
	}

	// TODO: get severity from callback
	if res.StatusCode < 400 {
		logz.Debugf(req.Context(), s)
	} else if res.StatusCode < 500 {
		logz.Warningf(req.Context(), s)
	}
	return nil
}
