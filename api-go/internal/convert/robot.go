// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package convert

import (
	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
)

// RobotToProto converts the OpenAPI Robot envelope to a proto Robot.
//
// All optional pointer fields in api.Robot are flattened to their zero
// value on the proto side — proto3 has no concept of optional doubles
// without wrapper types, and the f2c shim treats 0 as "use default"
// for every field except width_m and cov_width_m (the two required
// fields per the OpenAPI schema).
func RobotToProto(in api.Robot) *f2cclient.Robot {
	out := &f2cclient.Robot{
		WidthM:    in.WidthM,
		CovWidthM: in.CovWidthM,
	}
	if in.Name != nil {
		out.Name = *in.Name
	}
	if in.MinTurningRadiusM != nil {
		out.MinTurningRadiusM = *in.MinTurningRadiusM
	}
	if in.MaxCurv != nil {
		out.MaxCurv = *in.MaxCurv
	}
	if in.MaxDiffCurv != nil {
		out.MaxDiffCurv = *in.MaxDiffCurv
	}
	if in.CruiseVelMps != nil {
		out.CruiseVelMps = *in.CruiseVelMps
	}
	if in.TurnVelMps != nil {
		out.TurnVelMps = *in.TurnVelMps
	}
	return out
}

// RobotFromProto reverses RobotToProto. Nil input yields a zero-valued
// api.Robot (WidthM=0, CovWidthM=0). Optional fields on the OpenAPI
// side are populated only when the proto side carries a non-zero
// value, preserving the "absent vs zero" distinction the OpenAPI
// schema exposes through pointer types.
func RobotFromProto(in *f2cclient.Robot) api.Robot {
	if in == nil {
		return api.Robot{}
	}
	out := api.Robot{
		WidthM:    in.WidthM,
		CovWidthM: in.CovWidthM,
	}
	if in.Name != "" {
		name := in.Name
		out.Name = &name
	}
	if in.MinTurningRadiusM != 0 {
		v := in.MinTurningRadiusM
		out.MinTurningRadiusM = &v
	}
	if in.MaxCurv != 0 {
		v := in.MaxCurv
		out.MaxCurv = &v
	}
	if in.MaxDiffCurv != 0 {
		v := in.MaxDiffCurv
		out.MaxDiffCurv = &v
	}
	if in.CruiseVelMps != 0 {
		v := in.CruiseVelMps
		out.CruiseVelMps = &v
	}
	if in.TurnVelMps != 0 {
		v := in.TurnVelMps
		out.TurnVelMps = &v
	}
	return out
}
