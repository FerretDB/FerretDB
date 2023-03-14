# syntax=docker/dockerfile:1

# for all-in-one releases (`all-in-one` image)

# While we already know commit and version from commit.txt and version.txt inside image,
# it is not possible to use them in LABELs for the final image.
# We need to pass them as build arguments.
# Defining ARGs there makes them global.
ARG LABEL_VERSION
ARG LABEL_COMMIT


# build stage

FROM ghcr.io/ferretdb/golang:1.20.2-4 AS all-in-one-build

ARG LABEL_VERSION
ARG LABEL_COMMIT
RUN test -n "$LABEL_VERSION"
RUN test -n "$LABEL_COMMIT"

ARG TARGETARCH

# see .dockerignore
WORKDIR /src
COPY . .

# use a single directory for all Go caches to simpliy RUN --mount commands below
ENV GOPATH /cache/gopath
ENV GOCACHE /cache/gocache
ENV GOMODCACHE /cache/gomodcache

# copy cached stdlib builds from the image
RUN --mount=type=cache,target=/cache \
    /usr/bin/rsync -a /root/.cache/go-build/ /cache/gocache

# remove ",direct"
ENV GOPROXY https://proxy.golang.org

ENV CGO_ENABLED=1
ENV GOCOVERDIR=cover
ENV GORACE=halt_on_error=1,history_size=2
ENV GOARM=7

# do not raise it without providing a v1 build because v2+ is problematic for some virtualization platforms
ENV GOAMD64=v1

# TODO https://github.com/FerretDB/FerretDB/issues/2170
# That command could be run only once by using a separate stage;
# see https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide/
RUN --mount=type=cache,target=/cache \
    go mod download

# Do not trim paths to make debugging with delve easier.
#
# Disable race detector on arm64 due to https://github.com/golang/go/issues/29948
# (and that happens on GitHub-hosted Actions runners).
RUN --mount=type=cache,target=/cache <<EOF
set -ex

RACE=false
if test "$TARGETARCH" = "amd64"
then
    RACE=true
fi

# check that stdlib was cached
go install -v -trimpath=false -race=$RACE std

go build -v                 -o=bin/ferretdb -trimpath=false -race=$RACE -tags=ferretdb_testcover,ferretdb_tigris,ferretdb_hana ./cmd/ferretdb
go test  -c -coverpkg=./... -o=bin/ferretdb -trimpath=false -race=$RACE -tags=ferretdb_testcover,ferretdb_tigris,ferretdb_hana ./cmd/ferretdb
EOF

RUN go version -m bin/ferretdb
RUN bin/ferretdb --version


# final stage

FROM postgres:15.2 AS all-in-one

ARG LABEL_VERSION
ARG LABEL_COMMIT

COPY --from=all-in-one-build /src/bin/ferretdb /ferretdb

RUN <<EOF
set -ex

apt update
apt install -y curl runit
curl -L https://www.mongodb.org/static/pgp/server-6.0.asc | apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/debian bullseye/mongodb-org/6.0 main" | tee /etc/apt/sources.list.d/mongodb-org-6.0.list
apt update
apt install -y mongodb-mongosh
EOF

WORKDIR /
ENTRYPOINT [ "/ferretdb" ]
EXPOSE 27017 27018 8080

# don't forget to update documentation if you change defaults
ENV FERRETDB_LISTEN_ADDR=:27017
# ENV FERRETDB_LISTEN_TLS=:27018
ENV FERRETDB_DEBUG_ADDR=:8080
ENV FERRETDB_STATE_DIR=/state

# https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.description="A truly Open Source MongoDB alternative"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.revision="${LABEL_COMMIT}"
LABEL org.opencontainers.image.source="https://github.com/FerretDB/FerretDB"
LABEL org.opencontainers.image.title="FerretDB"
LABEL org.opencontainers.image.url="https://ferretdb.io/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
LABEL org.opencontainers.image.version="${LABEL_VERSION}"
