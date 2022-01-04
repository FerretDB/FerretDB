FROM scratch

ARG TARGETARCH

ADD bin/ferretdb-${TARGETARCH} /ferretdb

EXPOSE 27017

ENTRYPOINT [ "/ferretdb" ]
CMD [ "-mode=diff-normal" ]

LABEL org.opencontainers.image.source=https://github.com/FerretDB/FerretDB
LABEL org.opencontainers.image.title=FerretDB
LABEL org.opencontainers.image.description="A truly Open Source MongoDB alternative"
