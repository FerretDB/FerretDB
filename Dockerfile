ARG VERSION
ARG COMMIT

FROM golang:1.18.2 AS build

WORKDIR /src
ADD . .
RUN CGO_ENABLED=0 go test -c -trimpath -o=bin/ferretdb -tags=testcover -coverpkg=./... ./cmd/ferretdb

FROM scratch

COPY --from=build /src/bin/ferretdb /ferretdb

EXPOSE 27017

ENTRYPOINT [ "/ferretdb" ]
CMD [ "-listen-addr=:27017", "-postgresql-url=postgres://username:password@hostname:5432/ferretdb" ]

# https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.description="A truly Open Source MongoDB alternative"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.revision="${COMMIT}"
LABEL org.opencontainers.image.source="https://github.com/FerretDB/FerretDB"
LABEL org.opencontainers.image.title="FerretDB"
LABEL org.opencontainers.image.url="https://ferretdb.io/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
LABEL org.opencontainers.image.version="${VERSION}"
