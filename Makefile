.PHONY: test build coverage ci deps

default: test

ci: deps test vet build

deps:
	go get ./...
	go get code.google.com/p/go.tools/cmd/vet

build: $(wildcard **/*.go)
	@echo "Building CLI..."
	@mkdir -p build
	gox -verbose -osarch="darwin/amd64 linux/amd64" -output="build/{{.OS}}_{{.Arch}}"
	@echo "DONE"

test:
	go test ./...

vet:
	go tool vet -all .

coverage:
	@mkdir -p build
	gocov test github.com/fgrehm/devstep-cli/devstep | gocov-html > build/coverage.html

release: deps test vet build
	@./bin/release
