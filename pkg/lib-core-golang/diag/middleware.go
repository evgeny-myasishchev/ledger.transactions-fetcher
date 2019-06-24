package diag

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

// Note: we can not reference router here since it'll create a cyclic imports
// so MiddlewareFunc type can not be used

type requestIDMiddlewareCfg struct {
	newUUID func() uuid.UUID
}

type requestIDMiddlewareSetup func(cfg *requestIDMiddlewareCfg)

// NewRequestIDMiddleware - creates a middleware that will maintain the requestId header
func NewRequestIDMiddleware(setup ...requestIDMiddlewareSetup) func(next http.Handler) http.Handler {
	cfg := requestIDMiddlewareCfg{newUUID: uuid.NewV4}
	for _, setupFn := range setup {
		setupFn(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestID := req.Header.Get("x-request-id")
			if requestID == "" {
				requestID = cfg.newUUID().String()
			}
			nextCtx := ContextWithRequestID(req.Context(), requestID)
			w.Header().Add("x-request-id", requestID)
			next.ServeHTTP(w, req.WithContext(nextCtx))
		})
	}
}

type loggingMiddlewareResponseWrapper struct {
	target http.ResponseWriter
	status int
}

func (lmw *loggingMiddlewareResponseWrapper) Header() http.Header {
	return lmw.target.Header()
}

func (lmw *loggingMiddlewareResponseWrapper) Write(b []byte) (int, error) {
	return lmw.target.Write(b)
}

func (lmw *loggingMiddlewareResponseWrapper) WriteHeader(status int) {
	lmw.target.WriteHeader(status)
	lmw.status = status
}

func (lmw *loggingMiddlewareResponseWrapper) getStatus() int {
	if lmw.status == 0 {
		return 200
	}
	return lmw.status
}

// logRequestsMiddlewareCfg represents a config for the requests logging middleware
type logRequestsMiddlewareCfg struct {
	ignorePaths      map[string]bool
	obfuscateHeaders []string
	logger           Logger
	runtimeMemMb     func() float64
	now              func() time.Time
}

// LogRequestsMiddlewareOpt is a type used to supply various opts
// for requests logger middleware
type LogRequestsMiddlewareOpt func(*logRequestsMiddlewareCfg)

// IgnorePath option specify paths to skip log requests for
func IgnorePath(path string) LogRequestsMiddlewareOpt {
	return func(cfg *logRequestsMiddlewareCfg) {
		cfg.ignorePaths[path] = true
	}
}

// ObfuscateHeaders option provides a list of headers to obfuscate (e.g do not log values)
func ObfuscateHeaders(headers ...string) LogRequestsMiddlewareOpt {
	return func(cfg *logRequestsMiddlewareCfg) {
		cfg.obfuscateHeaders = append(cfg.obfuscateHeaders, headers...)
	}
}

func flattenAndObfuscate(values map[string][]string, obfuscateKeys ...string) map[string]string {
	flattened := make(map[string]string, len(values))
	for key, val := range values {
		flattened[key] = strings.Join(val, ", ")
	}
	for _, obfuscateKey := range obfuscateKeys {
		if val, ok := flattened[obfuscateKey]; ok {
			flattened[obfuscateKey] = fmt.Sprint("*obfuscated, length=", len(val), "*")
		}
	}
	return flattened
}

// NewLogRequestsMiddleware - log request start/end
func NewLogRequestsMiddleware(opts ...LogRequestsMiddlewareOpt) func(next http.Handler) http.Handler {
	cfg := logRequestsMiddlewareCfg{
		ignorePaths:      map[string]bool{},
		obfuscateHeaders: []string{"Authorization"},
	}
	cfg.ignorePaths["/v1/healthcheck/ping"] = true
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.logger == nil {
		cfg.logger = CreateLogger()
	}
	if cfg.runtimeMemMb == nil {
		cfg.runtimeMemMb = func() float64 {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			return math.Round(float64(memStats.Alloc)/1024.0/1024.0*1000) / 1000
		}
	}
	if cfg.now == nil {
		cfg.now = time.Now
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			method := req.Method
			path := req.URL.Path

			if _, ok := cfg.ignorePaths[path]; ok {
				next.ServeHTTP(w, req)
				return
			}

			ip, port, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				cfg.logger.Warn(req.Context(), "Can not parse remote addr: %v", req.RemoteAddr)
				ip = req.RemoteAddr
			}

			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			cfg.logger.
				WithData(MsgData{
					"method":        method,
					"url":           req.URL.RequestURI(),
					"path":          req.URL.Path,
					"userAgent":     req.UserAgent(),
					"headers":       flattenAndObfuscate(req.Header, cfg.obfuscateHeaders...),
					"query":         flattenAndObfuscate(req.URL.Query()),
					"remoteAddress": ip,
					"remotePort":    port,
					"memoryUsageMb": cfg.runtimeMemMb(),
				}).
				Info(req.Context(), "BEGIN REQ: %s %s", method, path)

			wrappedWriter := loggingMiddlewareResponseWrapper{
				target: w,
			}
			reqStartedAt := cfg.now()
			next.ServeHTTP(&wrappedWriter, req)
			reqDuration := cfg.now().Sub(reqStartedAt)

			responseStatus := wrappedWriter.getStatus()

			cfg.logger.
				WithData(MsgData{
					"statusCode":    responseStatus,
					"headers":       flattenAndObfuscate(w.Header()),
					"duration":      reqDuration.Seconds(),
					"memoryUsageMb": cfg.runtimeMemMb(),
				}).
				Info(req.Context(), "END REQ: %v - %v", responseStatus, path)
		})
	}
}
