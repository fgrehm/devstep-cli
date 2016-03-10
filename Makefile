.PHONY: test build coverage ci deps

default: test build

ci: deps test vet build-ci

deps:
	go get ./...
	go get golang.org/x/tools/cmd/vet

build: $(wildcard **/*.go)
	@echo "Building CLI..."
	@mkdir -p build
	go build -o build/linux_amd64
	@echo "DONE"

build-ci: $(wildcard **/*.go)
	@mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o build/linux_amd64 .
	GOOS=darwin GOARCH=amd64 go build -o build/darwin_amd64 .
	@echo 'DONE'

test:
	go test ./...

vet:
	go tool vet -all .

coverage:
	@mkdir -p build
	gocov test github.com/fgrehm/devstep-cli/devstep | gocov-html > build/coverage.html

watchf:
	go get github.com/parkghost/watchf/...
	watchf
