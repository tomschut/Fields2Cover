// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Command api is the Fields2Cover Go API server entry point. It loads
// configuration from environment variables, dials the f2c-grpc shim
// over a Unix domain socket, mounts the journey handlers (generated
// from openapi.yaml) on a chi router with logging + health
// middleware, listens on :8080, and exits cleanly on SIGTERM / SIGINT.
//
// Configuration (env vars):
//
//	API_PORT          Listen port (default: 8080)
//	F2C_GRPC_SOCKET   Path to the f2c-grpc-server Unix socket
//	                  (default: /tmp/f2c.sock)
//	SHUTDOWN_TIMEOUT  Max seconds to drain in-flight requests on
//	                  SIGTERM (default: 10)
//	API_LOG_LEVEL     debug | info | warn | error (default: info)
//	API_LOG_FORMAT    json | text (default: json)

package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
	"github.com/Fields2Cover/fields2cover/api-go/internal/health"
	"github.com/Fields2Cover/fields2cover/api-go/internal/logging"
	"github.com/Fields2Cover/fields2cover/api-go/internal/server"
	"github.com/Fields2Cover/fields2cover/api-go/internal/static"
)

type config struct {
	port            string
	socketPath      string
	shutdownTimeout time.Duration
	logLevel        string
	logFormat       string
	openapiPath     string
}

func loadConfig() config {
	c := config{
		port:            envOr("API_PORT", "8080"),
		socketPath:      envOr("F2C_GRPC_SOCKET", "/tmp/f2c.sock"),
		shutdownTimeout: 10 * time.Second,
		logLevel:        envOr("API_LOG_LEVEL", "info"),
		logFormat:       envOr("API_LOG_FORMAT", "json"),
		openapiPath:     envOr("OPENAPI_PATH", "openapi.yaml"),
	}
	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.shutdownTimeout = time.Duration(n) * time.Second
		}
	}
	return c
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// docsHTML is a self-hosted Swagger Editor page that loads the spec
// from /openapi.yaml (same origin) so there is no mixed-content issue
// when the API is reached over http from an https-hosted editor. The
// editor bundle itself is pulled from jsdelivr (https), which is fine
// as an https subresource on an http page.
const docsHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>f2c API — Swagger Editor</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-editor-dist@4/swagger-editor.css">
  <style>
    html, body, #swagger-editor { height: 100%; margin: 0; padding: 0; }
    /* swagger-editor-dist ships a react-split-pane layout whose right
       pane (rendered API) has no scroll container by default — long
       specs clip. Force vertical scroll on Pane2 and make sure every
       ancestor up to #swagger-editor keeps its 100% height. */
    #swagger-editor > div,
    #swagger-editor .Pane,
    #swagger-editor .Pane1,
    #swagger-editor .Pane2 { height: 100%; }
    #swagger-editor .Pane2 { overflow-y: auto !important; }
  </style>
</head>
<body>
<div id="swagger-editor"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-editor-dist@4/swagger-editor-bundle.js" charset="UTF-8"></script>
<script src="https://cdn.jsdelivr.net/npm/swagger-editor-dist@4/swagger-editor-standalone-preset.js" charset="UTF-8"></script>
<script>
  window.editor = SwaggerEditorBundle({
    dom_id: '#swagger-editor',
    layout: 'StandaloneLayout',
    presets: [SwaggerEditorStandalonePreset],
    url: '/openapi.yaml',
  });
</script>
</body>
</html>`

// buildRouter assembles the chi router with logging middleware,
// /healthz + /readyz, the generated journey handlers, and the
// self-hosted /docs Swagger Editor (+ /openapi.yaml). The readiness
// probe issues a CloneRobot RPC against the shim — it is the
// cheapest available probe per Phase 10 SUMMARY.
func buildRouter(logger *slog.Logger, srv *server.Server, client f2cclient.F2CClient, openapiPath string) http.Handler {
	r := chi.NewRouter()

	// Cross-cutting middleware: structured request logging + request_id.
	r.Use(logging.RequestLogger(logger))

	// Liveness + readiness.
	r.Get("/healthz", health.Healthz)

	probe := func(ctx context.Context) error {
		_, err := client.CloneRobot(ctx, &f2cclient.CloneRobotRequest{
			Robot: &f2cclient.Robot{WidthM: 1.0, CovWidthM: 1.0},
		})
		return err
	}
	r.Get("/readyz", health.NewReadyz(probe, health.ProbeTimeout))

	// /openapi.yaml + /docs: only mounted if the spec file is readable.
	// Serving from disk (not embedded) so edits to openapi.yaml are live
	// without a rebuild during development.
	if _, err := os.Stat(openapiPath); err == nil {
		r.Get("/openapi.yaml", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.ServeFile(w, req, openapiPath)
		})
		r.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(docsHTML))
		})
		logger.Info("mounted swagger editor", slog.String("spec", openapiPath), slog.String("path", "/docs"))
	} else {
		logger.Warn("openapi spec not found, /docs disabled", slog.String("path", openapiPath), slog.Any("err", err))
	}

	// Mount the generated journey handlers. HandlerFromMux is
	// emitted by oapi-codegen --generate chi-server.
	api.HandlerFromMux(srv, r)

	// SPA catch-all — registered LAST so all explicit API routes take priority.
	distFS, err := fs.Sub(static.FS, "dist")
	if err != nil {
		panic("static embed misconfigured: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(distFS))
	r.Get("/*", spaHandler(fileServer))
	logger.Info("mounted SPA catch-all", slog.String("path", "/*"))

	return r
}

// spaHandler wraps http.FileServer so paths that do not match a real
// embedded file fall back to serving index.html. This enables React
// Router's client-side routing (pushState URLs) to work on direct
// load or refresh without a 404.
func spaHandler(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}
		fs.ServeHTTP(rec, r)
		if rec.status == http.StatusNotFound {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fs.ServeHTTP(w, r2)
		}
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	if code != http.StatusNotFound {
		r.ResponseWriter.WriteHeader(code)
	}
}

// Write suppresses body bytes when a 404 has been signaled so that
// the file-server error body does not leak into the response before
// the SPA fallback (index.html) is written.
func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == http.StatusNotFound {
		return len(b), nil
	}
	return r.ResponseWriter.Write(b)
}

func run(ctx context.Context, cfg config, logger *slog.Logger) error {
	logger.Info("starting api-go server",
		slog.String("addr", ":"+cfg.port),
		slog.String("socket", cfg.socketPath))

	client, err := f2cclient.Dial(cfg.socketPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			logger.Warn("f2c client close error", slog.Any("err", cerr))
		}
	}()

	srv := server.NewServer(client)
	router := buildRouter(logger, srv, client, cfg.openapiPath)

	httpServer := &http.Server{
		Addr:              ":" + cfg.port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining requests",
			slog.Duration("timeout", cfg.shutdownTimeout))
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.shutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", slog.Any("err", err))
			return err
		}
		logger.Info("server stopped cleanly")
		return nil
	case err := <-serverErr:
		return err
	}
}

func main() {
	cfg := loadConfig()
	logger := logging.NewJSONLogger(os.Stdout, cfg.logFormat, cfg.logLevel)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("server exited with error", slog.Any("err", err))
		os.Exit(1)
	}
}
