FROM node:18.8.0-alpine3.16

WORKDIR /workdir/build/deps/textlint
COPY . .
RUN npm ci

ENTRYPOINT ["/workdir/build/deps/textlint/node_modules/.bin/textlint"]
CMD ["-h"]
