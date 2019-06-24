package diag

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	tst "github.com/evgeny-myasishchev/ledger.transactions-fetcher/pkg/internal/testing"

	"github.com/bxcodec/faker/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_logrusLogger_log(t *testing.T) {
	type args struct {
		ctx     context.Context
		level   logrus.Level
		msg     string
		args    []interface{}
		msgData MsgData
	}
	type testCase struct {
		name string
		args args
		want func(t *testing.T, actual map[string]interface{})
	}

	tests := []func() testCase{
		func() testCase {
			msg := faker.Sentence()
			return testCase{
				name: "regular msg",
				args: args{
					msg:   msg,
					level: logrus.InfoLevel,
					args:  []interface{}{},
				},
				want: func(t *testing.T, actual map[string]interface{}) {
					assert.Equal(t, msg, actual["msg"])
					assert.Equal(t, float64(1), actual["v"])
				},
			}
		},
		func() testCase {
			return testCase{
				name: "formatted msg",
				args: args{
					msg:   "Formatted msg %s",
					args:  []interface{}{"val1"},
					level: logrus.InfoLevel,
				},
				want: func(t *testing.T, actual map[string]interface{}) {
					assert.Equal(t, "Formatted msg val1", actual["msg"])
				},
			}
		},
		func() testCase {
			requestID := faker.Word()
			ctx := ContextWithRequestID(context.Background(), requestID)
			return testCase{
				name: "with requestID from context",
				args: args{
					ctx:   ctx,
					msg:   "Some msg",
					level: logrus.InfoLevel,
				},
				want: func(t *testing.T, actual map[string]interface{}) {
					if data, ok := actual["context"]; ok {
						contextData := data.(map[string]interface{})
						assert.Equal(t, requestID, contextData["requestID"], "Should have requestID added as context data")
					} else {
						assert.Fail(t, "Should add context")
					}
				},
			}
		},
	}
	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			logger := newLogrusLogger(&out)
			logger.target.Level = tt.args.level
			logger.log(tt.args.ctx, tt.args.level, tt.args.msg, tt.args.args...)

			actual := map[string]interface{}{}
			tst.JSONUnmarshalBuffer(&out, &actual)
			out.Reset()
			tt.want(t, actual)
		})
	}
}

func Benchmark_logrusLogger_log(b *testing.B) {
	var out bytes.Buffer
	logger := newLogrusLogger(&out)
	requestID := faker.Word()
	ctx := ContextWithRequestID(context.Background(), requestID)
	for n := 0; n < b.N; n++ {
		logger.log(ctx, logrus.DebugLevel, "Some msg %v", "val")
	}
}

func Test_Logger_Methods(t *testing.T) {
	now := time.Now()
	type testCase struct {
		name   string
		want   map[string]interface{}
		method func(logger Logger)
	}
	tests := []func() testCase{
		func() testCase {
			msg := faker.Sentence()
			err := errors.New(faker.Sentence())
			msgData := map[string]interface{}{
				"field1": faker.Word(),
				"field2": faker.Word(),
			}
			return testCase{
				name: "error message",
				method: func(logger Logger) {
					logger.WithError(err).WithData(msgData).Error(nil, msg)
				},
				want: map[string]interface{}{
					"level":   "error",
					"msg":     msg,
					"time":    now.Format(time.RFC3339),
					"error":   err.Error(),
					"msgData": msgData,
					"v":       float64(1),
				},
			}
		},
		func() testCase {
			msg := faker.Sentence()
			err := errors.New(faker.Sentence())
			msgData := map[string]interface{}{
				"field1": faker.Word(),
				"field2": faker.Word(),
			}
			return testCase{
				name: "warn message",
				method: func(logger Logger) {
					logger.WithError(err).WithData(msgData).Warn(nil, msg)
				},
				want: map[string]interface{}{
					"level":   "warning",
					"msg":     msg,
					"time":    now.Format(time.RFC3339),
					"error":   err.Error(),
					"msgData": msgData,
					"v":       float64(1),
				},
			}
		},
		func() testCase {
			msg := faker.Sentence()
			err := errors.New(faker.Sentence())
			msgData := map[string]interface{}{
				"field1": faker.Word(),
				"field2": faker.Word(),
			}
			return testCase{
				name: "info message",
				method: func(logger Logger) {
					logger.WithError(err).WithData(msgData).Info(nil, msg)
				},
				want: map[string]interface{}{
					"level":   "info",
					"msg":     msg,
					"time":    now.Format(time.RFC3339),
					"error":   err.Error(),
					"msgData": msgData,
					"v":       float64(1),
				},
			}
		},
		func() testCase {
			msg := faker.Sentence()
			err := errors.New(faker.Sentence())
			msgData := map[string]interface{}{
				"field1": faker.Word(),
				"field2": faker.Word(),
			}
			return testCase{
				name: "debug message",
				method: func(logger Logger) {
					logger.WithError(err).WithData(msgData).Debug(nil, msg)
				},
				want: map[string]interface{}{
					"level":   "debug",
					"msg":     msg,
					"time":    now.Format(time.RFC3339),
					"error":   err.Error(),
					"msgData": msgData,
					"v":       float64(1),
				},
			}
		},
	}
	for _, ttFn := range tests {
		tt := ttFn()
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			logrusLogger := newLogrusLogger(&out)
			logger := Logger(logrusLogger.WithTime(now))
			tt.method(logger)

			got := map[string]interface{}{}
			tst.JSONUnmarshalBuffer(&out, &got)
			out.Reset()

			assert.Equal(t, tt.want, got)
		})
	}
}
