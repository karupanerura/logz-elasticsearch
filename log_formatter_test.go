package logzelasticsearch_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"testing/quick"
	"time"

	"github.com/elastic/go-elasticsearch/v7/estransport"
	logzelasticsearch "github.com/karupanerura/logz-elasticsearch"
)

func TestLoggerBasedLogFormatterMessage(t *testing.T) {
	f := func(msg string) bool {
		formatter := logzelasticsearch.NewLoggerBasedLogFormatter(func(w io.Writer) estransport.Logger {
			return mockElasticsearchLogger{
				w: w,
				format: func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
					return msg, nil
				},
			}
		})

		gotMessage, err := formatter.Format(&http.Request{}, &http.Response{StatusCode: 200}, nil, time.Now(), 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		return gotMessage == msg
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestLoggerBasedLogFormatterError(t *testing.T) {
	f := func(msg string) bool {
		expectedErr := errors.New(msg)
		formatter := logzelasticsearch.NewLoggerBasedLogFormatter(func(w io.Writer) estransport.Logger {
			return mockElasticsearchLogger{
				w: w,
				format: func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
					return "", expectedErr
				},
			}
		})

		gotMessage, err := formatter.Format(&http.Request{}, &http.Response{StatusCode: 200}, nil, time.Now(), 0)
		if gotMessage != "" {
			t.Fatalf("unexpected message: %s", gotMessage)
		}

		return err == expectedErr
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestLoggerBasedLogFormatterParams(t *testing.T) {
	testCases := []struct {
		reqBody, resBody bool
	}{
		{
			reqBody: false,
			resBody: false,
		},
		{
			reqBody: true,
			resBody: false,
		},
		{
			reqBody: false,
			resBody: true,
		},
		{
			reqBody: true,
			resBody: true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%+v", tc), func(t *testing.T) {
			formatter := logzelasticsearch.NewLoggerBasedLogFormatter(func(w io.Writer) estransport.Logger {
				return mockElasticsearchLogger{
					w: w,
					format: func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
						return "OK", nil
					},
					requestBodyEnabled:  tc.reqBody,
					responseBodyEnabled: tc.resBody,
				}
			})
			if got := formatter.RequestBodyEnabled(); got != tc.reqBody {
				t.Errorf("unexpected RequestBodyEnabled: %t (expected: %t)", got, tc.reqBody)
			}
			if got := formatter.ResponseBodyEnabled(); got != tc.resBody {
				t.Errorf("unexpected ResponseBodyEnabled: %t (expected: %t)", got, tc.resBody)
			}
		})
	}
}

func TestPrefixedLogFormatterMessage(t *testing.T) {
	f := func(prefix, msg string) bool {
		expectedMessage := prefix + msg
		formatter := &logzelasticsearch.PrefixedLogFormatter{
			Prefix: prefix,
			LogFormatter: mockLogFormatter{
				format: func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
					return msg, nil
				},
			},
		}

		gotMessage, err := formatter.Format(&http.Request{}, &http.Response{StatusCode: 200}, nil, time.Now(), 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		return gotMessage == expectedMessage
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestPrefixedLogFormatterError(t *testing.T) {
	f := func(msg string) bool {
		expectedErr := errors.New(msg)
		formatter := &logzelasticsearch.PrefixedLogFormatter{
			Prefix: "Prefix: ",
			LogFormatter: mockLogFormatter{
				format: func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
					return "", expectedErr
				},
			},
		}

		gotMessage, err := formatter.Format(&http.Request{}, &http.Response{StatusCode: 200}, nil, time.Now(), 0)
		if gotMessage != "" {
			t.Fatalf("unexpected message: %s", gotMessage)
		}

		return err == expectedErr
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
