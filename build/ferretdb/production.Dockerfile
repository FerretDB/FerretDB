# syntax=docker/dockerfile:1

# for production releases (`ferret` image)

# While we already know commit and version from commit.txt and version.txt inside image,
# it is not possible to use them in LABELs for the final image.
# We need to pass them as build arguments.
# Defining ARGs there makes them global.
ARG LABEL_VERSION
ARG LABEL_COMMIT


# prepare stage

FROM --platform=$BUILDPLATFORM golang:1.22.3 AS production-prepare

# use a single directory for all Go caches to simpliy RUN --mount commands below
ENV GOPATH /cache/gopath
ENV GOCACHE /cache/gocache
ENV GOMODCACHE /cache/gomodcache

# remove ",direct"
ENV GOPROXY https://proxy.golang.org

COPY go.mod go.sum /src/

WORKDIR /src

RUN --mount=type=cache,target=/cache <<EOF
set -ex

go mod download
go mod verify
EOF


# build stage

FROM golang:1.22.3 AS production-build

ARG TARGETARCH
ARG TARGETVARIANT

ARG LABEL_VERSION
ARG LABEL_COMMIT
RUN test -n "$LABEL_VERSION"
RUN test -n "$LABEL_COMMIT"

# use the same directories for Go caches as above
ENV GOPATH /cache/gopath
ENV GOCACHE /cache/gocache
ENV GOMODCACHE /cache/gomodcache

# modules are already downloaded
ENV GOPROXY off

# see .dockerignore
WORKDIR /src
COPY . .

# to add a dependency
COPY --from=production-prepare /src/go.mod /src/go.sum /src/

RUN --mount=type=cache,target=/cache <<EOF
set -ex

git status

# Do not raise it without providing a separate v1 build
# because v2+ is problematic for some virtualization platforms and older hardware.
export GOAMD64=v1

# Set GOARM explicitly due to https://github.com/docker-library/golang/issues/494.
export GOARM=${TARGETVARIANT#v}

export CGO_ENABLED=0

go env

# Do not trim paths to reuse build cache.

# check if stdlib was cached
go install -v std

go build -v -o=bin/ferretdb ./cmd/ferretdb

go version -m bin/ferretdb
bin/ferretdb --version

mkdir /state
EOF

# stage for binary only

FROM scratch AS production-binary

COPY --from=production-build /src/bin/ferretdb /ferretdb


# final stage

FROM scratch AS production

COPY build/ferretdb/passwd /etc/passwd
COPY build/ferretdb/group  /etc/group
USER ferretdb:ferretdb

COPY --from=production-build /src/bin/ferretdb /ferretdb
COPY --from=production-build --chown=ferretdb:ferretdb /state /state

ENTRYPOINT [ "/ferretdb" ]

HEALTHCHECK --interval=30s --timeout=30s --start-period=0s --start-interval=5s --retries=3 \
  CMD /ferretdb ping

WORKDIR /
VOLUME /state
EXPOSE 27017 27018 8088

# don't forget to update documentation if you change defaults
ENV FERRETDB_LISTEN_ADDR=:27017
# ENV FERRETDB_LISTEN_TLS=:27018
ENV FERRETDB_DEBUG_ADDR=:8088
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
LABEL org.opencontainers.image.url="https://www.ferretdb.com/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
LABEL org.opencontainers.image.version="${LABEL_VERSION}"
