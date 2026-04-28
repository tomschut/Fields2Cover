// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Package health provides /healthz (liveness) and /readyz (readiness)
// HTTP handlers. The readiness probe is supplied by the server at
// startup — this package has zero dependencies on internal/f2cclient
// to keep it independently testable.

package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Probe is the readiness probe contract. The server wires in a
// closure that issues a cheap RPC against the f2c-grpc shim (e.g.
// CloneRobot) and returns nil on success.
//
// Probe receives a context with a short deadline (defaulting to 1s)
// so a hung shim does not stall the readiness endpoint.
type Probe func(ctx context.Context) error

// ProbeTimeout is the per-call deadline applied to the readiness
// probe. Override for tests via NewReadyz directly.
const ProbeTimeout = 1 * time.Second

type statusBody struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Healthz is the liveness endpoint. It always returns 200 OK with
// {"status":"ok"}. Use it for kubelet liveness probes — anything that
// kills the process should also fail this endpoint, but it makes no
// downstream calls itself.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, statusBody{Status: "ok"})
}

// NewReadyz returns an http.HandlerFunc that runs probe with the
// supplied timeout. probe == nil means the server has not finished
// startup yet — the handler returns 503 immediately.
func NewReadyz(probe Probe, timeout time.Duration) http.HandlerFunc {
	if timeout <= 0 {
		timeout = ProbeTimeout
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if probe == nil {
			writeJSON(w, http.StatusServiceUnavailable, statusBody{
				Status: "starting",
				Error:  "probe not configured",
			})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		if err := probe(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, statusBody{
				Status: "not_ready",
				Error:  err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, statusBody{Status: "ready"})
	}
}

func writeJSON(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
