.PHONY: test build coverage ci ci-deps

default: test

ci: ci-deps coverage build

ci-deps:
	go get -v github.com/axw/gocov/gocov
	go get -v gopkg.in/matm/v1/gocov-html

build: $(wildcard **/*.go)
	@mkdir -p build
	go build -o build/devstep

test:
	go test ./...

coverage:
	@mkdir -p build
	gocov test github.com/fgrehm/devstep-cli/devstep | gocov-html > build/coverage.html
