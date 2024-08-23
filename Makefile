test:
	go mod tidy
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -v

build:
	go run github.com/goreleaser/goreleaser@latest build --clean --skip=validate --snapshot
