FROM node:16.17.0-alpine3.16

WORKDIR /tool
COPY . .

RUN npm ci

WORKDIR /workdir
ENTRYPOINT ["textlint"]
CMD ["-h"]
