FROM ghcr.io/igorshubovych/markdownlint-cli:v0.32.1
ENTRYPOINT [""]

RUN npm install --location=global textlint textlint-rule-one-sentence-per-line
