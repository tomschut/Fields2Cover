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

const validSwathsBody = `{
  "field_with_headlands": {
    "field": {
      "id": "fld_test_001",
      "crs": {"epsg": 32631},
      "geometry": ` + validFeatureCollectionJSON + `
    },
    "headlands": ` + validFeatureCollectionJSON + `,
    "inner_field": ` + validFeatureCollectionJSON + `,
    "headland_width_m": 3.0
  },
  "robot": {"width_m": 3.0, "cov_width_m": 2.7},
  "angle_rad": 0.0,
  "width_m": 2.7
}`

func TestGenerateSwaths_Happy(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		generateSwaths: func(_ context.Context, req *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			if req.FieldWithHeadlands == nil || req.Robot == nil {
				t.Errorf("mock: missing fwh or robot")
			}
			return &f2cclient.GenerateSwathsResponse{
				Swaths: &f2cclient.Swaths{
					Field: &f2cclient.FieldWithHeadlands{
						Field: &f2cclient.Field{Id: "fld_test_001", Crs: &f2cclient.CRS{Epsg: 32631}},
					},
					Items: []*f2cclient.Swath{
						{Id: 0, WidthM: 2.7, Type: f2cclient.SwathType_SWATH_TYPE_MAINLAND, PathWkt: []byte(`{"type":"LineString","coordinates":[[0,0],[10,0]]}`)},
						{Id: 1, WidthM: 2.7, Type: f2cclient.SwathType_SWATH_TYPE_MAINLAND, PathWkt: []byte(`{"type":"LineString","coordinates":[[0,3],[10,3]]}`)},
						{Id: 2, WidthM: 2.7, Type: f2cclient.SwathType_SWATH_TYPE_MAINLAND, PathWkt: []byte(`{"type":"LineString","coordinates":[[0,6],[10,6]]}`)},
					},
				},
			}, nil
		},
	}
	srv := NewServer(mock)
	handler := func(w http.ResponseWriter, r *http.Request) {
		srv.GenerateSwaths(w, r)
	}
	rec := doRequest(t, handler, http.MethodPost, "/swaths:generate", validSwathsBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp api.GenerateSwathsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Swaths.Items) != 3 {
		t.Errorf("Items len = %d, want 3", len(resp.Swaths.Items))
	}
	if string(resp.Journey.StepName) != "swaths:generate" {
		t.Errorf("StepName = %q", resp.Journey.StepName)
	}
	if resp.Journey.Next == nil || resp.Journey.Next.Uri != "/swaths:sort" {
		t.Errorf("Journey.Next missing or wrong uri: %+v", resp.Journey.Next)
	}
}

func TestGenerateSwaths_GrpcInvalidArgument(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		generateSwaths: func(_ context.Context, _ *f2cclient.GenerateSwathsRequest) (*f2cclient.GenerateSwathsResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "no swaths generated")
		},
	}
	srv := NewServer(mock)
	handler := func(w http.ResponseWriter, r *http.Request) {
		srv.GenerateSwaths(w, r)
	}
	rec := doRequest(t, handler, http.MethodPost, "/swaths:generate", validSwathsBody)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "no swaths generated") {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestGenerateSwaths_BadJSON(t *testing.T) {
	t.Parallel()
	srv := NewServer(&mockF2CClient{})
	handler := func(w http.ResponseWriter, r *http.Request) {
		srv.GenerateSwaths(w, r)
	}
	rec := doRequest(t, handler, http.MethodPost, "/swaths:generate", "not json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
