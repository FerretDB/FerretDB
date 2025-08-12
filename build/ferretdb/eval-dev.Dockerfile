# syntax=docker/dockerfile:1

# evaluation image with development release (`ferretdb-eval-dev` image)

# While we already know commit and version from commit.txt and version.txt inside image,
# it is not possible to use them in LABELs for the final image.
# We need to pass them as build arguments.
# Defining ARGs there makes them global.
ARG LABEL_VERSION
ARG LABEL_COMMIT


# prepare stage

FROM --platform=$BUILDPLATFORM golang:1.24.5 AS eval-dev-prepare

# use a single directory for all Go caches to simplify RUN --mount commands below
ENV GOPATH=/cache/gopath
ENV GOCACHE=/cache/gocache
ENV GOMODCACHE=/cache/gomodcache

# remove ",direct"
ENV GOPROXY=https://proxy.golang.org

COPY go.mod go.sum /src/

WORKDIR /src

RUN --mount=type=cache,target=/cache <<EOF
set -ex

go mod download
go mod verify
EOF


# build stage

FROM golang:1.24.5 AS eval-dev-build

ARG TARGETARCH

ARG LABEL_VERSION
ARG LABEL_COMMIT
RUN test -n "$LABEL_VERSION"
RUN test -n "$LABEL_COMMIT"

# use the same directories for Go caches as above
ENV GOPATH=/cache/gopath
ENV GOCACHE=/cache/gocache
ENV GOMODCACHE=/cache/gomodcache

# modules are already downloaded
ENV GOPROXY=off

# see .dockerignore
WORKDIR /src
COPY . .

# to add a dependency
COPY --from=eval-dev-prepare /src/go.mod /src/go.sum /src/

RUN --mount=type=cache,target=/cache <<EOF
set -ex

git status

# Do not raise without providing separate builds with those values
# because higher versions are problematic for some virtualization platforms and older hardware.
export GOAMD64=v1
export GOARM64=v8.0

export CGO_ENABLED=1

# Disable race detector on arm64 due to https://github.com/golang/go/issues/29948
# (and that happens on GitHub-hosted Actions runners).
RACE=false
if test "$TARGETARCH" = "amd64"
then
    RACE=true
fi

go env

# Do not trim paths to make debugging with delve easier.

# check if stdlib was cached
go install -v -race=$RACE std

go build -v -o=bin/ferretdb -race=$RACE -tags=ferretdb_dev -coverpkg=./... ./cmd/ferretdb

go version -m bin/ferretdb
bin/ferretdb --version
EOF


# final stage

# Use development image and full tag close to the release.
FROM ghcr.io/ferretdb/postgres-documentdb-dev:17-0.106.0-ferretdb-2.5.0 AS eval-dev

# Use moving development image during development.
# FROM ghcr.io/ferretdb/postgres-documentdb-dev:17-ferretdb AS eval-dev

RUN --mount=type=cache,sharing=locked,target=/var/cache/apt <<EOF
apt install -y curl supervisor
curl -L https://pgp.mongodb.com/server-7.0.asc | apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/debian bookworm/mongodb-org/7.0 main" | tee /etc/apt/sources.list.d/mongodb-org-7.0.list
apt update
apt install -y mongodb-mongosh

mkdir /tmp/cover
EOF

COPY --from=eval-dev-build /src/bin/ferretdb /usr/local/bin/ferretdb

COPY --from=eval-dev-build /src/build/ferretdb/evaluation/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY --from=eval-dev-build /src/build/ferretdb/evaluation/entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]

HEALTHCHECK --interval=1m --timeout=5s --retries=1 --start-period=30s --start-interval=5s \
  CMD ["/usr/local/bin/ferretdb", "ping"]

VOLUME /state
EXPOSE 27017 27018 8088

ENV GOCOVERDIR=/tmp/cover
ENV GORACE=halt_on_error=1,history_size=2

# don't forget to update documentation if you change defaults
ENV FERRETDB_LISTEN_ADDR=:27017
# ENV FERRETDB_LISTEN_TLS=:27018
ENV FERRETDB_DEBUG_ADDR=:8088
ENV FERRETDB_STATE_DIR=/state

ARG LABEL_VERSION
ARG LABEL_COMMIT

# TODO https://github.com/FerretDB/FerretDB/issues/2212
LABEL org.opencontainers.image.description="A truly Open Source MongoDB alternative (evaluation development image)"
LABEL org.opencontainers.image.revision="${LABEL_COMMIT}"
LABEL org.opencontainers.image.source="https://github.com/FerretDB/FerretDB"
LABEL org.opencontainers.image.title="FerretDB (evaluation development image)"
LABEL org.opencontainers.image.url="https://www.ferretdb.com/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
LABEL org.opencontainers.image.version="${LABEL_VERSION}"
