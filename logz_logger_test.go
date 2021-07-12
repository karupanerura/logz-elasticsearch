package logzelasticsearch_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"testing/quick"
	"time"

	"github.com/glassonion1/logz/testhelper"
	logzelasticsearch "github.com/karupanerura/logz-elasticsearch"
)

func TestLogzLoggerSeverity(t *testing.T) {
	logger := &logzelasticsearch.LogzLogger{
		mockLogFormatter{
			format: func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
				return "OK", nil
			},
		},
	}

	testCases := []struct {
		Name string

		StatusCode int
		Severity   string
	}{
		{
			Name:       "200 OK",
			StatusCode: 200,
			Severity:   "DEBUG",
		},
		{
			Name:       "301 Moved",
			StatusCode: 301,
			Severity:   "DEBUG",
		},
		{
			Name:       "400 Bad Request",
			StatusCode: 400,
			Severity:   "INFO",
		},
		{
			Name:       "500 Internal Server Error",
			StatusCode: 500,
			Severity:   "WARNING",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			got := testhelper.ExtractApplicationLogOut(t, func() {
				err := logger.LogRoundTrip(&http.Request{}, &http.Response{StatusCode: tc.StatusCode}, nil, time.Now(), 0)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})

			var v struct {
				Severity string `json:"severity"`
			}
			err := json.Unmarshal([]byte(got), &v)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if v.Severity != tc.Severity {
				t.Errorf("unexpected severity: %q (expected=%q)", v.Severity, tc.Severity)
			}
		})
	}
}

func TestLogzLoggerMessage(t *testing.T) {
	f := func(msg string) bool {
		logger := &logzelasticsearch.LogzLogger{
			mockLogFormatter{
				format: func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
					return msg, nil
				},
			},
		}

		got := testhelper.ExtractApplicationLogOut(t, func() {
			err := logger.LogRoundTrip(&http.Request{}, &http.Response{StatusCode: 200}, nil, time.Now(), 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		var v struct {
			Severity string `json:"severity"`
			Message  string `json:"message"`
		}
		err := json.Unmarshal([]byte(got), &v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		return v.Severity == "DEBUG" && v.Message == msg
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestLogzLoggerError(t *testing.T) {
	f := func(msg string) bool {
		expectedErr := errors.New(msg)
		logger := &logzelasticsearch.LogzLogger{
			mockLogFormatter{
				format: func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
					return "", expectedErr
				},
			},
		}

		var called bool
		var gotErr error
		got := testhelper.ExtractApplicationLogOut(t, func() {
			if called {
				t.Fatalf("Duplicated called, already got error: %v", gotErr)
			}

			gotErr = logger.LogRoundTrip(&http.Request{}, &http.Response{StatusCode: 200}, nil, time.Now(), 0)
			called = true
		})
		if got != "" {
			t.Errorf("unexpected log: %s", got)
		}

		return gotErr == expectedErr
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
