// Copyright (C) 2026 Wageningen University — BSD-3-Clause

//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
)

// --- /pipeline/plan-coverage -------------------------------------------

func TestShortcut_PlanCoverage_Happy(t *testing.T) {
	fc := loadFeatureCollection(t, "small_field.geojson")

	req := api.PlanCoverageRequest{
		Geojson:          fc,
		Robot:            defaultRobot(),
		HeadlandWidthM:   3.0,
		SwathWidthM:      2.0,
		TurningAlgorithm: api.PlanCoverageRequestTurningAlgorithmDUBINS,
	}
	resp, body := postJSON(t, baseURL(t)+"/pipeline/plan-coverage", req)
	assertStatus(t, resp, body, http.StatusOK)

	var out api.PlanCoverageResponse
	decode(t, body, &out)

	if len(out.Path.States) == 0 {
		t.Errorf("path.states empty; body len=%d", len(body))
	}
	if out.Path.LengthM <= 0 {
		t.Errorf("path.length_m=%v; want > 0", out.Path.LengthM)
	}
	if out.Intermediates.Field == nil {
		t.Errorf("intermediates.field is nil")
	}
	if out.Intermediates.FieldWithHeadlands == nil {
		t.Errorf("intermediates.field_with_headlands is nil")
	}
	if out.Intermediates.Swaths == nil {
		t.Errorf("intermediates.swaths is nil")
	}
	if out.Intermediates.SortedSwaths == nil {
		t.Errorf("intermediates.sorted_swaths is nil")
	}
	if out.Intermediates.Route == nil {
		t.Errorf("intermediates.route is nil")
	}
	if out.Journey.Next != nil {
		t.Errorf("PlanCoverage is terminal; Journey.Next must be nil, got %+v", out.Journey.Next)
	}
}

func TestShortcut_PlanCoverage_Happy_MultiCell(t *testing.T) {
	fc := loadFeatureCollection(t, "multi_cell.geojson")

	req := api.PlanCoverageRequest{
		Geojson:          fc,
		Robot:            defaultRobot(),
		HeadlandWidthM:   3.0,
		SwathWidthM:      2.0,
		TurningAlgorithm: api.PlanCoverageRequestTurningAlgorithmDUBINS,
	}
	resp, body := postJSON(t, baseURL(t)+"/pipeline/plan-coverage", req)
	assertStatus(t, resp, body, http.StatusOK)

	var out api.PlanCoverageResponse
	decode(t, body, &out)

	if len(out.Path.States) == 0 {
		t.Errorf("multi-cell path.states empty; body len=%d", len(body))
	}
	if out.Path.LengthM <= 0 {
		t.Errorf("multi-cell path.length_m=%v; want > 0", out.Path.LengthM)
	}
	if out.Intermediates.Field == nil || out.Intermediates.Route == nil {
		t.Errorf("multi-cell intermediates not fully populated")
	}
}

func TestShortcut_PlanCoverage_Error_EmptyFeatures(t *testing.T) {
	fc := loadFeatureCollection(t, "empty_features.geojson")

	req := api.PlanCoverageRequest{
		Geojson:          fc,
		Robot:            defaultRobot(),
		HeadlandWidthM:   3.0,
		SwathWidthM:      2.0,
		TurningAlgorithm: api.PlanCoverageRequestTurningAlgorithmDUBINS,
	}
	resp, body := postJSON(t, baseURL(t)+"/pipeline/plan-coverage", req)
	if resp.StatusCode < 400 {
		t.Fatalf("expected error status, got %d; body=%s", resp.StatusCode, body)
	}
	var errResp api.Error
	decode(t, body, &errResp)
	if errResp.Code == "" {
		t.Errorf("error response has empty code; body=%s", body)
	}
}

// --- /pipeline/resume --------------------------------------------------

func TestShortcut_ResumePipeline_Error_ParseFieldStartRejected(t *testing.T) {
	// "parse-field" is not in the generated enum; pass the raw string
	// value through the untyped wire so the server's defensive reject
	// branch is exercised.
	rawBody := `{"start_step":"parse-field","robot":{"width_m":2,"cov_width_m":2},"turning_algorithm":"DUBINS"}`
	resp, body := postJSONRaw(t, baseURL(t)+"/pipeline/resume", rawBody)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d want 400; body=%s", resp.StatusCode, body)
	}
	var e api.Error
	decode(t, body, &e)
	if e.Code != "INVALID_START_STEP" {
		t.Errorf("error code=%q want INVALID_START_STEP", e.Code)
	}
}

func TestShortcut_ResumePipeline_Error_MissingPayload(t *testing.T) {
	req := api.ResumePipelineRequest{
		StartStep:        api.SortSwaths,
		Robot:            defaultRobot(),
		TurningAlgorithm: api.DUBINS,
		// No Swaths — should trigger MISSING_PAYLOAD
	}
	resp, body := postJSON(t, baseURL(t)+"/pipeline/resume", req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d want 400; body=%s", resp.StatusCode, body)
	}
	var e api.Error
	decode(t, body, &e)
	if e.Code != "MISSING_PAYLOAD" {
		t.Errorf("error code=%q want MISSING_PAYLOAD; body=%s", e.Code, body)
	}
}

// TestShortcut_ResumePipeline_Happy_FromSortSwaths_SentinelPreserved
// is the SHC-03 operational proof test. It drives the granular
// endpoints up to unsorted swaths, mutates the first swath's id to a
// sentinel value (9999), then POSTs to /pipeline/resume with
// start_step=sort-swaths. A correct implementation must NOT re-run
// parse-field / generate-headland / generate-swaths — those would
// overwrite the sentinel with a freshly-generated id. The assertions
// verify the sentinel survived and that intermediates.field /
// field_with_headlands are nil (earlier steps did not run).
func TestShortcut_ResumePipeline_Happy_FromSortSwaths_SentinelPreserved(t *testing.T) {
	fc := loadFeatureCollection(t, "small_field.geojson")

	// 1. Drive the granular endpoints up to "unsorted swaths".
	f := mustParseField(t, fc)
	fwh := mustGenerateHeadlands(t, f)

	swathsReq := api.GenerateSwathsRequest{
		FieldWithHeadlands: fwh,
		Robot:              defaultRobot(),
		AngleRad:           0,
		WidthM:             2.0,
	}
	resp, body := postJSON(t, baseURL(t)+"/swaths:generate", swathsReq)
	assertStatus(t, resp, body, http.StatusOK)
	var swathsOut api.GenerateSwathsResponse
	decode(t, body, &swathsOut)

	if len(swathsOut.Swaths.Items) == 0 {
		t.Fatalf("no swaths generated; cannot set sentinel")
	}

	// 2. Sentinel mutation on the first swath.
	swathsOut.Swaths.Items[0].Id = 9999

	// 3. Resume from sort-swaths with the mutated payload.
	resumeReq := api.ResumePipelineRequest{
		StartStep:        api.SortSwaths,
		Swaths:           &swathsOut.Swaths,
		Robot:            defaultRobot(),
		SwathWidthM:      float64Ptr(2.0),
		TurningAlgorithm: api.DUBINS,
	}
	resp, body = postJSON(t, baseURL(t)+"/pipeline/resume", resumeReq)
	assertStatus(t, resp, body, http.StatusOK)

	var out api.PlanCoverageResponse
	decode(t, body, &out)

	// 4. Assertions.
	if out.Intermediates.Swaths == nil {
		t.Fatalf("intermediates.swaths nil; body len=%d", len(body))
	}
	if len(out.Intermediates.Swaths.Items) == 0 {
		t.Fatalf("intermediates.swaths.items empty")
	}
	// Find the sentinel — ResumePayload echoes the caller's pre-sort
	// swaths back in intermediates.swaths verbatim (order unchanged).
	if out.Intermediates.Swaths.Items[0].Id != 9999 {
		t.Errorf("sentinel LOST: items[0].id=%d, want 9999. Earlier steps re-ran — SHC-03 violated",
			out.Intermediates.Swaths.Items[0].Id)
	}
	if out.Intermediates.SortedSwaths == nil {
		t.Errorf("intermediates.sorted_swaths nil; sort step did not run")
	}
	if out.Intermediates.Route == nil {
		t.Errorf("intermediates.route nil; route step did not run")
	}
	if len(out.Path.States) == 0 {
		t.Errorf("path.states empty; plan step did not run")
	}

	// These MUST be nil — earlier-step outputs the server should not have produced.
	if out.Intermediates.Field != nil {
		t.Errorf("intermediates.field is non-nil; parse-field step should not have run on resume")
	}
	if out.Intermediates.FieldWithHeadlands != nil {
		t.Errorf("intermediates.field_with_headlands is non-nil; headland step should not have run on resume")
	}
}
