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

WORKDIR /

COPY --from=build /src/bin/ferretdb /ferretdb

EXPOSE 27017

ENTRYPOINT [ "/ferretdb" ]
CMD [ "--listen-addr=:27017", "--postgresql-url=postgres://username:password@hostname:5432/ferretdb" ]

# https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.description="A truly Open Source MongoDB alternative"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.revision="${COMMIT}"
LABEL org.opencontainers.image.source="https://github.com/FerretDB/FerretDB"
LABEL org.opencontainers.image.title="FerretDB"
LABEL org.opencontainers.image.url="https://ferretdb.io/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
LABEL org.opencontainers.image.version="${VERSION}"
