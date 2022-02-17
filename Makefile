FUZZTIME ?= 20s
FUZZCORPUS ?= ../fuzz-corpus

all: fmt test

help:                                  ## Display this help message
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

env-up: env-up-detach env-setup        ## Start development environment
	docker-compose logs --follow

env-up-detach:
	docker-compose up --always-recreate-deps --force-recreate --remove-orphans --renew-anon-volumes --detach

env-setup: gen-version
	go run ./cmd/envtool/main.go

env-pull:
	docker-compose pull --include-deps --quiet

env-down:                              ## Stop development environment
	docker-compose down --remove-orphans --volumes

init: gen-version                      ## Install development tools
	rm -fr bin
	go mod tidy
	cd tools && go mod tidy
	go mod verify
	cd tools && go generate -tags=tools -x

gen: bin/gofumpt                       ## Generate code
	go generate -x ./...
	$(MAKE) fmt

gen-version:
	go generate -x ./internal/util/version

fmt: bin/gofumpt                       ## Format code
	bin/gofumpt -w .

test:                                  ## Run tests
	go test -race -shuffle=on -coverprofile=cover.txt -coverpkg=./... ./...
	go test -race -shuffle=on -bench=. -benchtime=1x ./...

# That's not quite correct: https://github.com/golang/go/issues/15513
# But good enough for us.
fuzz-init: gen-version
	go test -count=0 ./...

fuzz:                                  ## Fuzz for about 2 minutes (with default FUZZTIME)
	go test -list='Fuzz.*' ./...
	# Running seven functions for $(FUZZTIME) each..."
	go test -fuzz=FuzzArray -fuzztime=$(FUZZTIME) ./internal/bson/
	go test -fuzz=FuzzDocument -fuzztime=$(FUZZTIME) ./internal/bson/
	go test -fuzz=FuzzArray -fuzztime=$(FUZZTIME) ./internal/fjson/
	go test -fuzz=FuzzDocument -fuzztime=$(FUZZTIME) ./internal/fjson/
	go test -fuzz=FuzzMsg -fuzztime=$(FUZZTIME) ./internal/wire/
	go test -fuzz=FuzzQuery -fuzztime=$(FUZZTIME) ./internal/wire/
	go test -fuzz=FuzzReply -fuzztime=$(FUZZTIME) ./internal/wire/

fuzz-corpus:                           ## Sync generated fuzz corpus with FUZZCORPUS
	go run ./cmd/fuzztool/fuzztool.go -src=$(FUZZCORPUS) -dst=generated
	go run ./cmd/fuzztool/fuzztool.go -dst=$(FUZZCORPUS) -src=generated

bench-short:                           ## Benchmark for about 20 seconds
	go test -list='Benchmark.*' ./...
	rm -f new.txt
	go test -bench=BenchmarkArray    -benchtime=5s ./internal/bson/  | tee -a new.txt
	go test -bench=BenchmarkDocument -benchtime=5s ./internal/bson/  | tee -a new.txt
	go test -bench=BenchmarkArray    -benchtime=5s ./internal/fjson/ | tee -a new.txt
	go test -bench=BenchmarkDocument -benchtime=5s ./internal/fjson/ | tee -a new.txt
	bin/benchstat old.txt new.txt

build-testcover: gen-version           ## Build bin/ferretdb-testcover
	go test -c -o=bin/ferretdb-testcover -trimpath -tags=testcover -race -coverpkg=./... ./cmd/ferretdb

run: build-testcover                   ## Run FerretDB
	bin/ferretdb-testcover -test.coverprofile=cover.txt -mode=diff-normal -listen-addr=:27017

lint: bin/go-sumtype bin/golangci-lint ## Run linters
	bin/go-sumtype ./...
	bin/golangci-lint run --config=.golangci-required.yml
	bin/golangci-lint run --config=.golangci.yml
	bin/go-consistent -pedantic ./...

psql:                                  ## Run psql
	docker-compose exec postgres psql -U postgres -d ferretdb

mongosh:                               ## Run mongosh
	docker-compose exec mongodb mongosh mongodb://host.docker.internal:27017/monila?heartbeatFrequencyMS=300000 \
		--verbose --eval 'disableTelemetry()' --shell

mongo:                                 ## Run (legacy) mongo shell
	docker-compose exec mongodb mongo mongodb://host.docker.internal:27017/monila?heartbeatFrequencyMS=300000 \
		--verbose

docker-init:
	docker buildx create --driver=docker-container --name=ferretdb

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
