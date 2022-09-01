FROM node:16.17.0-alpine3.16

WORKDIR /tool
COPY . .

RUN npm ci
RUN ln -vs /tool/node_modules/.bin/* /usr/local/bin

WORKDIR /workdir
ENTRYPOINT ["textlint"]
CMD ["-h"]
