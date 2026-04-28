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

const validHeadlandsBody = `{
  "field": {
    "id": "fld_test_001",
    "crs": {"epsg": 32631},
    "geometry": ` + validFeatureCollectionJSON + `
  },
  "robot": {"width_m": 3.0, "cov_width_m": 2.7},
  "width_m": 3.0,
  "count": 2
}`

func TestGenerateHeadlands_Happy(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		generateHeadland: func(_ context.Context, req *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			if req.Field == nil || req.Robot == nil {
				t.Errorf("mock: missing field or robot")
			}
			if req.Count != 2 {
				t.Errorf("mock: count = %d, want 2", req.Count)
			}
			return &f2cclient.GenerateHeadlandResponse{
				FieldWithHeadlands: &f2cclient.FieldWithHeadlands{
					Field: &f2cclient.Field{
						Id:  "fld_test_001",
						Crs: &f2cclient.CRS{Epsg: 32631},
					},
					HeadlandWidthM: 3.0,
					HeadlandsWkt:   []byte(`{"type":"FeatureCollection","features":[]}`),
					InnerFieldWkt:  []byte(`{"type":"FeatureCollection","features":[]}`),
				},
			}, nil
		},
	}
	srv := NewServer(mock)
	handler := func(w http.ResponseWriter, r *http.Request) {
		srv.GenerateHeadlands(w, r)
	}
	rec := doRequest(t, handler, http.MethodPost, "/headlands:generate", validHeadlandsBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp api.GenerateHeadlandsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.FieldWithHeadlands.Field.Id != "fld_test_001" {
		t.Errorf("Field.Id = %q", resp.FieldWithHeadlands.Field.Id)
	}
	if string(resp.Journey.StepName) != "headlands:generate" {
		t.Errorf("StepName = %q", resp.Journey.StepName)
	}
	if resp.Journey.Next == nil || resp.Journey.Next.Uri != "/swaths:generate" {
		t.Errorf("Journey.Next missing or wrong uri: %+v", resp.Journey.Next)
	}
}

func TestGenerateHeadlands_BadJSON(t *testing.T) {
	t.Parallel()
	srv := NewServer(&mockF2CClient{})
	handler := func(w http.ResponseWriter, r *http.Request) {
		srv.GenerateHeadlands(w, r)
	}
	rec := doRequest(t, handler, http.MethodPost, "/headlands:generate", "not json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "INVALID_JSON") {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestGenerateHeadlands_GrpcFailedPrecondition(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		generateHeadland: func(_ context.Context, _ *f2cclient.GenerateHeadlandRequest) (*f2cclient.GenerateHeadlandResponse, error) {
			return nil, status.Error(codes.FailedPrecondition, "field too small")
		},
	}
	srv := NewServer(mock)
	handler := func(w http.ResponseWriter, r *http.Request) {
		srv.GenerateHeadlands(w, r)
	}
	rec := doRequest(t, handler, http.MethodPost, "/headlands:generate", validHeadlandsBody)
	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("status = %d, want 412", rec.Code)
	}
}
