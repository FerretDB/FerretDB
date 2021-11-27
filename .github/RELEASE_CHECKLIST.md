# Release Checklist

1. Check tests, linters.
2. Check issues and pull requests, update milestones and labels.
3. Update CHANGELOG.md.
4. Make a signed tag `vX.Y.Z` with the relevant section of the changelog (without leading `##`).
5. Push it!
6. Make [release](https://github.com/FerretDB/FerretDB/releases).
7. Refresh
   * `env GO111MODULE=on GOPROXY=https://proxy.golang.org go get -v github.com/FerretDB/FerretDB@<tag>`
   * https://pkg.go.dev/github.com/FerretDB/FerretDB
8. `make docker`, add tag, push image.
