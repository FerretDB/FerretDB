# for development releases (`ferret-dev` image)

# While we already know commit and version from commit.txt and version.txt inside image,
# it is not possible to use them in LABELs for the final image.
# We need to pass them as build arguments.
# Defining ARGs there makes them global.
ARG LABEL_VERSION
ARG LABEL_COMMIT

FROM ghcr.io/ferretdb/golang:1.20.1-1 AS development-build

ARG CACHEBUST=8
RUN echo "$CACHEBUST"

ARG LABEL_VERSION
ARG LABEL_COMMIT
RUN test -n "$LABEL_VERSION"
RUN test -n "$LABEL_COMMIT"

RUN env

WORKDIR /src

# see .dockerignore
COPY . .

ENV CGO_ENABLED=1
ENV GORACE=halt_on_error=1,history_size=2
ENV GOCOVERDIR=cover

# TODO
# That command could be run only once by using a separate stage and/or cache;
# see  https://medium.com/@tonistiigi/faster-multi-platform-builds-dockerfile-cross-compilation-guide-part-1-ec087c719eaf
RUN go mod download

RUN ls -al
RUN go test -v -c -o=bin/ferretdb -trimpath -tags=ferretdb_testcover,ferretdb_tigris,ferretdb_hana -race -coverpkg=./... ./cmd/ferretdb
RUN go version -m bin/ferretdb
RUN bin/ferretdb --version


FROM ghcr.io/ferretdb/golang:1.20.1-1

ARG LABEL_VERSION
ARG LABEL_COMMIT

COPY --from=development-build /src/bin/ferretdb /ferretdb

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
