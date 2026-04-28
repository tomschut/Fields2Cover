// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

const validRouteBody = `{
  "swaths": {
    "field": {
      "field": {
        "id": "fld_test_001",
        "crs": {"epsg": 32631},
        "geometry": ` + validFeatureCollectionJSON + `
      },
      "headlands": ` + validFeatureCollectionJSON + `,
      "inner_field": ` + validFeatureCollectionJSON + `,
      "headland_width_m": 3.0
    },
    "items": [
      {"id": 0, "width_m": 2.7, "type": "MAINLAND", "path": {"type":"LineString","coordinates":[[0,0],[10,0]]}},
      {"id": 1, "width_m": 2.7, "type": "MAINLAND", "path": {"type":"LineString","coordinates":[[0,3],[10,3]]}}
    ]
  }
}`

func TestGenerateRoute_Happy(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		generateRoute: func(_ context.Context, req *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			if req.Swaths == nil {
				t.Errorf("mock: nil swaths")
			}
			return &f2cclient.GenerateRouteResponse{
				Route: &f2cclient.Route{
					Field: &f2cclient.FieldWithHeadlands{
						Field: &f2cclient.Field{Id: "fld_test_001", Crs: &f2cclient.CRS{Epsg: 32631}},
					},
					Swaths: &f2cclient.Swaths{
						Field: &f2cclient.FieldWithHeadlands{
							Field: &f2cclient.Field{Id: "fld_test_001", Crs: &f2cclient.CRS{Epsg: 32631}},
						},
						Items: []*f2cclient.Swath{},
					},
					Connections: []*f2cclient.RouteConnection{
						{FromSwathId: 0, ToSwathId: 1, PointsWkt: []byte(`{"type":"MultiPoint","coordinates":[[10,0],[10,3]]}`)},
					},
				},
			}, nil
		},
	}
	srv := NewServer(mock)
	rec := doRequest(t, srv.GenerateRoute, http.MethodPost, "/routes:generate", validRouteBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp api.GenerateRouteResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Route.Connections) != 1 {
		t.Errorf("Connections len = %d, want 1", len(resp.Route.Connections))
	}
	if string(resp.Journey.StepName) != "routes:generate" {
		t.Errorf("StepName = %q", resp.Journey.StepName)
	}
	if resp.Journey.Next == nil {
		t.Errorf("Journey.Next is nil")
	}
}

func TestGenerateRoute_GrpcError(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		generateRoute: func(_ context.Context, _ *f2cclient.GenerateRouteRequest) (*f2cclient.GenerateRouteResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "no swaths to route")
		},
	}
	srv := NewServer(mock)
	rec := doRequest(t, srv.GenerateRoute, http.MethodPost, "/routes:generate", validRouteBody)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGenerateRoute_BadJSON(t *testing.T) {
	t.Parallel()
	srv := NewServer(&mockF2CClient{})
	rec := doRequest(t, srv.GenerateRoute, http.MethodPost, "/routes:generate", "no")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
