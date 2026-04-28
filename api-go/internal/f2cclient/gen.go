// Package f2cclient hosts the protoc-generated f2c.v1 gRPC client.
//
// The proto file's go_package is overridden via -M flags to land the
// generated code inside this package (api-go/internal/f2cclient) instead
// of at the path declared in proto/f2c.proto. Rationale: Phase 11 keeps
// the generated client internal to the api-go module — there is no
// external Go consumer of f2c.v1 yet.
//
// Re-run `go generate ./...` from the api-go/ module root after editing
// ../../../proto/f2c.proto to refresh f2c.pb.go and f2c_grpc.pb.go.
package f2cclient

//go:generate protoc --proto_path=../../../proto --go_out=. --go_opt=paths=source_relative --go_opt=Mf2c.proto=github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient;f2cclient --go-grpc_out=. --go-grpc_opt=paths=source_relative --go-grpc_opt=Mf2c.proto=github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient;f2cclient f2c.proto
