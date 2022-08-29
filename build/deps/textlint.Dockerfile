FROM node:16.17.0-alpine3.16

# Not using package-lock.json nor package.json here because packages are installed globally
RUN npm install -g textlint@=12.2.1 textlint-rule-one-sentence-per-line@=2.0.0

WORKDIR /workdir
ENTRYPOINT ["textlint"]
CMD ["-h"]
