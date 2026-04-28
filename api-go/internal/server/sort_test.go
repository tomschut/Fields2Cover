// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

const validSortBody = `{
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
      {"id": 0, "width_m": 2.7, "type": "MAINLAND", "path": {"type":"LineString","coordinates":[[0,0],[10,0]]}}
    ]
  },
  "algorithm": "SNAKE"
}`

func TestSortSwaths_Happy(t *testing.T) {
	t.Parallel()
	sortedTrue := true
	mock := &mockF2CClient{
		sortSwaths: func(_ context.Context, req *f2cclient.SortSwathsRequest) (*f2cclient.SortSwathsResponse, error) {
			if req.Algorithm != f2cclient.SortAlgorithm_SORT_ALGORITHM_SNAKE {
				t.Errorf("mock: algorithm = %v, want SNAKE", req.Algorithm)
			}
			return &f2cclient.SortSwathsResponse{
				Swaths: &f2cclient.Swaths{
					Field: &f2cclient.FieldWithHeadlands{
						Field: &f2cclient.Field{Id: "fld_test_001", Crs: &f2cclient.CRS{Epsg: 32631}},
					},
					Items: []*f2cclient.Swath{
						{Id: 0, WidthM: 2.7, Type: f2cclient.SwathType_SWATH_TYPE_MAINLAND},
					},
					Sorted: true,
				},
			}, nil
		},
	}
	srv := NewServer(mock)
	rec := doRequest(t, srv.SortSwaths, http.MethodPost, "/swaths:sort", validSortBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp api.SortSwathsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Swaths.Sorted == nil || *resp.Swaths.Sorted != sortedTrue {
		t.Errorf("Sorted = %v, want *true", resp.Swaths.Sorted)
	}
	if string(resp.Journey.StepName) != "swaths:sort" {
		t.Errorf("StepName = %q", resp.Journey.StepName)
	}
}

func TestSortSwaths_BadJSON(t *testing.T) {
	t.Parallel()
	srv := NewServer(&mockF2CClient{})
	rec := doRequest(t, srv.SortSwaths, http.MethodPost, "/swaths:sort", "}}}")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "INVALID_JSON") {
		t.Errorf("body = %s", rec.Body.String())
	}
}
