package logzelasticsearch_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/elastic/go-elasticsearch/v7/estransport"
	logzelasticsearch "github.com/karupanerura/logz-elasticsearch"
)

func TestDefaultLogFormatter(t *testing.T) {
	formatter := logzelasticsearch.DefaultLogFormatter

	dummyReq := &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "dummy", Path: "/dummy"}}
	dummyRes := &http.Response{StatusCode: 200}

	gotMessage, err := formatter.Format(dummyReq, dummyRes, nil, time.Now(), 1234*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("message: %q", gotMessage)
	if strings.HasSuffix(gotMessage, "\n") {
		t.Errorf("unexpected message (should chomp line break): %q", gotMessage)
	}
}

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

		return gotMessage == strings.TrimSuffix(msg, "\n")
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

func TestLoggerBasedLogFormatterChomp(t *testing.T) {
	testCases := []struct {
		reqBody, resBody bool
		expected         string
	}{
		{
			reqBody:  false,
			resBody:  false,
			expected: "OK",
		},
		{
			reqBody:  true,
			resBody:  false,
			expected: "OK\n",
		},
		{
			reqBody:  false,
			resBody:  true,
			expected: "OK\n",
		},
		{
			reqBody:  true,
			resBody:  true,
			expected: "OK\n",
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%+v", tc), func(t *testing.T) {
			formatter := logzelasticsearch.NewLoggerBasedLogFormatter(func(w io.Writer) estransport.Logger {
				return mockElasticsearchLogger{
					w: w,
					format: func(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) (string, error) {
						return "OK\n", nil
					},
					requestBodyEnabled:  tc.reqBody,
					responseBodyEnabled: tc.resBody,
				}
			})

			gotMessage, err := formatter.Format(&http.Request{}, &http.Response{StatusCode: 200}, nil, time.Now(), 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMessage != tc.expected {
				t.Fatalf("unexpected message: %q", gotMessage)
			}
		})
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
