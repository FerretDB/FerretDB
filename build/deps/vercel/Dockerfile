FROM node:16.17.0-alpine3.16

WORKDIR /workdir/build/deps/vercel
COPY . .
RUN npm ci

WORKDIR /workdir
ENTRYPOINT ["/workdir/build/deps/vercel/node_modules/.bin/vercel"]
CMD ["-h"]
