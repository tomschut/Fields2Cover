// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/codes"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
	"github.com/Fields2Cover/fields2cover/api-go/internal/pipeline"
)

func TestGrpcCodeToHTTP(t *testing.T) {
	t.Parallel()
	cases := []struct {
		code codes.Code
		want int
	}{
		{codes.OK, http.StatusOK},
		{codes.InvalidArgument, http.StatusBadRequest},
		{codes.OutOfRange, http.StatusBadRequest},
		{codes.NotFound, http.StatusNotFound},
		{codes.AlreadyExists, http.StatusConflict},
		{codes.PermissionDenied, http.StatusForbidden},
		{codes.Unauthenticated, http.StatusUnauthorized},
		{codes.FailedPrecondition, http.StatusPreconditionFailed},
		{codes.ResourceExhausted, http.StatusTooManyRequests},
		{codes.Unimplemented, http.StatusNotImplemented},
		{codes.DeadlineExceeded, http.StatusGatewayTimeout},
		{codes.Unavailable, http.StatusServiceUnavailable},
		{codes.Internal, http.StatusInternalServerError},
		{codes.Unknown, http.StatusInternalServerError},
		{codes.DataLoss, http.StatusInternalServerError},
		{codes.Aborted, http.StatusInternalServerError},
		{codes.Canceled, http.StatusInternalServerError},
	}
	for _, tc := range cases {
		if got := grpcCodeToHTTP(tc.code); got != tc.want {
			t.Errorf("grpcCodeToHTTP(%v) = %d, want %d", tc.code, got, tc.want)
		}
	}
}

func TestWriteGrpcError_NonGrpcError(t *testing.T) {
	t.Parallel()
	srv := NewServer(&mockF2CClient{})
	rec := httptest.NewRecorder()
	srv.writeGrpcError(rec, errors.New("plain error"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestSortAlgorithmToProto(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   api.SortSwathsRequestAlgorithm
		want f2cclient.SortAlgorithm
	}{
		{api.BOUSTROPHEDON, f2cclient.SortAlgorithm_SORT_ALGORITHM_BOUSTROPHEDON},
		{api.SNAKE, f2cclient.SortAlgorithm_SORT_ALGORITHM_SNAKE},
		{api.SPIRAL, f2cclient.SortAlgorithm_SORT_ALGORITHM_SPIRAL},
		{api.SortSwathsRequestAlgorithm("BOGUS"), f2cclient.SortAlgorithm_SORT_ALGORITHM_UNSPECIFIED},
	}
	for _, tc := range cases {
		if got := pipeline.SortAlgorithmToProto(tc.in); got != tc.want {
			t.Errorf("pipeline.SortAlgorithmToProto(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestTurningAlgorithmToProto(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   api.PlanPathRequestTurningAlgorithm
		want f2cclient.TurningAlgorithm
	}{
		{api.PlanPathRequestTurningAlgorithmDUBINS, f2cclient.TurningAlgorithm_TURNING_ALGORITHM_DUBINS},
		{api.PlanPathRequestTurningAlgorithmDUBINSCC, f2cclient.TurningAlgorithm_TURNING_ALGORITHM_DUBINS_CC},
		{api.PlanPathRequestTurningAlgorithmREEDSSHEPP, f2cclient.TurningAlgorithm_TURNING_ALGORITHM_REEDS_SHEPP},
		{api.PlanPathRequestTurningAlgorithmREEDSSHEPPHC, f2cclient.TurningAlgorithm_TURNING_ALGORITHM_REEDS_SHEPP_HC},
		{api.PlanPathRequestTurningAlgorithm("BOGUS"), f2cclient.TurningAlgorithm_TURNING_ALGORITHM_UNSPECIFIED},
	}
	for _, tc := range cases {
		if got := pipeline.TurningAlgorithmToProto(tc.in); got != tc.want {
			t.Errorf("pipeline.TurningAlgorithmToProto(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
