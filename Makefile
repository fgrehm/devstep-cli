.PHONY: test build coverage ci deps

default: test

ci: deps test vet build

deps:
	go get ./...

build: $(wildcard **/*.go)
	@echo "Building CLI..."
	@mkdir -p build
	go build -o build/devstep
	@echo "DONE"

test:
	go test ./...

vet:
	go tool vet -all .

coverage:
	@mkdir -p build
	gocov test github.com/fgrehm/devstep-cli/devstep | gocov-html > build/coverage.html
