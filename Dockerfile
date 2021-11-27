FROM scratch

ADD bin/ferretdb /ferretdb

EXPOSE 27017

ENTRYPOINT [ "/ferretdb" ]
CMD [ "-mode=diff-normal" ]

LABEL org.opencontainers.image.source=https://github.com/FerretDB/FerretDB
LABEL org.opencontainers.image.title=FerretDB
