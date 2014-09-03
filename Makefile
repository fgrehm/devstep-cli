.PHONY: test build

default: test

build: $(wildcard **/*.go)
	@mkdir -p build
	go build -v -o build/devstep

test:
	go test ./...

coverage:
	@mkdir -p build
	gocov test github.com/fgrehm/devstep-cli/devstep | gocov-html > build/coverage.html
