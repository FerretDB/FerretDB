FROM node:lts-alpine

RUN npm install --location=global textlint textlint-rule-one-sentence-per-line

WORKDIR /workdir
ENTRYPOINT ["textlint"]
CMD ["-h"]
