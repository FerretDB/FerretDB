FROM node:lts-alpine3.16

RUN npm install --location=global textlint@12.2.1 textlint-rule-one-sentence-per-line

WORKDIR /workdir
ENTRYPOINT ["textlint"]
CMD ["-h"]
