.PHONY: test build coverage ci deps

default: test

ci: deps test build

deps:
	go get ./...

build: $(wildcard **/*.go)
	@mkdir -p build
	go build -o build/devstep

test:
	go test ./...

coverage:
	@mkdir -p build
	gocov test github.com/fgrehm/devstep-cli/devstep | gocov-html > build/coverage.html
