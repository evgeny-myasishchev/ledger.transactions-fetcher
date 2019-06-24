package diag

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

type wantLogData struct {
	ctx     context.Context
	msg     string
	msgData MsgData
}

type mockLogger struct {
	gotLogs       []wantLogData
	recentMsgData MsgData
}

func (l *mockLogger) log(ctx context.Context, msg string, args ...interface{}) {
	l.gotLogs = append(l.gotLogs, wantLogData{
		ctx:     ctx,
		msg:     fmt.Sprintf(msg, args...),
		msgData: l.recentMsgData,
	})
	l.recentMsgData = nil
}

func (l *mockLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, msg, args...)
}

func (l *mockLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, msg, args...)
}

func (l *mockLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, msg, args...)
}

func (l *mockLogger) Debug(ctx context.Context, msg string, args ...interface{}) {
	l.log(ctx, msg, args...)
}

func (l *mockLogger) WithError(err error) Logger {
	panic("not implemented")
}

func (l *mockLogger) WithData(data MsgData) Logger {
	l.recentMsgData = data
	return l
}

func TestRequestIDMiddleware(t *testing.T) {
	type args struct {
		w     http.ResponseWriter
		req   *http.Request
		setup []requestIDMiddlewareSetup
	}
	type testCase struct {
		name         string
		args         args
		want         string
		wantNotEmpty bool
	}
	tests := []func() testCase{
		func() testCase {
			requestID := uuid.NewV4().String()
			req := httptest.NewRequest("GET", "/not-important", nil)
			req.Header.Add("X-Request-ID", requestID)
			return testCase{
				name: "reuse requestID from header",
				args: args{
					req: req,
					w:   httptest.NewRecorder(),
				},
				want: requestID,
			}
		},
		func() testCase {
			requestID := uuid.NewV4()
			req := httptest.NewRequest("GET", "/not-important", nil)
			return testCase{
				name: "generate a new requestID",
				args: args{
					req: req,
					w:   httptest.NewRecorder(),
					setup: []requestIDMiddlewareSetup{
						func(cfg *requestIDMiddlewareCfg) {
							cfg.newUUID = func() uuid.UUID { return requestID }
						},
					},
				},
				want: requestID.String(),
			}
		},
		func() testCase {
			req := httptest.NewRequest("GET", "/not-important", nil)
			return testCase{
				name: "generate a new requestID with a default cfg",
				args: args{
					req: req,
					w:   httptest.NewRecorder(),
				},
				wantNotEmpty: true,
			}
		},
	}
	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				nextCalled = true
				requestID := RequestIDValue(req.Context())
				if tt.wantNotEmpty {
					assert.NotEmpty(t, requestID)
				} else {
					assert.Equal(t, tt.want, requestID)
				}
			})
			mw := NewRequestIDMiddleware(tt.args.setup...)
			mw(next).ServeHTTP(tt.args.w, tt.args.req)
			assert.True(t, nextCalled, "Next should have been called")
		})
	}
}

func TestLogRequestsMiddleware(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	tests := []struct {
		name     string
		testCase func(t *testing.T)
	}{
		{
			name: "log request start/end",
			testCase: func(t *testing.T) {
				fakeUserAgent := faker.Word()
				fakeURL := url.URL{
					Path: fmt.Sprintf("/path/%v", faker.Word()),
					RawQuery: fmt.Sprintf("key1=%v&key2=%v&key3=%v&key3=%v",
						faker.Word(), faker.Word(),
						faker.Word(), faker.Word(),
					),
				}
				req := httptest.NewRequest("GET", fakeURL.RequestURI(), nil)
				req.Header.Set("User-Agent", fakeUserAgent)
				req.Header.Add("X-Multi-Values", faker.Word())
				req.Header.Add("X-Multi-Values", faker.Word())
				remoteIP := faker.IPv4()
				remotePort := strconv.FormatInt(rand.Int63n(255), 10)
				req.RemoteAddr = fmt.Sprintf("%v:%v", remoteIP, remotePort)

				l := mockLogger{
					gotLogs: []wantLogData{},
				}

				duration := time.Duration(int64(time.Millisecond) * rand.Int63n(1000))

				reqStart := time.Now()
				reqEnd := reqStart.Add(duration)
				times := []time.Time{reqStart, reqEnd}

				fakeMemoryStats := rand.Float64()
				mw := NewLogRequestsMiddleware(func(cfg *logRequestsMiddlewareCfg) {
					cfg.logger = &l
					cfg.runtimeMemMb = func() float64 {
						return fakeMemoryStats
					}
					cfg.now = func() time.Time {
						var now time.Time
						now, times = times[0], times[1:]
						return now
					}
				})

				code := 100 + rand.Intn(500)
				w := httptest.NewRecorder()
				nextCalled := false
				next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					nextCalled = true
					assert.Implements(t, (*http.ResponseWriter)(nil), w)
					assert.IsType(t, (*http.Request)(nil), req)
					w.Header().Add("X-Multi-Values", faker.Word())
					w.Header().Add("X-Multi-Values", faker.Word())
					w.WriteHeader(code)
				})
				mw(next).ServeHTTP(w, req)
				assert.True(t, nextCalled, "Next should have been called")

				wantLogs := []wantLogData{
					wantLogData{
						ctx: req.Context(),
						msg: fmt.Sprintf("BEGIN REQ: GET %v", fakeURL.Path),
						msgData: MsgData{
							"method":        "GET",
							"url":           fakeURL.RequestURI(),
							"path":          fakeURL.Path,
							"userAgent":     req.UserAgent(),
							"query":         flattenAndObfuscate(req.URL.Query()),
							"headers":       flattenAndObfuscate(req.Header),
							"remoteAddress": remoteIP,
							"remotePort":    remotePort,
							"memoryUsageMb": fakeMemoryStats,
						},
					},
					wantLogData{
						ctx: req.Context(),
						msg: fmt.Sprintf("END REQ: %v - %v", code, fakeURL.Path),
						msgData: MsgData{
							"statusCode":    code,
							"headers":       flattenAndObfuscate(w.Header()),
							"duration":      duration.Seconds(),
							"memoryUsageMb": fakeMemoryStats,
						},
					},
				}
				assert.Equal(t, wantLogs, l.gotLogs)
			},
		},
		{
			name: "use default status",
			testCase: func(t *testing.T) {
				req := httptest.NewRequest("GET", "/fake", nil)
				l := mockLogger{gotLogs: []wantLogData{}}

				mw := NewLogRequestsMiddleware(func(cfg *logRequestsMiddlewareCfg) {
					cfg.logger = &l
				})

				w := httptest.NewRecorder()
				nextCalled := false
				next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					nextCalled = true
					assert.Implements(t, (*http.ResponseWriter)(nil), w)
					assert.IsType(t, (*http.Request)(nil), req)
				})
				mw(next).ServeHTTP(w, req)
				assert.True(t, nextCalled, "Next should have been called")
				if !assert.Len(t, l.gotLogs, 2) {
					assert.FailNow(t, "Can not continue")
				}

				endLog := l.gotLogs[1]
				assert.Contains(t, endLog.msg, "END REQ: 200")
				assert.Equal(t, 200, endLog.msgData["statusCode"])
			},
		},
		{
			name: "ignore paths",
			testCase: func(t *testing.T) {
				fakePath := fmt.Sprintf("/fake/path/%v", faker.Word())
				req := httptest.NewRequest("GET", fakePath, nil)
				l := mockLogger{gotLogs: []wantLogData{}}

				mw := NewLogRequestsMiddleware(
					func(cfg *logRequestsMiddlewareCfg) {
						cfg.logger = &l
					},
					IgnorePath(fakePath),
				)

				w := httptest.NewRecorder()
				nextCalled := false
				next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					nextCalled = true
					assert.Implements(t, (*http.ResponseWriter)(nil), w)
					assert.IsType(t, (*http.Request)(nil), req)
				})
				mw(next).ServeHTTP(w, req)
				assert.True(t, nextCalled, "Next should have been called")
				assert.Len(t, l.gotLogs, 0)
			},
		},
		{
			name: "ignore default paths",
			testCase: func(t *testing.T) {
				req := httptest.NewRequest("GET", "/v1/healthcheck/ping", nil)
				l := mockLogger{gotLogs: []wantLogData{}}

				mw := NewLogRequestsMiddleware(func(cfg *logRequestsMiddlewareCfg) {
					cfg.logger = &l
				})

				w := httptest.NewRecorder()
				nextCalled := false
				next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					nextCalled = true
					assert.Implements(t, (*http.ResponseWriter)(nil), w)
					assert.IsType(t, (*http.Request)(nil), req)
				})
				mw(next).ServeHTTP(w, req)
				assert.True(t, nextCalled, "Next should have been called")
				assert.Len(t, l.gotLogs, 0)
			},
		},
		{
			name: "obfuscate default headers",
			testCase: func(t *testing.T) {
				req := httptest.NewRequest("GET", "/", nil)
				l := mockLogger{gotLogs: []wantLogData{}}

				mw := NewLogRequestsMiddleware(
					func(cfg *logRequestsMiddlewareCfg) {
						cfg.logger = &l
					},
				)

				w := httptest.NewRecorder()
				nextCalled := false
				authToken := faker.Word()
				req.Header.Add("Authorization", authToken)
				next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					nextCalled = true
				})
				mw(next).ServeHTTP(w, req)
				assert.True(t, nextCalled)
				assert.Len(t, l.gotLogs, 2)
				loggedHeaders := l.gotLogs[0].msgData["headers"].(map[string]string)
				assert.Equal(t,
					fmt.Sprint("*obfuscated, length=", len(authToken), "*"),
					loggedHeaders["Authorization"],
				)
			},
		},
		{
			name: "obfuscate custom headers",
			testCase: func(t *testing.T) {
				customHeader1 := "X-Custom1-H" + faker.Word()
				customValue1 := faker.Word()

				customHeader2 := "X-Custom2-H" + faker.Word()
				customValue2 := faker.Word()

				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Add(customHeader1, customValue1)
				req.Header.Add(customHeader2, customValue2)

				l := mockLogger{gotLogs: []wantLogData{}}

				mw := NewLogRequestsMiddleware(
					func(cfg *logRequestsMiddlewareCfg) {
						cfg.logger = &l
					},
					ObfuscateHeaders(customHeader1, customHeader2),
				)

				w := httptest.NewRecorder()
				nextCalled := false
				next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					nextCalled = true
				})
				mw(next).ServeHTTP(w, req)
				assert.True(t, nextCalled)
				assert.Len(t, l.gotLogs, 2)
				loggedHeaders := l.gotLogs[0].msgData["headers"].(map[string]string)
				assert.Equal(t,
					fmt.Sprint("*obfuscated, length=", len(customValue1), "*"),
					loggedHeaders[customHeader1],
				)
				assert.Equal(t,
					fmt.Sprint("*obfuscated, length=", len(customValue2), "*"),
					loggedHeaders[customHeader2],
				)
			},
		},
		{
			name: "remote address without port",
			testCase: func(t *testing.T) {
				req := httptest.NewRequest("GET", "/fake", nil)
				remoteIP := faker.IPv4()
				req.RemoteAddr = remoteIP

				l := mockLogger{
					gotLogs: []wantLogData{},
				}
				mw := NewLogRequestsMiddleware(func(cfg *logRequestsMiddlewareCfg) {
					cfg.logger = &l
				})

				w := httptest.NewRecorder()
				nextCalled := false
				next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					nextCalled = true
				})
				mw(next).ServeHTTP(w, req)
				assert.True(t, nextCalled, "Next should have been called")
				assert.Len(t, l.gotLogs, 3)
				startLog := l.gotLogs[1]
				assert.Equal(t, remoteIP, startLog.msgData["remoteAddress"])
				assert.Equal(t, "", startLog.msgData["remotePort"])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.testCase)
	}
}

func Test_flattenAndObfuscate(t *testing.T) {
	type args struct {
		values    map[string][]string
		obfuscate []string
	}
	type testCase struct {
		name string
		args args
		want map[string]string
	}
	tests := []func() testCase{
		func() testCase {
			key1 := "key1-" + faker.Word()
			val1 := "val1-" + faker.Word()
			key2 := "key2-" + faker.Word()
			val2 := "val2-" + faker.Word()

			values := map[string][]string{
				key1: []string{val1},
				key2: []string{val2},
			}
			return testCase{
				name: "single value",
				args: args{values: values},
				want: map[string]string{
					key1: val1,
					key2: val2,
				},
			}
		},
		func() testCase {
			key1 := "key1-" + faker.Word()
			val11 := "val11-" + faker.Word()
			val12 := "val12-" + faker.Word()
			key2 := "key2-" + faker.Word()
			val21 := "val21-" + faker.Word()
			val22 := "val22" + faker.Word()

			values := map[string][]string{
				key1: []string{val11, val12},
				key2: []string{val21, val22},
			}
			return testCase{
				name: "multiple value",
				args: args{values: values},
				want: map[string]string{
					key1: val11 + ", " + val12,
					key2: val21 + ", " + val22,
				},
			}
		},
		func() testCase {
			key1 := "key1-" + faker.Word()
			val1 := "val1-" + faker.Word()
			key2 := "key2-" + faker.Word()
			val2 := "val2-" + faker.Word()

			values := map[string][]string{
				key1: []string{val1},
				key2: []string{val2},
			}
			return testCase{
				name: "obfuscate",
				args: args{values: values, obfuscate: []string{key1}},
				want: map[string]string{
					key1: fmt.Sprint("*obfuscated, length=", len(val1), "*"),
					key2: val2,
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			got := flattenAndObfuscate(tt.args.values, tt.args.obfuscate...)
			assert.Equal(t, tt.want, got)
		})
	}
}
