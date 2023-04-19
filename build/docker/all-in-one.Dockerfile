# syntax=docker/dockerfile:1

# for all-in-one releases (`all-in-one` image)

# While we already know commit and version from commit.txt and version.txt inside image,
# it is not possible to use them in LABELs for the final image.
# We need to pass them as build arguments.
# Defining ARGs there makes them global.
ARG LABEL_VERSION
ARG LABEL_COMMIT


# build stage

FROM golang:1.20.3 AS build

ARG LABEL_VERSION
ARG LABEL_COMMIT
RUN test -n "$LABEL_VERSION"
RUN test -n "$LABEL_COMMIT"

ARG TARGETARCH

# see .dockerignore
WORKDIR /src
COPY . .

# use a single directory for all Go caches to simpliy RUN --mount commands below
ENV GOPATH /gocaches/gopath
ENV GOCACHE /gocaches/gocache-${TARGETARCH}
ENV GOMODCACHE /gocaches/gomodcache

# to make caching easier
ENV GOFLAGS -modcacherw

# remove ",direct"
ENV GOPROXY https://proxy.golang.org

ENV CGO_ENABLED=1
ENV GOCOVERDIR=cover
ENV GORACE=halt_on_error=1,history_size=2
ENV GOARM=7

# do not raise it without providing a v1 build because v2+ is problematic for some virtualization platforms
ENV GOAMD64=v1

RUN --mount=type=bind,source=./tmp/docker/gocaches,target=/gocaches-host \
    --mount=type=cache,target=/gocaches \
<<EOF

set -ex

cp -R /gocaches-host/* /gocaches

go mod download
go mod verify

# Disable race detector on arm64 due to https://github.com/golang/go/issues/29948
# (and that happens on GitHub-hosted Actions runners).
RACE=false
if test "$TARGETARCH" = "amd64"
then
    RACE=true
fi

# build stdlib separately to check if it was cached
go install -v -race=$RACE std

# Do not use -trimpath to make debugging with delve easier.
go build -v                 -o=bin/ferretdb -race=$RACE -tags=ferretdb_testcover,ferretdb_tigris ./cmd/ferretdb
go test  -c -coverpkg=./... -o=bin/ferretdb -race=$RACE -tags=ferretdb_testcover,ferretdb_tigris ./cmd/ferretdb

go version -m bin/ferretdb
bin/ferretdb --version

EOF


# export cache stage
# use busybox and tar to export less files faster

FROM busybox AS export-cache

RUN --mount=type=cache,target=/gocaches tar cf /gocaches.tar gocaches


# final stage

FROM postgres:15.2 AS all-in-one

ARG LABEL_VERSION
ARG LABEL_COMMIT

COPY --from=build /src/bin/ferretdb /ferretdb

# all-in-one hacks start there

COPY --from=build /src/build/docker/all-in-one/ferretdb.sh /etc/service/ferretdb/run
COPY --from=build /src/build/docker/all-in-one/postgresql.sh /etc/service/postgresql/run
COPY --from=build /src/build/docker/all-in-one/entrypoint.sh /entrypoint.sh

RUN <<EOF
set -ex

apt update
apt install -y curl runit
curl -L https://www.mongodb.org/static/pgp/server-6.0.asc | apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/debian bullseye/mongodb-org/6.0 main" | tee /etc/apt/sources.list.d/mongodb-org-6.0.list
apt update
apt install -y mongodb-mongosh
EOF

ENV FERRETDB_POSTGRESQL_URL=postgres://username:password@127.0.0.1:5432/ferretdb
ENV POSTGRES_USER=username
ENV POSTGRES_PASSWORD=password
ENV POSTGRES_DB=ferretdb

STOPSIGNAL SIGHUP

WORKDIR /
ENTRYPOINT [ "/entrypoint.sh" ]
EXPOSE 27017

# all-in-one hacks stop there

# don't forget to update documentation if you change defaults
ENV FERRETDB_LISTEN_ADDR=:27017
# ENV FERRETDB_LISTEN_TLS=:27018
ENV FERRETDB_DEBUG_ADDR=:8080
ENV FERRETDB_STATE_DIR=/state

# TODO https://github.com/FerretDB/FerretDB/issues/2212
LABEL org.opencontainers.image.description="A truly Open Source MongoDB alternative"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.revision="${LABEL_COMMIT}"
LABEL org.opencontainers.image.source="https://github.com/FerretDB/FerretDB"
LABEL org.opencontainers.image.title="FerretDB"
LABEL org.opencontainers.image.url="https://ferretdb.io/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
LABEL org.opencontainers.image.version="${LABEL_VERSION}"
