package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	connect "connectrpc.com/connect"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/nrzap"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	health_grpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/code-payments/ocp-server/grpc/headers"
	grpc_metrics "github.com/code-payments/ocp-server/grpc/metrics"
	"github.com/code-payments/ocp-server/grpc/protobuf/validation"
	"github.com/code-payments/ocp-server/metrics"
	newrelic_metrics "github.com/code-payments/ocp-server/metrics/newrelic"
	noop_metrics "github.com/code-payments/ocp-server/metrics/noop"

	indexerpb "github.com/code-payments/code-vm-indexer/generated/indexer/v1"
	"github.com/code-payments/code-vm-indexer/generated/indexer/v1/indexerv1connect"

	indexerapp "github.com/code-payments/code-vm-indexer/app"
	"github.com/code-payments/code-vm-indexer/rpc"
)

const (
	listenAddressEnv     = "LISTEN_ADDRESS"
	defaultListenAddress = ":8086"

	corsAllowedOriginsEnv     = "CORS_ALLOWED_ORIGINS"
	defaultCORSAllowedOrigins = "*"

	shutdownGracePeriod = 30 * time.Second
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	logger := newLogger(logLevel)

	metricsProvider, nrLogCore, err := newMetricsProvider(
		os.Getenv("APP_NAME"),
		os.Getenv("NEW_RELIC_LICENSE_KEY"),
		logLevel,
	)
	if err != nil {
		logger.Fatal("failed to initialize metrics provider", zap.Error(err))
	}
	if nrLogCore != nil {
		logger = zap.New(nrLogCore)
	}

	dataProvider, err := indexerapp.NewDataProvider()
	if err != nil {
		logger.Fatal("failed to create data provider", zap.Error(err))
	}

	indexerServer := rpc.NewServer(logger, dataProvider.Ram)

	addr := os.Getenv(listenAddressEnv)
	if addr == "" {
		addr = defaultListenAddress
	}

	mux := http.NewServeMux()

	path, handler := indexerv1connect.NewIndexerHandler(
		&connectAdapter{grpcServer: indexerServer},
		connect.WithInterceptors(grpcCompatInterceptor(logger, safeMetricsProvider(metricsProvider))),
	)
	mux.Handle(path, handler)

	// The Connect handler natively speaks gRPC, gRPC-Web, and Connect for the
	// Indexer service. Health probes (grpc_health_probe) use the standard
	// grpc.health.v1 package, which we register on a throwaway grpc.Server
	// and serve via its experimental ServeHTTP so it shares the same port.
	healthSrv := grpc.NewServer()
	health_grpc.RegisterHealthServer(healthSrv, health.NewServer())
	mux.Handle("/grpc.health.v1.Health/", healthSrv)

	allowedOrigins := os.Getenv(corsAllowedOriginsEnv)
	if allowedOrigins == "" {
		allowedOrigins = defaultCORSAllowedOrigins
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(withCORS(mux, allowedOrigins), &http2.Server{}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting server", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case sig := <-sigCh:
		logger.Info("received signal, shutting down", zap.Stringer("signal", sig))
	case err := <-errCh:
		if err != nil {
			logger.Fatal("server exited unexpectedly", zap.Error(err))
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("http server shutdown error", zap.Error(err))
	}
	healthSrv.GracefulStop()
}

func newLogger(level string) *zap.Logger {
	return zap.New(newLogCore(level))
}

func newLogCore(level string) zapcore.Core {
	config := zap.NewProductionEncoderConfig()
	config.StacktraceKey = ""
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.AddSync(os.Stdout),
		parseLogLevel(level),
	)
}

func parseLogLevel(v string) zapcore.Level {
	switch strings.ToLower(v) {
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	}
	return zap.InfoLevel
}

func newMetricsProvider(appName, licenseKey, logLevel string) (metrics.Provider, zapcore.Core, error) {
	if licenseKey == "" {
		return noop_metrics.NewProvider(), nil, nil
	}

	nr, err := newrelic.NewApplication(
		newrelic.ConfigFromEnvironment(),
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(licenseKey),
		newrelic.ConfigDistributedTracerEnabled(true),
		newrelic.ConfigAppLogForwardingEnabled(true),
		newrelic.ConfigAppLogForwardingLabelsEnabled(true),
	)
	if err != nil {
		return nil, nil, err
	}

	provider := newrelic_metrics.NewProvider(nr)

	nrCore, err := nrzap.WrapBackgroundCore(newLogCore(logLevel), provider.Application())
	if err != nil {
		return nil, nil, err
	}
	return provider, nrCore, nil
}

// withCORS applies CORS headers required for browser-based Connect and
// gRPC-Web clients. allowedOrigins is a comma-separated list; "*" allows any
// origin. Preflight (OPTIONS) requests are answered directly.
func withCORS(next http.Handler, allowedOrigins string) http.Handler {
	origins := make(map[string]struct{})
	wildcard := false
	for _, o := range strings.Split(allowedOrigins, ",") {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if o == "*" {
			wildcard = true
			continue
		}
		origins[o] = struct{}{}
	}

	allowHeaders := strings.Join([]string{
		"Content-Type",
		"Connect-Protocol-Version",
		"Connect-Timeout-Ms",
		"Grpc-Timeout",
		"X-Grpc-Web",
		"X-User-Agent",
	}, ", ")
	exposeHeaders := strings.Join([]string{
		"Grpc-Status",
		"Grpc-Message",
		"Grpc-Status-Details-Bin",
	}, ", ")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if wildcard {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if _, ok := origins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
			w.Header().Set("Access-Control-Max-Age", "7200")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// connectAdapter bridges the existing gRPC indexerpb.IndexerServer into the
// indexerv1connect.IndexerHandler interface, so the same business logic is
// served over both transports.
type connectAdapter struct {
	grpcServer indexerpb.IndexerServer
}

func (a *connectAdapter) GetVirtualTimelockAccounts(ctx context.Context, req *connect.Request[indexerpb.GetVirtualTimelockAccountsRequest]) (*connect.Response[indexerpb.GetVirtualTimelockAccountsResponse], error) {
	resp, err := a.grpcServer.GetVirtualTimelockAccounts(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *connectAdapter) GetVirtualDurableNonce(ctx context.Context, req *connect.Request[indexerpb.GetVirtualDurableNonceRequest]) (*connect.Response[indexerpb.GetVirtualDurableNonceResponse], error) {
	resp, err := a.grpcServer.GetVirtualDurableNonce(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *connectAdapter) SearchVirtualTimelockAccounts(ctx context.Context, req *connect.Request[indexerpb.SearchVirtualTimelockAccountsRequest]) (*connect.Response[indexerpb.SearchVirtualTimelockAccountsResponse], error) {
	resp, err := a.grpcServer.SearchVirtualTimelockAccounts(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// grpcCompatInterceptor wraps the ocp-server gRPC unary interceptor chain
// (headers -> metrics -> validation) so the Connect transport runs the same
// middleware as the gRPC server. Stream RPCs are not used by this service,
// so only the unary path is adapted.
func grpcCompatInterceptor(log *zap.Logger, provider metrics.Provider) connect.UnaryInterceptorFunc {
	chain := grpc_middleware.ChainUnaryServer(
		headers.UnaryServerInterceptor(),
		grpc_metrics.UnaryServerInterceptor(provider),
		validation.UnaryServerInterceptor(log),
	)

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Surface the Connect/HTTP request headers as gRPC incoming
			// metadata so headers.UnaryServerInterceptor can populate the
			// context the same way it does for gRPC requests.
			md := metadata.MD{}
			for k, vs := range req.Header() {
				md[strings.ToLower(k)] = append(md[strings.ToLower(k)], vs...)
			}
			ctx = metadata.NewIncomingContext(ctx, md)

			info := &grpc.UnaryServerInfo{FullMethod: req.Spec().Procedure}

			var connectResp connect.AnyResponse
			handler := func(ctx context.Context, _ interface{}) (interface{}, error) {
				resp, err := next(ctx, req)
				if err != nil {
					return nil, err
				}
				connectResp = resp
				return resp.Any(), nil
			}

			if _, err := chain(ctx, req.Any(), info, handler); err != nil {
				return nil, toConnectError(err)
			}
			return connectResp, nil
		}
	}
}

// toConnectError converts a gRPC status error returned by the chained
// interceptors into a connect.Error. Errors that already are connect errors
// (or wrap one) are passed through unchanged.
func toConnectError(err error) error {
	if err == nil {
		return nil
	}
	var ce *connect.Error
	if errors.As(err, &ce) {
		return err
	}
	s, ok := status.FromError(err)
	if !ok {
		return err
	}
	// gRPC and Connect codes share the same numeric values for 1..16.
	return connect.NewError(connect.Code(s.Code()), errors.New(s.Message()))
}

// safeMetricsProvider wraps a metrics.Provider so that Trace.SetResponse never
// returns nil. The ocp-server metrics interceptor calls
// `trace.SetResponse(nil).WriteHeader(...)`, which panics under the noop
// provider (whose SetResponse returns its argument unchanged). Wrapping makes
// the connect path robust regardless of which provider is configured.
func safeMetricsProvider(p metrics.Provider) metrics.Provider {
	if p == nil {
		return nil
	}
	return &safeProvider{inner: p}
}

type safeProvider struct {
	inner metrics.Provider
}

func (p *safeProvider) StartTrace(name string) metrics.Trace {
	return &safeTrace{inner: p.inner.StartTrace(name)}
}

func (p *safeProvider) RecordEvent(eventName string, attributes map[string]interface{}) {
	p.inner.RecordEvent(eventName, attributes)
}

func (p *safeProvider) RecordCount(metricName string, count uint64) {
	p.inner.RecordCount(metricName, count)
}

func (p *safeProvider) RecordDuration(metricName string, duration time.Duration) {
	p.inner.RecordDuration(metricName, duration)
}

type safeTrace struct {
	inner metrics.Trace
}

func (t *safeTrace) StartSpan(name string) metrics.Span   { return t.inner.StartSpan(name) }
func (t *safeTrace) AddAttribute(k string, v interface{}) { t.inner.AddAttribute(k, v) }
func (t *safeTrace) OnError(err error)                    { t.inner.OnError(err) }
func (t *safeTrace) SetRequest(r metrics.Request)         { t.inner.SetRequest(r) }
func (t *safeTrace) End()                                 { t.inner.End() }

func (t *safeTrace) SetResponse(w http.ResponseWriter) http.ResponseWriter {
	if w == nil {
		w = discardResponseWriter{header: http.Header{}}
	}
	rw := t.inner.SetResponse(w)
	if rw == nil {
		return w
	}
	return rw
}

// discardResponseWriter is an http.ResponseWriter that discards everything.
// Used as a placeholder when the metrics interceptor needs a writer but no
// real HTTP response is available (e.g., when the connect transport is
// recording status before the handler-level writer is exposed).
type discardResponseWriter struct {
	header http.Header
}

func (d discardResponseWriter) Header() http.Header         { return d.header }
func (d discardResponseWriter) Write(b []byte) (int, error) { return io.Discard.Write(b) }
func (d discardResponseWriter) WriteHeader(int)             {}
