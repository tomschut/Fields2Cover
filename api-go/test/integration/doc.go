// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Package integration hosts the end-to-end integration tests for the
// Fields2Cover v2.0 Go API + C++ gRPC shim. Every file in this package
// is gated by //go:build integration so `go test ./...` in the default
// mode skips it — keeping the unit-test loop fast. CI runs a dedicated
// job with -tags=integration (see Plan 12-04).
//
// The fixture (fixture.go) starts f2c-grpc-server as a subprocess on a
// temp Unix socket, dials it from Go, and serves the HTTP API via
// httptest.Server on a kernel-assigned port. Tests share a single shim
// subprocess across the whole package (via TestMain) to avoid paying
// the process-launch cost per test.
//
// See Plan 12-03 for the spec.

//go:build integration

package integration
