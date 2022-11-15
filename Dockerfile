ARG VERSION
ARG COMMIT
ARG RACEFLAG

FROM golang:1.19.3 AS build

WORKDIR /src
ADD . .
ENV CGO_ENABLED=1
ENV GORACE=halt_on_error=1,history_size=2

# split into several commands for better logging on GitHub Actions
RUN go mod download
RUN go build -v -o=bin/ferretdb -trimpath -tags=ferretdb_testcover,ferretdb_tigris ${RACEFLAG}                 ./cmd/ferretdb
RUN go test  -c -o=bin/ferretdb -trimpath -tags=ferretdb_testcover,ferretdb_tigris ${RACEFLAG} -coverpkg=./... ./cmd/ferretdb

FROM golang:1.19.3

COPY --from=build /src/bin/ferretdb /ferretdb

WORKDIR /
ENTRYPOINT [ "/ferretdb" ]
EXPOSE 27017 8080
ENV FERRETDB_LISTEN_ADDR=:27017
ENV FERRETDB_DEBUG_ADDR=:8080
ENV FERRETDB_STATE_DIR=/state

# https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.description="A truly Open Source MongoDB alternative"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.revision="${COMMIT}"
LABEL org.opencontainers.image.source="https://github.com/FerretDB/FerretDB"
LABEL org.opencontainers.image.title="FerretDB"
LABEL org.opencontainers.image.url="https://ferretdb.io/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
LABEL org.opencontainers.image.version="${VERSION}"
