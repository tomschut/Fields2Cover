// Copyright (C) 2026 Wageningen University — BSD-3-Clause

//go:build integration

package integration

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
)

// --- /fields:parse ------------------------------------------------------

func TestJourney_ParseField_Happy(t *testing.T) {
	fc := loadFeatureCollection(t, "small_field.geojson")
	req := api.ParseFieldRequest{Geojson: fc}

	resp, body := postJSON(t, baseURL(t)+"/fields:parse", req)
	assertStatus(t, resp, body, http.StatusOK)

	var out api.ParseFieldResponse
	decode(t, body, &out)

	if out.Field.Id == "" {
		t.Errorf("Field.Id is empty; body=%s", body)
	}
	if len(out.Field.Geometry.Features) == 0 {
		t.Errorf("Field.Geometry.Features is empty; body=%s", body)
	}
	if out.Journey.Next == nil {
		t.Fatalf("Journey.Next is nil on /fields:parse; want headlands:generate link")
	}
	if !strings.Contains(out.Journey.Next.Uri, "headlands:generate") {
		t.Errorf("Journey.Next.Uri=%q; want contains headlands:generate", out.Journey.Next.Uri)
	}
}

func TestJourney_ParseField_Error_EmptyFeatures(t *testing.T) {
	fc := loadFeatureCollection(t, "empty_features.geojson")
	req := api.ParseFieldRequest{Geojson: fc}

	resp, body := postJSON(t, baseURL(t)+"/fields:parse", req)
	if resp.StatusCode < 400 {
		t.Fatalf("expected error status, got %d; body=%s", resp.StatusCode, body)
	}
	var errResp api.Error
	decode(t, body, &errResp)
	if errResp.Code == "" {
		t.Errorf("error response has empty code; body=%s", body)
	}
}

// --- /headlands:generate -----------------------------------------------

func TestJourney_GenerateHeadlands_Happy(t *testing.T) {
	fc := loadFeatureCollection(t, "small_field.geojson")
	f := mustParseField(t, fc)

	req := api.GenerateHeadlandsRequest{
		Field:  f,
		Robot:  defaultRobot(),
		WidthM: 3.0,
	}
	resp, body := postJSON(t, baseURL(t)+"/headlands:generate", req)
	assertStatus(t, resp, body, http.StatusOK)

	var out api.GenerateHeadlandsResponse
	decode(t, body, &out)

	if len(out.FieldWithHeadlands.Headlands.Features) == 0 {
		t.Errorf("headlands empty; body=%s", body)
	}
	if len(out.FieldWithHeadlands.InnerField.Features) == 0 {
		t.Errorf("inner_field empty; body=%s", body)
	}
}

// --- /swaths:generate --------------------------------------------------

func TestJourney_GenerateSwaths_Happy(t *testing.T) {
	fc := loadFeatureCollection(t, "small_field.geojson")
	f := mustParseField(t, fc)
	fwh := mustGenerateHeadlands(t, f)

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

	if len(out.Swaths.Items) == 0 {
		t.Errorf("swaths.items empty; body=%s", body)
	}
}

// --- /swaths:sort ------------------------------------------------------

func TestJourney_SortSwaths_Happy(t *testing.T) {
	fc := loadFeatureCollection(t, "small_field.geojson")
	f := mustParseField(t, fc)
	fwh := mustGenerateHeadlands(t, f)
	sw := mustGenerateSwaths(t, f, fwh)

	req := api.SortSwathsRequest{Swaths: sw}
	resp, body := postJSON(t, baseURL(t)+"/swaths:sort", req)
	assertStatus(t, resp, body, http.StatusOK)

	var out api.SortSwathsResponse
	decode(t, body, &out)
	if len(out.Swaths.Items) != len(sw.Items) {
		t.Errorf("sorted count=%d, want %d", len(out.Swaths.Items), len(sw.Items))
	}
}

func TestJourney_SortSwaths_Error_BadJSON(t *testing.T) {
	resp, body := postJSONRaw(t, baseURL(t)+"/swaths:sort", "{not json")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d want 400; body=%s", resp.StatusCode, body)
	}
	var errResp api.Error
	decode(t, body, &errResp)
	if errResp.Code != "INVALID_JSON" {
		t.Errorf("error code=%q want INVALID_JSON; body=%s", errResp.Code, body)
	}
}

// --- /routes:generate --------------------------------------------------

func TestJourney_GenerateRoute_Happy(t *testing.T) {
	fc := loadFeatureCollection(t, "small_field.geojson")
	f := mustParseField(t, fc)
	fwh := mustGenerateHeadlands(t, f)
	sw := mustGenerateSwaths(t, f, fwh)
	sorted := mustSortSwaths(t, sw)

	req := api.GenerateRouteRequest{Swaths: sorted}
	resp, body := postJSON(t, baseURL(t)+"/routes:generate", req)
	assertStatus(t, resp, body, http.StatusOK)

	var out api.GenerateRouteResponse
	decode(t, body, &out)
	if len(out.Route.Swaths.Items) == 0 {
		t.Errorf("route.swaths.items empty; body=%s", body)
	}
}

func TestJourney_GenerateRoute_Error_BadJSON(t *testing.T) {
	resp, body := postJSONRaw(t, baseURL(t)+"/routes:generate", "{not json")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d want 400; body=%s", resp.StatusCode, body)
	}
	var errResp api.Error
	decode(t, body, &errResp)
	if errResp.Code != "INVALID_JSON" {
		t.Errorf("error code=%q want INVALID_JSON; body=%s", errResp.Code, body)
	}
}

// --- /paths:plan -------------------------------------------------------

func TestJourney_PlanPath_Happy(t *testing.T) {
	fc := loadFeatureCollection(t, "small_field.geojson")
	f := mustParseField(t, fc)
	fwh := mustGenerateHeadlands(t, f)
	sw := mustGenerateSwaths(t, f, fwh)
	sorted := mustSortSwaths(t, sw)
	route := mustGenerateRoute(t, sorted)

	req := api.PlanPathRequest{
		Route:            route,
		Robot:            defaultRobot(),
		TurningAlgorithm: api.PlanPathRequestTurningAlgorithmDUBINS,
	}
	resp, body := postJSON(t, baseURL(t)+"/paths:plan", req)
	assertStatus(t, resp, body, http.StatusOK)

	var out api.PlanPathResponse
	decode(t, body, &out)

	if len(out.Path.States) == 0 {
		t.Errorf("path.states empty; body=%s", body)
	}
	if out.Path.LengthM <= 0 {
		t.Errorf("path.length_m=%v; want > 0", out.Path.LengthM)
	}
	if out.Journey.Next != nil {
		t.Errorf("PlanPath is terminal; Journey.Next must be nil, got %+v", out.Journey.Next)
	}
}

func TestJourney_PlanPath_Error_BadJSON(t *testing.T) {
	resp, body := postJSONRaw(t, baseURL(t)+"/paths:plan", "{not json")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d want 400; body=%s", resp.StatusCode, body)
	}
	var errResp api.Error
	decode(t, body, &errResp)
	if errResp.Code != "INVALID_JSON" {
		t.Errorf("error code=%q want INVALID_JSON; body=%s", errResp.Code, body)
	}
}
