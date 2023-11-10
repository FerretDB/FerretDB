# syntax=docker/dockerfile:1

# Dockerfile for development releases (`ferret-dev` image).
# It uses cross-compilation at first,
# the re-builds binaries with a race detector if possible.

# While we already know commit and version from commit.txt and version.txt inside image,
# it is not possible to use them in LABELs for the final image.
# We need to pass them as build arguments.
# Defining ARGs there makes them global.
ARG LABEL_VERSION
ARG LABEL_COMMIT


# prepare stage

FROM --platform=$BUILDPLATFORM ghcr.io/ferretdb/golang:1.21.4-1 AS development-prepare

# use a single directory for all Go caches to simpliy RUN --mount commands below
ENV GOPATH /cache/gopath
ENV GOCACHE /cache/gocache
ENV GOMODCACHE /cache/gomodcache

# remove ",direct"
ENV GOPROXY https://proxy.golang.org

# see .dockerignore
WORKDIR /src
COPY . .

RUN --mount=type=cache,target=/cache <<EOF
set -ex

go mod download
go mod verify

# copy cached stdlib builds from base image
cp -Rn /root/.cache/go-build/. /cache/gocache
EOF

ARG TARGETARCH TARGETVARIANT

RUN --mount=type=cache,target=/cache <<EOF
set -ex

# Leave GOAMD64 unset for implicit v1
# because v2+ is problematic for some virtualization platforms and older hardware.

flock --verbose /cache/ \
env GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    go build -v -o=/tmp/ferretdb/$TARGETARCH_$TARGETVARIANT -tags=ferretdb_debug -coverpkg=./... ./cmd/ferretdb
EOF


# build stage

FROM ghcr.io/ferretdb/golang:1.21.4-1 AS development-build

ARG LABEL_VERSION
ARG LABEL_COMMIT
RUN test -n "$LABEL_VERSION"
RUN test -n "$LABEL_COMMIT"

# use a single directory for all Go caches to simpliy RUN --mount commands below
ENV GOPATH /cache/gopath
ENV GOCACHE /cache/gocache
ENV GOMODCACHE /cache/gomodcache

# modules are already downloaded
ENV GOPROXY off

ENV CGO_ENABLED=0

ARG TARGETARCH TARGETVARIANT

# to add a dependency
COPY --from=development-prepare /tmp/ferretdb/$TARGETARCH_$TARGETVARIANT /tmp/ferretdb/

# see .dockerignore
WORKDIR /src
COPY . .

RUN --mount=type=cache,target=/cache <<EOF
set -ex

# copy cached stdlib builds from base image
flock --verbose /cache/ cp -Rn /root/.cache/go-build/. /cache/gocache

git status

# Disable race detector on arm64 due to https://github.com/golang/go/issues/29948
# (and that happens on GitHub-hosted Actions runners).
# Also disable it on arm/v6 and arm/v7 because it is not supported there.
RACE=false
if test "$TARGETARCH" = "amd64"
then
    RACE=true
fi

# Do not trim paths to make debugging with delve easier.

# Leave GOAMD64 unset for implicit v1
# because v2+ is problematic for some virtualization platforms and older hardware.

# Leave GOARM unset for autodetection.

# check that stdlib was cached
# go install -v -race=$RACE std

go build -v -o=bin/ferretdb -race=$RACE -tags=ferretdb_debug -coverpkg=./... ./cmd/ferretdb

go version -m bin/ferretdb
bin/ferretdb --version
EOF


# stage for binary only

FROM scratch AS development-binary

COPY --from=development-build /src/bin/ferretdb /ferretdb


# final stage

FROM golang:1.21.4 AS development

ENV GOCOVERDIR=/tmp/cover
ENV GORACE=halt_on_error=1,history_size=2

RUN mkdir /tmp/cover

COPY --from=development-build /src/bin/ferretdb /ferretdb

ENTRYPOINT [ "/ferretdb" ]

WORKDIR /
VOLUME /state
EXPOSE 27017 27018 8080

# don't forget to update documentation if you change defaults
ENV FERRETDB_LISTEN_ADDR=:27017
# ENV FERRETDB_LISTEN_TLS=:27018
ENV FERRETDB_DEBUG_ADDR=:8080
ENV FERRETDB_STATE_DIR=/state
ENV FERRETDB_SQLITE_URL=file:/state/

ARG LABEL_VERSION
ARG LABEL_COMMIT

# TODO https://github.com/FerretDB/FerretDB/issues/2212
LABEL org.opencontainers.image.description="A truly Open Source MongoDB alternative"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.revision="${LABEL_COMMIT}"
LABEL org.opencontainers.image.source="https://github.com/FerretDB/FerretDB"
LABEL org.opencontainers.image.title="FerretDB"
LABEL org.opencontainers.image.url="https://ferretdb.io/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
LABEL org.opencontainers.image.version="${LABEL_VERSION}"
