ARG BASE_IMAGE=scratch
ARG COMMIT
ARG VERSION

FROM golang:1.18.3 AS build

ARG BUILD_ARGS=
ARG CGO_ENABLED=0

WORKDIR /src
ADD . .
RUN CGO_ENABLED=${CGO_ENABLED} go test -c -trimpath -o=bin/ferretdb -tags=testcover ${BUILD_ARGS} ./cmd/ferretdb

FROM ${BASE_IMAGE}

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
