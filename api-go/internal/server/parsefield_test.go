// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

const validParseFieldBody = `{
  "geojson": ` + validFeatureCollectionJSON + `
}`

func TestParseField_Happy(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, req *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			if req.Geojson == "" {
				t.Errorf("mock: empty geojson in request")
			}
			return &f2cclient.ParseGeoJSONResponse{
				Field: &f2cclient.Field{
					Id: "fld_test_001",
					Crs: &f2cclient.CRS{
						Epsg:    32631,
						UtmZone: "31N",
					},
					GeometryWkt: []byte(`{"type":"FeatureCollection","features":[]}`),
				},
			}, nil
		},
	}
	srv := NewServer(mock)

	rec := doRequest(t, srv.ParseField, http.MethodPost, "/fields:parse", validParseFieldBody)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var resp api.ParseFieldResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Field.Id != "fld_test_001" {
		t.Errorf("Field.Id = %q, want fld_test_001", resp.Field.Id)
	}
	if string(resp.Journey.StepName) != "fields:parse" {
		t.Errorf("Journey.StepName = %q", resp.Journey.StepName)
	}
	if resp.Journey.Next == nil {
		t.Fatal("Journey.Next is nil")
	}
	if resp.Journey.Next.Uri != "/headlands:generate" {
		t.Errorf("Journey.Next.Uri = %q, want /headlands:generate", resp.Journey.Next.Uri)
	}
}

func TestParseField_BadJSON(t *testing.T) {
	t.Parallel()
	srv := NewServer(&mockF2CClient{})
	rec := doRequest(t, srv.ParseField, http.MethodPost, "/fields:parse", "not json{{{")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "INVALID_JSON") {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestParseField_GrpcInvalidArgument(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "empty feature collection")
		},
	}
	srv := NewServer(mock)
	rec := doRequest(t, srv.ParseField, http.MethodPost, "/fields:parse", validParseFieldBody)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "empty feature collection") {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestParseField_GrpcInternal(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			return nil, status.Error(codes.Internal, "shim crashed")
		},
	}
	srv := NewServer(mock)
	rec := doRequest(t, srv.ParseField, http.MethodPost, "/fields:parse", validParseFieldBody)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestParseField_EmptyResponse(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		parseGeoJSON: func(_ context.Context, _ *f2cclient.ParseGeoJSONRequest) (*f2cclient.ParseGeoJSONResponse, error) {
			return &f2cclient.ParseGeoJSONResponse{Field: nil}, nil
		},
	}
	srv := NewServer(mock)
	rec := doRequest(t, srv.ParseField, http.MethodPost, "/fields:parse", validParseFieldBody)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "EMPTY_RESPONSE") {
		t.Errorf("body = %s", rec.Body.String())
	}
}
