// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package convert

import (
	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

// directionToProto maps the OpenAPI FORWARD/BACKWARD enum to the proto
// PathDirection enum. Unknown values fall through to UNSPECIFIED.
func directionToProto(in api.PathStateDirection) f2cclient.PathDirection {
	switch in {
	case api.FORWARD:
		return f2cclient.PathDirection_PATH_DIRECTION_FORWARD
	case api.BACKWARD:
		return f2cclient.PathDirection_PATH_DIRECTION_BACKWARD
	default:
		return f2cclient.PathDirection_PATH_DIRECTION_UNSPECIFIED
	}
}

// directionFromProto reverses directionToProto.
func directionFromProto(in f2cclient.PathDirection) api.PathStateDirection {
	switch in {
	case f2cclient.PathDirection_PATH_DIRECTION_FORWARD:
		return api.FORWARD
	case f2cclient.PathDirection_PATH_DIRECTION_BACKWARD:
		return api.BACKWARD
	default:
		return api.PathStateDirection("")
	}
}

// sectionTypeToProto maps the OpenAPI SWATH/TURN enum to the proto
// PathSectionType enum.
func sectionTypeToProto(in api.PathStateSectionType) f2cclient.PathSectionType {
	switch in {
	case api.SWATH:
		return f2cclient.PathSectionType_PATH_SECTION_TYPE_SWATH
	case api.TURN:
		return f2cclient.PathSectionType_PATH_SECTION_TYPE_TURN
	default:
		return f2cclient.PathSectionType_PATH_SECTION_TYPE_UNSPECIFIED
	}
}

// sectionTypeFromProto reverses sectionTypeToProto.
func sectionTypeFromProto(in f2cclient.PathSectionType) api.PathStateSectionType {
	switch in {
	case f2cclient.PathSectionType_PATH_SECTION_TYPE_SWATH:
		return api.SWATH
	case f2cclient.PathSectionType_PATH_SECTION_TYPE_TURN:
		return api.TURN
	default:
		return api.PathStateSectionType("")
	}
}

// pointSliceToProto converts the OpenAPI [x, y] inline slice to a
// proto Point. Defaults to a zero Point if the slice is malformed.
func pointSliceToProto(in []float64) *f2cclient.Point {
	if len(in) < 2 {
		return &f2cclient.Point{}
	}
	return &f2cclient.Point{X: in[0], Y: in[1]}
}

// pointSliceFromProto reverses pointSliceToProto. Nil input yields an
// empty slice rather than nil so JSON encodes it as `[]`.
func pointSliceFromProto(in *f2cclient.Point) []float64 {
	if in == nil {
		return []float64{}
	}
	return []float64{in.X, in.Y}
}

// PathStateToProto converts the OpenAPI PathState envelope to a proto
// PathState.
func PathStateToProto(in api.PathState) *f2cclient.PathState {
	return &f2cclient.PathState{
		Point:       pointSliceToProto(in.Point),
		AngleRad:    in.AngleRad,
		LengthM:     in.LengthM,
		Velocity:    in.Velocity,
		Direction:   directionToProto(in.Direction),
		SectionType: sectionTypeToProto(in.SectionType),
	}
}

// PathStateFromProto reverses PathStateToProto. Nil input yields a
// zero value.
func PathStateFromProto(in *f2cclient.PathState) api.PathState {
	if in == nil {
		return api.PathState{Point: []float64{}}
	}
	return api.PathState{
		Point:       pointSliceFromProto(in.Point),
		AngleRad:    in.AngleRad,
		LengthM:     in.LengthM,
		Velocity:    in.Velocity,
		Direction:   directionFromProto(in.Direction),
		SectionType: sectionTypeFromProto(in.SectionType),
	}
}

// PathToProto converts the OpenAPI Path envelope to a proto Path.
func PathToProto(in api.Path) *f2cclient.Path {
	out := &f2cclient.Path{
		Field:     FieldWithHeadlandsToProto(in.Field),
		LengthM:   in.LengthM,
		TaskTimeS: in.TaskTimeS,
	}
	if len(in.States) > 0 {
		out.States = make([]*f2cclient.PathState, 0, len(in.States))
		for i := range in.States {
			out.States = append(out.States, PathStateToProto(in.States[i]))
		}
	}
	return out
}

// PathFromProto reverses PathToProto. Nil input yields a zero value.
func PathFromProto(in *f2cclient.Path) api.Path {
	if in == nil {
		return api.Path{States: []api.PathState{}}
	}
	out := api.Path{
		Field:     FieldWithHeadlandsFromProto(in.Field),
		LengthM:   in.LengthM,
		TaskTimeS: in.TaskTimeS,
	}
	if len(in.States) > 0 {
		out.States = make([]api.PathState, 0, len(in.States))
		for _, s := range in.States {
			out.States = append(out.States, PathStateFromProto(s))
		}
	} else {
		out.States = []api.PathState{}
	}
	return out
}
