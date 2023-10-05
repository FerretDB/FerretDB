# syntax=docker/dockerfile:1

# for all-in-one releases (`all-in-one` image)

# While we already know commit and version from commit.txt and version.txt inside image,
# it is not possible to use them in LABELs for the final image.
# We need to pass them as build arguments.
# Defining ARGs there makes them global.
ARG LABEL_VERSION
ARG LABEL_COMMIT


# build stage

FROM ghcr.io/ferretdb/golang:1.21.1-2 AS all-in-one-build

ARG TARGETARCH

ARG LABEL_VERSION
ARG LABEL_COMMIT
RUN test -n "$LABEL_VERSION"
RUN test -n "$LABEL_COMMIT"

# use a single directory for all Go caches to simpliy RUN --mount commands below
ENV GOPATH /cache/gopath
ENV GOCACHE /cache/gocache
ENV GOMODCACHE /cache/gomodcache

# remove ",direct"
ENV GOPROXY https://proxy.golang.org

# do not raise it from the default value of v1 without providing a separate v1 build
# because v2+ is problematic for some virtualization platforms and older hardware
# ENV GOAMD64=v1

# leave GOARM unset for autodetection

ENV CGO_ENABLED=1

# see .dockerignore
WORKDIR /src
COPY . .

RUN --mount=type=cache,target=/cache <<EOF
set -ex

# copy cached stdlib builds from base image
flock --verbose /cache/ cp -Rnv /root/.cache/go-build/. /cache/gocache

# TODO https://github.com/FerretDB/FerretDB/issues/2170
# That command could be run only once by using a separate stage;
# see https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide/
flock --verbose /cache/ go mod download

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

# check that stdlib was cached
# env GODEBUG=gocachehash=1 go install -v -race=$RACE std
go install -v -race=$RACE std

go build -v -o=bin/ferretdb -race=$RACE -tags=ferretdb_debug -coverpkg=./... ./cmd/ferretdb

go version -m bin/ferretdb
bin/ferretdb --version
EOF


# final stage

FROM postgres:16.0 AS all-in-one

COPY --from=all-in-one-build /src/bin/ferretdb /ferretdb

ENV GOCOVERDIR=/tmp/cover
ENV GORACE=halt_on_error=1,history_size=2

RUN mkdir /tmp/cover

# all-in-one hacks start there

COPY --from=all-in-one-build /src/build/docker/all-in-one/ferretdb.sh /etc/service/ferretdb/run
COPY --from=all-in-one-build /src/build/docker/all-in-one/postgresql.sh /etc/service/postgresql/run
COPY --from=all-in-one-build /src/build/docker/all-in-one/entrypoint.sh /entrypoint.sh

RUN <<EOF
set -ex

apt update
apt install -y curl runit sqlite3

curl -L https://pgp.mongodb.com/server-7.0.asc | apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/debian bookworm/mongodb-org/7.0 main" | tee /etc/apt/sources.list.d/mongodb-org-7.0.list
apt update
apt install -y mongodb-mongosh
EOF

ENV FERRETDB_POSTGRESQL_URL=postgres://username:password@127.0.0.1:5432/ferretdb
ENV POSTGRES_USER=username
ENV POSTGRES_PASSWORD=password
ENV POSTGRES_DB=ferretdb

STOPSIGNAL SIGHUP
ENTRYPOINT [ "/entrypoint.sh" ]

# all-in-one hacks stop there

WORKDIR /
EXPOSE 27017 27018 8080

# don't forget to update documentation if you change defaults
ENV FERRETDB_LISTEN_ADDR=:27017
# ENV FERRETDB_LISTEN_TLS=:27018
ENV FERRETDB_DEBUG_ADDR=:8080
ENV FERRETDB_STATE_DIR=/state

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
