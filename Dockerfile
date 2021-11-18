FROM scratch

ADD bin/mangodb /mangodb

EXPOSE 27017

ENTRYPOINT [ "/mangodb" ]
CMD [ "-mode=diff-normal" ]

LABEL org.opencontainers.image.source=https://github.com/MangoDB-io/MangoDB
LABEL org.opencontainers.image.title=MangoDB
