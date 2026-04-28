// Package api hosts the oapi-codegen generated HTTP types + ServerInterface.
//
// Re-run `go generate ./...` from the api-go/ module root after editing
// ../../openapi.yaml to refresh api.gen.go.
package api

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../oapi-codegen.yaml ../../openapi.yaml
