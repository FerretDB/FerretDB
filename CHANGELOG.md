# Changelog

<!-- markdownlint-disable MD024 MD034 -->

## [v2.0.0-rc.2](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0-rc.2) (2025-02-25)

This version works best with
[DocumentDB v0.102.0-ferretdb-2.0.0-rc.2](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.0.0-rc.2).

### Breaking changes

The `:latest` Docker image tag now points to this release.
It is highly recommended always to specify the full version (like `:2.0.0-rc.1`).

### What's Changed

#### Docker images

Docker images are again available on [Docker Hub](https://hub.docker.com/r/ferretdb/ferretdb)
and [quay.io](https://quay.io/repository/ferretdb/ferretdb)
(although we recommend using [GitHub Packages](https://github.com/ferretdb/FerretDB/pkgs/container/ferretdb)).

#### Embeddable Go package

Once again, FerretDB can be used as a [Go library](https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/ferretdb).

#### `.deb` packages for DocumentDB

`.deb` packages for DocumentDB are now available [there](https://github.com/FerretDB/documentdb/releases).

#### Indexes

Multiple issues with indexes were resolved, and support for TTL indexes was added.
After updating DocumentDB and FerretDB, rebuilding indexes using the `reIndex` command is recommended.

### New Features 🎉

- Update Docker tags by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4812
- Enable Docker Hub and quay.io by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4815
- Re-introduce embeddable package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4778
- Make embeddable package logging configurable by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4807
- Implement `reIndex` command by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4735
- Implement `dbStats` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/4804
- Add `mongo` slog format by @noisersup in https://github.com/FerretDB/FerretDB/pull/4716
- Enable building FerretDB as PostgreSQL background worker by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4802

### Enhancements 🛠

- Set `GOARM64` explicitly by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4760
- Tweak CLI for disabling interfaces by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4781

### Documentation 📄

- Add replication blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/4717
- Add explicit `platform: linux/amd64` to docs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4749
- Add missing space in command by @tlinhart in https://github.com/FerretDB/FerretDB/pull/4753
- Update config flags by @Fashander in https://github.com/FerretDB/FerretDB/pull/4740
- Update TLS docs by @Fashander in https://github.com/FerretDB/FerretDB/pull/4752
- Use `/` consistently in MongoDB URIs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4759
- Reformat request/response `.js` files by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4787
- Update Docker image tag in v1 docs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4796
- Convert `json5` to `js` responses in docs by @Fashander in https://github.com/FerretDB/FerretDB/pull/4800
- Tweak Docusaurus configuration by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4805
- Update release checklist and `.deb` installation guide by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4767
- Update blogs to use `js` instead of `json5` by @Fashander in https://github.com/FerretDB/FerretDB/pull/4801

### Other Changes 🤖

- Disable xfail CI configuration for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4698
- Skip private issue links in `checkcomments` by a flag by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4695
- Fix Dependabot config by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4708
- Fix error mapping generation by @noisersup in https://github.com/FerretDB/FerretDB/pull/4691
- Do not use Git LFS for Go files by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4713
- Refactor extract and convert in `genwrap` tool by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4703
- Introduce smoke test for Data API by @noisersup in https://github.com/FerretDB/FerretDB/pull/4712
- Do not use a global variable for logging by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4727
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4743
- Refactor and generate using `genwrap` tool by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4705
- Check DocumentDB issues in `checkcomments` and `checkdocs` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4714
- Use `wireclient` for readiness probe by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4718
- Remove "all backends" from the issue template by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4751
- Add/update TODO comments for some issues by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4723
- Remove `golang.org/x/exp/maps` package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4711
- Use Go 1.24.0 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4775
- Use directories instead of files for state and telemetry by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4784
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4788
- Make handler take over the pool by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4795
- Add test helper by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4799
- Bump Go deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4803
- Fix `envtool` for Go 1.24 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4809
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4811

### New Contributors

- @tlinhart made their first contribution in https://github.com/FerretDB/FerretDB/pull/4753

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/70?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v2.0.0-rc.1...v2.0.0-rc.2).

## [v2.0.0-rc.1](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0-rc.1) (2025-01-23)

The first release candidate of FerretDB v2, powered by [DocumentDB PostgreSQL extension](https://github.com/microsoft/documentdb)!

## Older Releases

See https://github.com/FerretDB/FerretDB/blob/main-v1/CHANGELOG.md.
