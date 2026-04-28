// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package server

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fields2Cover/fields2cover/api-go/internal/api"
	"github.com/Fields2Cover/fields2cover/api-go/internal/pipeline"
)

// grpcCodeToHTTP maps gRPC status codes to HTTP status codes per the
// Phase 11 CONTEXT.md error mapping table.
func grpcCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// writeError writes an OpenAPI Error envelope. It uses the generated
// api.Error type so the wire format stays in sync with the spec.
func (s *Server) writeError(w http.ResponseWriter, code int, errCode, msg string) {
	body := api.Error{
		Code:    errCode,
		Message: msg,
	}
	s.writeJSON(w, code, body)
}

// writeGrpcError pulls a gRPC status off the supplied error, maps it
// to HTTP, and writes the OpenAPI Error envelope.
func (s *Server) writeGrpcError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		s.writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	httpCode := grpcCodeToHTTP(st.Code())
	s.writeError(w, httpCode, st.Code().String(), st.Message())
}

// writeStepError routes a pipeline.Steps error to the right HTTP
// response. It distinguishes the two error kinds:
//
//   - *pipeline.EmptyResponseError  -> 500 EMPTY_RESPONSE
//   - anything else                 -> delegate to writeGrpcError
func (s *Server) writeStepError(w http.ResponseWriter, err error) {
	var empty *pipeline.EmptyResponseError
	if errors.As(err, &empty) {
		s.writeError(w, http.StatusInternalServerError, "EMPTY_RESPONSE",
			fmt.Sprintf("shim returned empty response from %s", empty.Op))
		return
	}
	s.writeGrpcError(w, err)
}
