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

const validPlanBody = `{
  "route": {
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
      "items": []
    },
    "connections": []
  },
  "robot": {"width_m": 3.0, "cov_width_m": 2.7, "min_turning_radius_m": 4.5},
  "turning_algorithm": "DUBINS"
}`

func TestPlanPath_Happy(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		planPath: func(_ context.Context, req *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			if req.TurningAlgorithm != f2cclient.TurningAlgorithm_TURNING_ALGORITHM_DUBINS {
				t.Errorf("mock: turning algo = %v, want DUBINS", req.TurningAlgorithm)
			}
			return &f2cclient.PlanPathResponse{
				Path: &f2cclient.Path{
					Field: &f2cclient.FieldWithHeadlands{
						Field: &f2cclient.Field{Id: "fld_test_001", Crs: &f2cclient.CRS{Epsg: 32631}},
					},
					LengthM:   42.0,
					TaskTimeS: 18.0,
					States: []*f2cclient.PathState{
						{
							Point:       &f2cclient.Point{X: 0, Y: 0},
							AngleRad:    0,
							LengthM:     1,
							Velocity:    1.5,
							Direction:   f2cclient.PathDirection_PATH_DIRECTION_FORWARD,
							SectionType: f2cclient.PathSectionType_PATH_SECTION_TYPE_SWATH,
						},
					},
				},
			}, nil
		},
	}
	srv := NewServer(mock)
	rec := doRequest(t, srv.PlanPath, http.MethodPost, "/paths:plan", validPlanBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp api.PlanPathResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(resp.Journey.StepName) != "paths:plan" {
		t.Errorf("StepName = %q", resp.Journey.StepName)
	}
	// Terminal step contract: Next must be nil.
	if resp.Journey.Next != nil {
		t.Errorf("Journey.Next = %+v, want nil (terminal step)", resp.Journey.Next)
	}
	if resp.Journey.Back == nil {
		t.Errorf("Journey.Back is nil, want non-nil")
	}
	if len(resp.Path.States) != 1 {
		t.Errorf("Path.States len = %d, want 1", len(resp.Path.States))
	}
}

func TestPlanPath_BadJSON(t *testing.T) {
	t.Parallel()
	srv := NewServer(&mockF2CClient{})
	rec := doRequest(t, srv.PlanPath, http.MethodPost, "/paths:plan", "{{{")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestPlanPath_GrpcDeadlineExceeded(t *testing.T) {
	t.Parallel()
	mock := &mockF2CClient{
		planPath: func(_ context.Context, _ *f2cclient.PlanPathRequest) (*f2cclient.PlanPathResponse, error) {
			return nil, status.Error(codes.DeadlineExceeded, "shim too slow")
		},
	}
	srv := NewServer(mock)
	rec := doRequest(t, srv.PlanPath, http.MethodPost, "/paths:plan", validPlanBody)
	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want 504", rec.Code)
	}
}
