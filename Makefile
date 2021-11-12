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
	until [ "`docker inspect mangodb_postgres -f {{.State.Health.Status}}`" = "healthy" ]; do sleep 1; done
	until [ "`docker inspect mangodb_mongodb  -f {{.State.Health.Status}}`" = "healthy" ]; do sleep 1; done
	go run ./tools/envtool/main.go

env-pull:
	docker-compose pull --include-deps --quiet

env-down:                              ## Stop development environment
	docker-compose down --remove-orphans

init:                                  ## Install development tools
	go mod tidy
	cd tools && go mod tidy && go generate -tags=tools -x

gen: bin/gofumpt                       ## Generate code
	go generate -x ./...
	$(MAKE) fmt

gen-version:
	go generate -x ./internal/util/version

fmt: bin/gofumpt                       ## Format code
	bin/gofumpt -w .

test:                                  ## Run tests
	go test -race -coverprofile=cover.txt -coverpkg=./... -shuffle=on ./...

# That's not quite correct: https://github.com/golang/go/issues/15513
# But good enough for us.
fuzz-prepare: gen-version
	go test -count=0 ./...

fuzz-short:                            ## Fuzz for 1 minute
	go test -list='Fuzz.*' ./...
	go test -fuzz=FuzzArrayBinary -fuzztime=1m ./internal/bson/
	go test -fuzz=FuzzArrayJSON -fuzztime=1m ./internal/bson/
	go test -fuzz=FuzzDocumentBinary -fuzztime=1m ./internal/bson/
	go test -fuzz=FuzzDocumentJSON -fuzztime=1m ./internal/bson/
	go test -fuzz=FuzzMsg -fuzztime=1m ./internal/wire/
	go test -fuzz=FuzzQuery -fuzztime=1m ./internal/wire/
	go test -fuzz=FuzzReply -fuzztime=1m ./internal/wire/

bench-short:                           ## Benchmark for 5 seconds
	go test -list='Bench.*' ./...
	go test -bench=BenchmarkArray -benchtime=5s ./internal/bson/
	go test -bench=BenchmarkDocument -benchtime=5s ./internal/bson/

build-testcover: gen-version           ## Build bin/mangodb-testcover
	go test -c -o=bin/mangodb-testcover -trimpath -tags=testcover -race -coverpkg=./... ./cmd/mangodb

run: build-testcover                   ## Run MangoDB
	bin/mangodb-testcover -test.coverprofile=cover.txt -mode=diff

lint: bin/go-sumtype bin/golangci-lint ## Run linters
	bin/go-sumtype ./...
	bin/golangci-lint run --config=.golangci-required.yml
	bin/golangci-lint run --config=.golangci.yml

psql:                                  ## Run psql
	docker-compose exec postgres psql -U postgres -d mangodb

mongosh:                               ## Run mongosh
	docker-compose exec mongodb mongosh mongodb://host.docker.internal:27017/monila \
		--verbose --eval 'disableTelemetry()' --shell

mongo:                                 ## Run (legacy) mongo shell
	docker-compose exec mongodb mongo mongodb://host.docker.internal:27017/monila \
		--verbose

docker: build-testcover gen-version
	env GOOS=linux go test -c -o=bin/mangodb -trimpath -tags=testcover -coverpkg=./... ./cmd/mangodb
	docker build --tag=ghcr.io/mangodb-io/mangodb:latest .

bin/golangci-lint:
	$(MAKE) init

bin/go-sumtype:
	$(MAKE) init

bin/gofumports:
	$(MAKE) init
