// Copyright (C) 2026 Wageningen University — BSD-3-Clause

//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
	"github.com/Fields2Cover/fields2cover/api-go/internal/health"
	"github.com/Fields2Cover/fields2cover/api-go/internal/pipeline"
	"github.com/Fields2Cover/fields2cover/api-go/internal/server"
)

// buildRouterForIntegration mirrors cmd/api/main.buildRouter. Must
// stay in sync manually — the integration suite itself is the sync
// check (if the production router adds a route the tests don't hit,
// nothing breaks; if it removes one, tests fail).
func buildRouterForIntegration(p *pipeline.Pipeline) http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", health.Healthz)

	client := p.Steps().Client()
	probe := func(ctx context.Context) error {
		_, err := client.CloneRobot(ctx, &f2cclient.CloneRobotRequest{
			Robot: &f2cclient.Robot{WidthM: 1.0, CovWidthM: 1.0},
		})
		return err
	}
	r.Get("/readyz", health.NewReadyz(probe, health.ProbeTimeout))

	srv := server.NewServerFromPipeline(p)
	api.HandlerFromMux(srv, r)
	return r
}

// --- HTTP helpers -------------------------------------------------------

// postJSON POSTs body as JSON and returns the response + body bytes.
func postJSON(t *testing.T, url string, body any) (*http.Response, []byte) {
	t.Helper()
	buf, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http do: %v", err)
	}
	out, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, out
}

// postJSONRaw posts a raw string body (for malformed-JSON error tests).
func postJSONRaw(t *testing.T, url, rawBody string) (*http.Response, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte(rawBody)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http do: %v", err)
	}
	out, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, out
}

// decode unmarshals body into dst, failing the test on error.
func decode(t *testing.T, body []byte, dst any) {
	t.Helper()
	if err := json.Unmarshal(body, dst); err != nil {
		t.Fatalf("decode %T: %v\nbody: %s", dst, err, string(body))
	}
}

// loadTestdata reads a testdata file relative to this package.
func loadTestdata(t *testing.T, name string) []byte {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(thisFile), "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("loadTestdata(%s): %v", name, err)
	}
	return data
}

// loadFeatureCollection reads a testdata file and decodes it as an
// api.GeoJSONFeatureCollection ready for embedding in request bodies.
func loadFeatureCollection(t *testing.T, name string) api.GeoJSONFeatureCollection {
	t.Helper()
	var fc api.GeoJSONFeatureCollection
	decode(t, loadTestdata(t, name), &fc)
	return fc
}

// assertStatus fails fast if the response code is unexpected.
func assertStatus(t *testing.T, resp *http.Response, body []byte, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("status = %d, want %d; body=%s", resp.StatusCode, want, string(body))
	}
}

// defaultRobot returns a minimal Robot spec suitable for every
// test endpoint. Uses 2m width, basic turning radius + velocities.
func defaultRobot() api.Robot {
	mtr := 3.0
	cv := 1.5
	tv := 0.5
	name := "integration-robot"
	return api.Robot{
		WidthM:            2.0,
		CovWidthM:         2.0,
		MinTurningRadiusM: &mtr,
		CruiseVelMps:      &cv,
		TurnVelMps:        &tv,
		Name:              &name,
	}
}

func stringPtr(s string) *string    { return &s }
func float64Ptr(f float64) *float64 { return &f }
func intPtr(i int) *int             { return &i }

// mustParseField runs /fields:parse and returns the resulting Field.
func mustParseField(t *testing.T, fc api.GeoJSONFeatureCollection) api.Field {
	t.Helper()
	req := api.ParseFieldRequest{Geojson: fc}
	resp, body := postJSON(t, baseURL(t)+"/fields:parse", req)
	assertStatus(t, resp, body, http.StatusOK)
	var out api.ParseFieldResponse
	decode(t, body, &out)
	return out.Field
}

// mustGenerateHeadlands chains field -> headlands.
func mustGenerateHeadlands(t *testing.T, f api.Field) api.FieldWithHeadlands {
	t.Helper()
	req := api.GenerateHeadlandsRequest{
		Field:  f,
		Robot:  defaultRobot(),
		WidthM: 3.0,
	}
	resp, body := postJSON(t, baseURL(t)+"/headlands:generate", req)
	assertStatus(t, resp, body, http.StatusOK)
	var out api.GenerateHeadlandsResponse
	decode(t, body, &out)
	return out.FieldWithHeadlands
}

// mustGenerateSwaths chains fwh -> swaths.
func mustGenerateSwaths(t *testing.T, f api.Field, fwh api.FieldWithHeadlands) api.Swaths {
	t.Helper()
	req := api.GenerateSwathsRequest{
		FieldWithHeadlands: fwh,
		Robot:              defaultRobot(),
		AngleRad:           0,
		WidthM:             2.0,
	}
	resp, body := postJSON(t, baseURL(t)+"/swaths:generate", req)
	assertStatus(t, resp, body, http.StatusOK)
	var out api.GenerateSwathsResponse
	decode(t, body, &out)
	return out.Swaths
}

// mustSortSwaths chains swaths -> sorted swaths.
func mustSortSwaths(t *testing.T, sw api.Swaths) api.Swaths {
	t.Helper()
	req := api.SortSwathsRequest{Swaths: sw}
	resp, body := postJSON(t, baseURL(t)+"/swaths:sort", req)
	assertStatus(t, resp, body, http.StatusOK)
	var out api.SortSwathsResponse
	decode(t, body, &out)
	return out.Swaths
}

// mustGenerateRoute chains sorted swaths -> route.
func mustGenerateRoute(t *testing.T, sw api.Swaths) api.Route {
	t.Helper()
	req := api.GenerateRouteRequest{Swaths: sw}
	resp, body := postJSON(t, baseURL(t)+"/routes:generate", req)
	assertStatus(t, resp, body, http.StatusOK)
	var out api.GenerateRouteResponse
	decode(t, body, &out)
	return out.Route
}

// unusedPlaceholders keeps the linter quiet for helpers used by only
// a subset of tests (e.g. stringPtr/intPtr). Gets inlined away.
var (
	_ = stringPtr
	_ = intPtr
)
