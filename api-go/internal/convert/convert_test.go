// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package convert

import (
	"reflect"
	"testing"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
)

// helper: build a populated GeoJSONFeatureCollection that round-trips
// through json.Marshal/json.Unmarshal cleanly.
func sampleFeatureCollection() api.GeoJSONFeatureCollection {
	props := map[string]interface{}{"name": "test"}
	return api.GeoJSONFeatureCollection{
		Type: api.FeatureCollection,
		Features: []struct {
			Geometry   map[string]interface{}                   `json:"geometry"`
			Properties *map[string]interface{}                  `json:"properties"`
			Type       api.GeoJSONFeatureCollectionFeaturesType `json:"type"`
		}{
			{
				Geometry: map[string]interface{}{
					"type":        "Polygon",
					"coordinates": []interface{}{},
				},
				Properties: &props,
				Type:       api.Feature,
			},
		},
	}
}

func TestRoundTrip_CRS(t *testing.T) {
	t.Parallel()
	zone := "31N"
	original := api.CRS{Epsg: 32631, UtmZone: &zone}
	got := CRSFromProto(CRSToProto(original))
	if got.Epsg != 32631 {
		t.Errorf("Epsg: got %d, want 32631", got.Epsg)
	}
	if got.UtmZone == nil || *got.UtmZone != "31N" {
		t.Errorf("UtmZone: got %v, want 31N", got.UtmZone)
	}

	// Minimal: zero epsg, no zone
	min := api.CRS{Epsg: 0}
	gotMin := CRSFromProto(CRSToProto(min))
	if gotMin.Epsg != 0 {
		t.Errorf("min Epsg: got %d, want 0", gotMin.Epsg)
	}
	if gotMin.UtmZone != nil {
		t.Errorf("min UtmZone: got %v, want nil", gotMin.UtmZone)
	}
}

func TestRoundTrip_Robot_Full(t *testing.T) {
	t.Parallel()
	name := "test-robot"
	mtr := 1.5
	maxCurv := 0.66
	maxDiff := 0.12
	cruise := 5.0
	turn := 2.5
	original := api.Robot{
		Name:              &name,
		WidthM:            2.0,
		CovWidthM:         1.8,
		MinTurningRadiusM: &mtr,
		MaxCurv:           &maxCurv,
		MaxDiffCurv:       &maxDiff,
		CruiseVelMps:      &cruise,
		TurnVelMps:        &turn,
	}
	got := RobotFromProto(RobotToProto(original))
	if got.WidthM != 2.0 {
		t.Errorf("WidthM: got %v, want 2.0", got.WidthM)
	}
	if got.CovWidthM != 1.8 {
		t.Errorf("CovWidthM: got %v, want 1.8", got.CovWidthM)
	}
	if got.Name == nil || *got.Name != name {
		t.Errorf("Name: got %v, want %v", got.Name, name)
	}
	if got.MinTurningRadiusM == nil || *got.MinTurningRadiusM != mtr {
		t.Errorf("MinTurningRadiusM: got %v, want %v", got.MinTurningRadiusM, mtr)
	}
	if got.MaxCurv == nil || *got.MaxCurv != maxCurv {
		t.Errorf("MaxCurv: got %v, want %v", got.MaxCurv, maxCurv)
	}
	if got.MaxDiffCurv == nil || *got.MaxDiffCurv != maxDiff {
		t.Errorf("MaxDiffCurv: got %v, want %v", got.MaxDiffCurv, maxDiff)
	}
	if got.CruiseVelMps == nil || *got.CruiseVelMps != cruise {
		t.Errorf("CruiseVelMps: got %v, want %v", got.CruiseVelMps, cruise)
	}
	if got.TurnVelMps == nil || *got.TurnVelMps != turn {
		t.Errorf("TurnVelMps: got %v, want %v", got.TurnVelMps, turn)
	}
}

func TestRoundTrip_Robot_Minimal(t *testing.T) {
	t.Parallel()
	original := api.Robot{
		WidthM:    2.0,
		CovWidthM: 1.8,
	}
	got := RobotFromProto(RobotToProto(original))
	if got.WidthM != 2.0 || got.CovWidthM != 1.8 {
		t.Errorf("widths: got %v/%v, want 2.0/1.8", got.WidthM, got.CovWidthM)
	}
	// Optional fields must remain nil after round-trip (proto3 zero ↔ absent).
	if got.Name != nil {
		t.Errorf("Name: got %v, want nil", got.Name)
	}
	if got.MinTurningRadiusM != nil {
		t.Errorf("MinTurningRadiusM: got %v, want nil", got.MinTurningRadiusM)
	}
	if got.MaxCurv != nil || got.MaxDiffCurv != nil {
		t.Errorf("curv: got %v/%v, want nil/nil", got.MaxCurv, got.MaxDiffCurv)
	}
	if got.CruiseVelMps != nil || got.TurnVelMps != nil {
		t.Errorf("vel: got %v/%v, want nil/nil", got.CruiseVelMps, got.TurnVelMps)
	}
}

func TestRoundTrip_Field(t *testing.T) {
	t.Parallel()
	zone := "31N"
	ref := []float64{500000.0, 5000000.0}
	original := api.Field{
		Id:       "field-123",
		Crs:      api.CRS{Epsg: 32631, UtmZone: &zone},
		Geometry: sampleFeatureCollection(),
		RefPoint: &ref,
	}
	got := FieldFromProto(FieldToProto(original))
	if got.Id != "field-123" {
		t.Errorf("Id: got %q, want field-123", got.Id)
	}
	if got.Crs.Epsg != 32631 {
		t.Errorf("Crs.Epsg: got %d, want 32631", got.Crs.Epsg)
	}
	if got.Geometry.Type != api.FeatureCollection {
		t.Errorf("Geometry.Type: got %q, want FeatureCollection", got.Geometry.Type)
	}
	if len(got.Geometry.Features) != 1 {
		t.Errorf("Geometry.Features len: got %d, want 1", len(got.Geometry.Features))
	}
	if got.RefPoint == nil || len(*got.RefPoint) != 2 {
		t.Fatalf("RefPoint: got %v, want length 2", got.RefPoint)
	}
	if (*got.RefPoint)[0] != 500000.0 || (*got.RefPoint)[1] != 5000000.0 {
		t.Errorf("RefPoint coords: got %v, want [500000, 5000000]", *got.RefPoint)
	}
}

func TestRoundTrip_FieldWithHeadlands(t *testing.T) {
	t.Parallel()
	original := api.FieldWithHeadlands{
		Field: api.Field{
			Id:       "fwh-1",
			Crs:      api.CRS{Epsg: 32631},
			Geometry: sampleFeatureCollection(),
		},
		Headlands:      sampleFeatureCollection(),
		InnerField:     sampleFeatureCollection(),
		HeadlandWidthM: 6.5,
	}
	got := FieldWithHeadlandsFromProto(FieldWithHeadlandsToProto(original))
	if got.Field.Id != "fwh-1" {
		t.Errorf("Field.Id: got %q, want fwh-1", got.Field.Id)
	}
	if got.HeadlandWidthM != 6.5 {
		t.Errorf("HeadlandWidthM: got %v, want 6.5", got.HeadlandWidthM)
	}
	if got.Headlands.Type != api.FeatureCollection {
		t.Errorf("Headlands.Type: got %q, want FeatureCollection", got.Headlands.Type)
	}
	if got.InnerField.Type != api.FeatureCollection {
		t.Errorf("InnerField.Type: got %q, want FeatureCollection", got.InnerField.Type)
	}
}

func TestRoundTrip_Swaths(t *testing.T) {
	t.Parallel()
	sortedT := true
	original := api.Swaths{
		Field: api.FieldWithHeadlands{
			Field: api.Field{Id: "f1", Crs: api.CRS{Epsg: 32631}, Geometry: sampleFeatureCollection()},
		},
		Items: []api.Swath{
			{
				Id:     1,
				WidthM: 2.0,
				Type:   api.MAINLAND,
				Path: struct {
					Coordinates [][]float64       `json:"coordinates"`
					Type        api.SwathPathType `json:"type"`
				}{
					Coordinates: [][]float64{{0, 0}, {10, 0}},
					Type:        api.LineString,
				},
			},
			{
				Id:     2,
				WidthM: 2.0,
				Type:   api.HEADLAND,
				Path: struct {
					Coordinates [][]float64       `json:"coordinates"`
					Type        api.SwathPathType `json:"type"`
				}{
					Coordinates: [][]float64{{0, 5}, {10, 5}},
					Type:        api.LineString,
				},
			},
		},
		Sorted: &sortedT,
	}
	got := SwathsFromProto(SwathsToProto(original))
	if len(got.Items) != 2 {
		t.Fatalf("Items len: got %d, want 2", len(got.Items))
	}
	if got.Items[0].Id != 1 || got.Items[0].Type != api.MAINLAND {
		t.Errorf("Items[0]: got id=%d type=%q, want 1/MAINLAND", got.Items[0].Id, got.Items[0].Type)
	}
	if got.Items[1].Id != 2 || got.Items[1].Type != api.HEADLAND {
		t.Errorf("Items[1]: got id=%d type=%q, want 2/HEADLAND", got.Items[1].Id, got.Items[1].Type)
	}
	if got.Items[0].Path.Type != api.LineString {
		t.Errorf("Items[0].Path.Type: got %q, want LineString", got.Items[0].Path.Type)
	}
	if !reflect.DeepEqual(got.Items[0].Path.Coordinates, [][]float64{{0, 0}, {10, 0}}) {
		t.Errorf("Items[0].Path.Coordinates round-trip mismatch: got %v", got.Items[0].Path.Coordinates)
	}
	if got.Sorted == nil || !*got.Sorted {
		t.Errorf("Sorted: got %v, want true", got.Sorted)
	}
}

func TestRoundTrip_Route(t *testing.T) {
	t.Parallel()
	original := api.Route{
		Field: api.FieldWithHeadlands{
			Field: api.Field{Id: "f1", Crs: api.CRS{Epsg: 32631}, Geometry: sampleFeatureCollection()},
		},
		Swaths: api.Swaths{
			Field: api.FieldWithHeadlands{Field: api.Field{Id: "f1", Crs: api.CRS{Epsg: 32631}, Geometry: sampleFeatureCollection()}},
			Items: []api.Swath{},
		},
		Connections: []api.RouteConnection{
			{FromSwathId: 1, ToSwathId: 2, Points: [][]float64{{0, 0}, {1, 1}}},
			{FromSwathId: 2, ToSwathId: 3, Points: [][]float64{{1, 1}, {2, 2}}},
		},
	}
	got := RouteFromProto(RouteToProto(original))
	if len(got.Connections) != 2 {
		t.Fatalf("Connections len: got %d, want 2", len(got.Connections))
	}
	if got.Connections[0].FromSwathId != 1 || got.Connections[0].ToSwathId != 2 {
		t.Errorf("Connections[0]: got from=%d to=%d, want 1/2",
			got.Connections[0].FromSwathId, got.Connections[0].ToSwathId)
	}
	if !reflect.DeepEqual(got.Connections[0].Points, [][]float64{{0, 0}, {1, 1}}) {
		t.Errorf("Connections[0].Points round-trip mismatch: got %v", got.Connections[0].Points)
	}
	if got.Connections[1].FromSwathId != 2 || got.Connections[1].ToSwathId != 3 {
		t.Errorf("Connections[1]: got from=%d to=%d, want 2/3",
			got.Connections[1].FromSwathId, got.Connections[1].ToSwathId)
	}
	if got.Field.Field.Id != "f1" {
		t.Errorf("nested Field.Id: got %q, want f1", got.Field.Field.Id)
	}
}

func TestRoundTrip_Path(t *testing.T) {
	t.Parallel()
	original := api.Path{
		Field: api.FieldWithHeadlands{
			Field: api.Field{Id: "f1", Crs: api.CRS{Epsg: 32631}, Geometry: sampleFeatureCollection()},
		},
		LengthM:   123.456,
		TaskTimeS: 78.9,
		States: []api.PathState{
			{
				Point:       []float64{0, 0},
				AngleRad:    0,
				LengthM:     0,
				Velocity:    5.0,
				Direction:   api.FORWARD,
				SectionType: api.SWATH,
			},
			{
				Point:       []float64{10, 0},
				AngleRad:    1.57,
				LengthM:     10,
				Velocity:    2.5,
				Direction:   api.BACKWARD,
				SectionType: api.TURN,
			},
		},
	}
	got := PathFromProto(PathToProto(original))
	if got.LengthM != 123.456 {
		t.Errorf("LengthM: got %v, want 123.456", got.LengthM)
	}
	if got.TaskTimeS != 78.9 {
		t.Errorf("TaskTimeS: got %v, want 78.9", got.TaskTimeS)
	}
	if len(got.States) != 2 {
		t.Fatalf("States len: got %d, want 2", len(got.States))
	}
	if got.States[0].Direction != api.FORWARD || got.States[0].SectionType != api.SWATH {
		t.Errorf("States[0]: got dir=%q sec=%q, want FORWARD/SWATH",
			got.States[0].Direction, got.States[0].SectionType)
	}
	if got.States[1].Direction != api.BACKWARD || got.States[1].SectionType != api.TURN {
		t.Errorf("States[1]: got dir=%q sec=%q, want BACKWARD/TURN",
			got.States[1].Direction, got.States[1].SectionType)
	}
	if !reflect.DeepEqual(got.States[0].Point, []float64{0, 0}) {
		t.Errorf("States[0].Point: got %v, want [0,0]", got.States[0].Point)
	}
	if !reflect.DeepEqual(got.States[1].Point, []float64{10, 0}) {
		t.Errorf("States[1].Point: got %v, want [10,0]", got.States[1].Point)
	}
	if got.States[1].AngleRad != 1.57 || got.States[1].Velocity != 2.5 {
		t.Errorf("States[1] scalars: got angle=%v vel=%v, want 1.57/2.5",
			got.States[1].AngleRad, got.States[1].Velocity)
	}
}

func TestNilSafety(t *testing.T) {
	t.Parallel()
	// Each call must return a zero value without panicking.
	_ = CRSFromProto(nil)
	_ = RobotFromProto(nil)
	_ = FieldFromProto(nil)
	_ = FieldWithHeadlandsFromProto(nil)
	_ = SwathFromProto(nil)
	_ = SwathsFromProto(nil)
	_ = RouteConnectionFromProto(nil)
	_ = RouteFromProto(nil)
	_ = PathStateFromProto(nil)
	_ = PathFromProto(nil)
}
