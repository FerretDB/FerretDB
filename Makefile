BENCHTIME ?= 5s
FUZZTIME ?= 15s
FUZZCORPUS ?= ../fuzz-corpus

all: fmt test

help:                                  ## Display this help message
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'


init: gen-version                      ## Install development tools
	rm -fr bin
	go mod tidy
	cd tools && go mod tidy
	go mod verify
	cd tools && go generate -x

gen: bin/gofumpt                       ## Generate code
	go generate -x ./...
	$(MAKE) fmt


fmt: bin/gofumpt                       ## Format code
	bin/gofumpt -w .


bench-short:                           ## Benchmark for about 20 seconds (with default BENCHTIME)
	go test -list='Benchmark.*' ./...
	rm -f new.txt
	# Running four functions for $(BENCHTIME) each..."
	go test -bench=BenchmarkArray    -benchtime=$(BENCHTIME) ./internal/bson/  | tee -a new.txt
	go test -bench=BenchmarkDocument -benchtime=$(BENCHTIME) ./internal/bson/  | tee -a new.txt
	go test -bench=BenchmarkArray    -benchtime=$(BENCHTIME) ./internal/fjson/ | tee -a new.txt
	go test -bench=BenchmarkDocument -benchtime=$(BENCHTIME) ./internal/fjson/ | tee -a new.txt
	bin/benchstat old.txt new.txt



docker-build: build-testcover
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go test -c -o=bin/ferretdb-arm64 -trimpath -tags=testcover -coverpkg=./... ./cmd/ferretdb
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -c -o=bin/ferretdb-amd64 -trimpath -tags=testcover -coverpkg=./... ./cmd/ferretdb

docker-local: docker-build
	docker buildx build --builder=ferretdb \
		--build-arg VERSION=$(shell cat internal/util/version/version.txt) \
		--build-arg COMMIT=$(shell cat internal/util/version/commit.txt) \
		--tag=ferretdb-local \
		--load .

docker-push: docker-build
	test $(DOCKER_IMAGE)
	docker buildx build --builder=ferretdb --platform=linux/arm64,linux/amd64 \
		--build-arg VERSION=$(shell cat internal/util/version/version.txt) \
		--build-arg COMMIT=$(shell cat internal/util/version/commit.txt) \
		--tag=$(DOCKER_IMAGE) \
		--push .

bin/golangci-lint:
	$(MAKE) init

bin/go-sumtype:
	$(MAKE) init

bin/gofumports:
	$(MAKE) init
