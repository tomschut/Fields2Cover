//go:build tools
// +build tools

// Package tools pins build-time tool dependencies so `go mod tidy` does
// not drop them. They are referenced via blank imports here and invoked
// via `go run` from go:generate directives elsewhere in the module.
//
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package tools

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
