all: test

PROCS = 2

help:                        ## Display this help message.
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' 'NF==2 {printf "  %-20s%s\n", $$1, $$2}'

docker:
	docker run -it -v $(PWD):/workshop -w /workshop ghcr.io/aleksi/golang-tip:dev.fuzz /bin/bash

init:                        ## Install tools.
	go generate -x -tags=tools ./tools
	make build

build:                       ## Build with race detector.
	go install -v -race ./...

test: build                  ## Test with race detector.
	go test -v -race ./...

fuzz-reverse:                ## Fuzz reverse function using dev.fuzz.
	go test -v -race -fuzz=FuzzReverse -fuzztime=50000x -parallel=$(PROCS)

gofuzz-reverse:              ## Fuzz reverse function using dvyukov/go-fuzz.
	./bin/go-fuzz-build -race
	./bin/go-fuzz -procs=$(PROCS)

fuzz-protocol:               ## Fuzz protocol using dev.fuzz.
	cd protocol && go test -v -race -fuzz=FuzzHandler -parallel=$(PROCS)

gofuzz-protocol:             ## Fuzz protocol using dvyukov/go-fuzz.
	cd protocol && ../bin/go-fuzz-build -race
	cd protocol && ../bin/go-fuzz -procs=$(PROCS)
