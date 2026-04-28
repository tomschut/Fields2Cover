.PHONY: generate test vet lint build clean

# Run all go:generate directives (oapi-codegen + protoc).
# Re-running this target on a clean tree must produce zero diff.
generate:
	go generate ./...
	go mod tidy

test:
	go test ./... -race -cover

vet:
	go vet ./...

build:
	go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api

clean:
	rm -rf bin/ dist/ coverage.out coverage.html
