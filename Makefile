image=paskalmaksim/developer-proxy:dev

test:
	go mod tidy
	go test ./pkg/...
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -v

build:
	go run github.com/goreleaser/goreleaser@latest build --clean --skip=validate --snapshot
	cp ./dist/developer-proxy_linux_amd64_v1/developer-proxy ./developer-proxy
	docker build --pull --push -t $(image) .
