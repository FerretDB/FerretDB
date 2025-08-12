# Changelog

<!-- markdownlint-disable MD024 MD034 -->

## [v2.5.0](https://github.com/FerretDB/FerretDB/releases/tag/v2.5.0) (2025-08-12)

This version works best with
[DocumentDB v0.106.0-ferretdb-2.5.0](https://github.com/FerretDB/documentdb/releases/tag/v0.106.0-ferretdb-2.5.0).

### Breaking changes

#### Metric names

We changed Prometheus metric names and `HELP` texts to highlight the fact that they are
[not stable yet](https://docs.ferretdb.io/configuration/observability/#metrics).
We plan to promote most of them to stable in the next release.

### What's Changed

#### RPM packages

`.rpm` packages for Red Hat Enterprise Linux are
[now available](https://docs.ferretdb.io/installation/documentdb/rpm/)!

### Documentation 📄

- Enable full visibility for images on zoom by @Fashander in https://github.com/FerretDB/FerretDB/pull/5346
- Add Vault blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/5368
- Add WeKan blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/5381

### Other Changes 🤖

- Add YugabyteDB integration tests by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5274
- Run all YugabyteDB integration tests by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5283
- Refactor normal and error responses by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5326
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5353
- Use Codecov for flaky integration tests detection by @noisersup in https://github.com/FerretDB/FerretDB/pull/5361
- Make accessing YugabyteDB logs easier by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5362
- Add components in codecov by @noisersup in https://github.com/FerretDB/FerretDB/pull/5364
- Skip YugabyteDB setup for `arm64` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5372
- Tweak Ollama and MCPHost setup by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5375
- Simplify flags handling by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5378
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5380
- Extract wiring into a separate package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5383
- Small cleanup by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5388
- Encapsulate PostgreSQL pool in the Handler by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5389
- Bump DocumentDB version by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5392
- Test flaky tests detection by @noisersup in https://github.com/FerretDB/FerretDB/pull/5393
- Small tweaks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5394
- Make a resource.Untrack thread-safe by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5396
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5398
- Enable MongoDB test commands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5400
- Update Prometheus metrics by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5403

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/79?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v2.4.0...v2.5.0).

## [v2.4.0](https://github.com/FerretDB/FerretDB/releases/tag/v2.4.0) (2025-07-16)

This version works best with
[DocumentDB v0.105.0-ferretdb-2.4.0](https://github.com/FerretDB/documentdb/releases/tag/v0.105.0-ferretdb-2.4.0).

### New Features 🎉

- Allow `clusterAdmin` users to perform user management commands by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5314

### Enhancements 🛠

- Use debug-level log messages for continuations by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5266

### Documentation 📄

- Add blog post on setting up LibreChat with FerretDB by @Fashander in https://github.com/FerretDB/FerretDB/pull/4999
- Add blog post for Payload CMS by @Fashander in https://github.com/FerretDB/FerretDB/pull/5257
- Add blog post on using Compass GUI by @Fashander in https://github.com/FerretDB/FerretDB/pull/5263
- Fix social links by @Fashander in https://github.com/FerretDB/FerretDB/pull/5264
- Add PayloadCMS to compatible apps by @Fashander in https://github.com/FerretDB/FerretDB/pull/5270
- Add Compass to compatible apps by @Fashander in https://github.com/FerretDB/FerretDB/pull/5275
- Wrap long lines in a blog post by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5280
- Add NodeBB compatibility blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/5285
- Add Librechat to compatibility section by @Fashander in https://github.com/FerretDB/FerretDB/pull/5286
- Add NodeBB to compatible apps by @Fashander in https://github.com/FerretDB/FerretDB/pull/5293
- Cleanup NodeBB blog post by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5294
- Add Growthbook blogpost by @Fashander in https://github.com/FerretDB/FerretDB/pull/5304
- Add Heyform compatibility blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/5316
- Add GrowthBook to compatible apps by @Fashander in https://github.com/FerretDB/FerretDB/pull/5319
- Add DBeaver compatibility blogpost by @Fashander in https://github.com/FerretDB/FerretDB/pull/5321
- Add HyperDX blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/5334
- Add compatible applications by @Fashander in https://github.com/FerretDB/FerretDB/pull/5335
- Prepare v2.4.0 release by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5344

### Other Changes 🤖

- Add YugabyteDB to local setup by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5258
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5261
- Remove skipping of `bypassEmptyTsReplacement` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5267
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5268
- Bring back Codecov and Coveralls integration by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5276
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5282
- Add `mcphost` as tool by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5287
- Refactor requests handling by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5292
- Add Dependabot configuration for mcphost by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5296
- Extend middleware requests and responses by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5301
- Minor refactoring of Data API by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5302
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5307
- Refactor responses handling by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5310
- Add authentication test for client reconnecting by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5311
- Do not use deprecated method by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5312
- Update TODO URLs by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5317
- Remove `must.NotFail(wire[bson].NewXXX(` pattern by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5318
- Update TODO URL for `dropUsers` test by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5324
- Avoid extra marshaling by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5327
- Split command handling and handlers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5328
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5342

### New Contributors

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/77?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v2.3.1...v2.4.0).

## [v2.3.1](https://github.com/FerretDB/FerretDB/releases/tag/v2.3.1) (2025-06-12)

Hotfix release to restore compatibility with various tools such as `mongoimport` and `mongorestore`.

### Fixed Bugs 🐛

- Revert "Remove hack for `bypassEmptyTsReplacement` (#5217)" by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5256

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/78?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v2.3.0...v2.3.1).

## [v2.3.0](https://github.com/FerretDB/FerretDB/releases/tag/v2.3.0) (2025-06-11)

This version works best with
[DocumentDB v0.104.0-ferretdb-2.3.0](https://github.com/FerretDB/documentdb/releases/tag/v0.104.0-ferretdb-2.3.0).

### New Features 🎉

- Add telemetry configuration for embedded FerretDB by @jyz0309 in https://github.com/FerretDB/FerretDB/pull/5109
- Use DocumentDB's support for `bypassEmptyTsReplacement` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5217
- Use DocumentDB's `compat` command by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5242

### Documentation 📄

- Enable `onInlineAuthors` in blog settings by @Balou9 in https://github.com/FerretDB/FerretDB/pull/5079
- Create compatible apps section by @Fashander in https://github.com/FerretDB/FerretDB/pull/5108
- Improve insert operations in guides by @Fashander in https://github.com/FerretDB/FerretDB/pull/5131
- Update versions to point to the next release by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5144
- Update Docker tags in docs by @Fashander in https://github.com/FerretDB/FerretDB/pull/5156
- Update Docker tags for next release by @Fashander in https://github.com/FerretDB/FerretDB/pull/5157
- Do not format MDX as Markdown by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5186
- Add blog post on Mongo Express compatibility by @Fashander in https://github.com/FerretDB/FerretDB/pull/5194
- Add blog post for Novu compatibility by @Fashander in https://github.com/FerretDB/FerretDB/pull/5195
- Update CTS tool to wrap long lines by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5202
- Disable Markdownlint rule that clashes with MDX by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5210
- Small documentation tweaks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5218
- Prepare v2.3.0 release by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5254

### Other Changes 🤖

- Use `runtime.Cleanup` for resource tracking by @sahinakyol in https://github.com/FerretDB/FerretDB/pull/5077
- Add Dependabot configuration for `main-v1` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5098
- Bump deps by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5159
- Make `conninfo` a resource by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5166
- Revert "Update CODEOWNERS" by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5169
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5170
- Add `permissions` to GitHub Actions workflows by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5182
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5184
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5197
- Add test for inserting zero timestamp by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5199
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5201
- Refactor resource tracking tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5208
- Unskip test by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5211
- Use Ollama in local setup by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5215
- Update TODO comments by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5223
- Bump Go and deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5236
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5240
- Build production Docker images for PRs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5251
- Improve telemetry configuration for embedded FerretDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5252

### New Contributors

- @sahinakyol made their first contribution in https://github.com/FerretDB/FerretDB/pull/5077
- @Balou9 made their first contribution in https://github.com/FerretDB/FerretDB/pull/5079

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/74?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v2.2.0...v2.3.0).

## [v2.2.0](https://github.com/FerretDB/FerretDB/releases/tag/v2.2.0) (2025-05-09)

This version works best with
[DocumentDB v0.103.0-ferretdb-2.2.0](https://github.com/FerretDB/documentdb/releases/tag/v0.103.0-ferretdb-2.2.0).

### New Features 🎉

- Add full arm64 support by @AlekSi, @chilagrow in https://github.com/FerretDB/FerretDB/pull/5113
- Rename old `ferretdb-eval` image to `ferretdb-eval-dev` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5040
- Provide `ferretdb-eval` Docker image with production build by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5051
- Supervise services in evaluation images by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5041
- Use volume for `state` directory by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5083

### Enhancements 🛠

- Rename binaries and packages by @vardbabayan in https://github.com/FerretDB/FerretDB/pull/5078
- Decode `dropIndexes` response by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5117

### Documentation 📄

- Update DocumentDB debian packages by @Fashander in https://github.com/FerretDB/FerretDB/pull/4959
- Add Kubernetes installation guide by @Fashander in https://github.com/FerretDB/FerretDB/pull/4971
- Add blog post on FerretDB and CNPG by @Fashander in https://github.com/FerretDB/FerretDB/pull/4998
- Update Docker images by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5015
- Add blog post on migrating to FerretDB with dsync by @Fashander in https://github.com/FerretDB/FerretDB/pull/5033
- Use &#34;PostgreSQL with DocumentDB extension&#34; phrase by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5050
- Backport Kubernetes docs to 2.1 by @Fashander in https://github.com/FerretDB/FerretDB/pull/5053
- Rename and move code files in guides by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5057
- Update redirects by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5060
- Update OpenAPI spec description by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5062
- Document Data API usage by @Fashander in https://github.com/FerretDB/FerretDB/pull/5063
- Backport `.deb` installation guide by @Fashander in https://github.com/FerretDB/FerretDB/pull/5066
- Unconvert syntax examples in documentation by @noisersup in https://github.com/FerretDB/FerretDB/pull/5072
- Reformat linter by @Fashander in https://github.com/FerretDB/FerretDB/pull/5075
- Update evaluation Docker image documentation by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5076
- Document required features by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5086
- Document update process for new releases by @Fashander in https://github.com/FerretDB/FerretDB/pull/5124
- Remove old TODOs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5132

### Other Changes 🤖

- Remove MongoDB driver v1 by @KrishnaSindhur in https://github.com/FerretDB/FerretDB/pull/4961
- Hook CTS tool into documentation building by @noisersup in https://github.com/FerretDB/FerretDB/pull/4990
- Do not send zero values to telemetry by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5016
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5020
- Update deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5024
- Minor tweaks for proxy code by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5025
- Move tests for sessions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5026
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5027
- Merge Msg and Query into a single type by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5032
- Make proxy handler implement an interface by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5036
- Move `findAndModify` integration tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5037
- Update wire by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5042
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5045
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5054
- Implement handler `Handle` function by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5055
- Use new `wire` helpers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5056
- Fix docker health check on evaluation image by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5067
- Define docker tags for `ferretdb-eval` image by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5068
- Use `Handle` function in Data API by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5069
- Use `Handle` function in `clientconn` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5070
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5074
- Skip `bypassEmptyTsReplacement` parameters for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5080
- Reorgonize cts output files by @noisersup in https://github.com/FerretDB/FerretDB/pull/5082
- Unexport handlers by @chilagrow in https://github.com/FerretDB/FerretDB/pull/5084
- Use new `wire` helpers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5094
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5096
- Fix Docker tags for pre-release git tags by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5099
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5115
- Hook CTS into CI by @noisersup in https://github.com/FerretDB/FerretDB/pull/5120
- Run CTS tests against Full Text Search guide examples by @noisersup in https://github.com/FerretDB/FerretDB/pull/5121
- Run CTS tests against Vector Search guide examples by @noisersup in https://github.com/FerretDB/FerretDB/pull/5122
- Run CTS tests against TTL Indexes guide examples by @noisersup in https://github.com/FerretDB/FerretDB/pull/5123
- Remove `$db` parameter from generated mongosh requests by @noisersup in https://github.com/FerretDB/FerretDB/pull/5125
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5141
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5142
- Prepare v2.2.0 release by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5138

### New Contributors

- @vardbabayan made their first contribution in https://github.com/FerretDB/FerretDB/pull/5078

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/73?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v2.1.0...v2.2.0).

## [v2.1.0](https://github.com/FerretDB/FerretDB/releases/tag/v2.1.0) (2025-04-03)

This version works only with
[DocumentDB v0.102.0-ferretdb-2.1.0](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.1.0).

### Breaking changes

<!-- textlint-disable one-sentence-per-line -->

> [!CAUTION]
> Please note that due to incompatibilities in our previous releases, DocumentDB can't be updated in place,
> even with a manual `ALTER EXTENSION UPDATE` query or other means.
> A new clean installation into an empty data directory/volume is required.
> All data should be backed up with `mongodump`/`mongoexport` before and restored
> with `mongorestore`/`mongoimport` after.
> [See our blog post for more details](https://blog.ferretdb.io/ferretdb-v210-release-performance-improvements-bug-fixes/).
>
> We expect future updates to be much smoother.

<!-- textlint-enable one-sentence-per-line -->

### Fixed Bugs 🐛

- Fix version detection for embeddable package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4936

### Enhancements 🛠

- Add colorized levels to `console` logger by @noisersup in https://github.com/FerretDB/FerretDB/pull/4904
- Improve `--help` output by @KrishnaSindhur in https://github.com/FerretDB/FerretDB/pull/4918
- Add support for reading PostgreSQL URL from a file by @KrishnaSindhur in https://github.com/FerretDB/FerretDB/pull/4937
- Do not decode incoming document twice by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4981

### Documentation 📄

- Add post on MongoDB queries and operations on FerretDB by @Fashander in https://github.com/FerretDB/FerretDB/pull/4732
- Add example telemetry report to documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4861
- Add FerretDB v2 GA blog post by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4884
- Add full text search guide by @Fashander in https://github.com/FerretDB/FerretDB/pull/4886
- Add guide for GUI apps by @Fashander in https://github.com/FerretDB/FerretDB/pull/4906
- Add TTL index guide by @Fashander in https://github.com/FerretDB/FerretDB/pull/4926
- Update `deb` and `rpm` installation docs by @Fashander in https://github.com/FerretDB/FerretDB/pull/4927
- Sync flags grouping with docs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4933
- Add a note to documentation about PR titles by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4944
- Update/backport full-text and TTL indexes guides by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4948
- Backport documentation changes to v2.0 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4952
- Add basic documentation for supported commands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4970
- Update feature blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/4991
- Add blogpost on FerretDB v2.1.0 release by @Fashander in https://github.com/FerretDB/FerretDB/pull/5004
- Create redirects for `/v2.0/` documentation URLs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5011

### Other Changes 🤖

- Update changelog generator by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4810
- Update TODO URLs for `listDatabase` tests by @noisersup in https://github.com/FerretDB/FerretDB/pull/4863
- Document non-enforced format of log messages in `envtool` package by @noisersup in https://github.com/FerretDB/FerretDB/pull/4867
- Start working on a new release by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4885
- Adjust pool connection timeout by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4888
- Check issue URLs in documentation by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4890
- Add TODO comments for observability tasks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4896
- Refactor `clientconn` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4907
- Add basic structure for middlewares by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4908
- Make production builds of the `main` branch by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4911
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4916
- Move message processing in `clientconn` to a function by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4919
- Change the way OP_MSG handlers are invoked by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4923
- Improve integration benchmarks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4925
- Configure connection pool size in tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4932
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4941
- Handle `OP_QUERY` in middleware using interface by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4947
- Implement error middleware by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4951
- Use ERROR level logging for failed tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4974
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4975
- Disable tracing in benchmarks for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4980
- Update DocumentDB by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4982
- Add TODOs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4986
- Partially revert middleware changes by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4995
- Add more tests for error handling by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4996
- Use Go 1.24.2 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5005
- Remove error middleware for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5008
- Disable commit check for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/5012

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/72?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v2.0.0...v2.1.0).

## [v2.0.0](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0) (2025-03-05)

The first generally available release of FerretDB v2,
powered by [DocumentDB PostgreSQL extension](https://github.com/microsoft/documentdb)!

This version works best with
[DocumentDB v0.102.0-ferretdb-2.0.0](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.0.0).

[Release blog post](https://blog.ferretdb.io/ferretdb-v2-ga-open-source-mongodb-alternative-ready-for-production/).

### Documentation 📄

- Add migration guide from v1.x to v2.x by @Fashander in https://github.com/FerretDB/FerretDB/pull/4850
- Add basic troubleshooting guide by @Fashander in https://github.com/FerretDB/FerretDB/pull/4854
- Final preparations by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4874

### Other Changes 🤖

- Unskip tests that refer to closed issue by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4865
- Use GitHub-hosted CI runners by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4866
- Skip flaky `currentOp` test and add TODO by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4872
- Bump Go and safe deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4875

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/69?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v2.0.0-rc.5...v2.0.0).

## [v2.0.0-rc.5](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0-rc.5) (2025-03-04)

This is the final release candidate before the GA release tomorrow!
Most users don't need to update.

This version works best with
[DocumentDB v0.102.0-ferretdb-2.0.0-rc.5](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.0.0-rc.5).

### New Features 🎉

- Use DocumentDB API for `listDatabases` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4841
- Implement `currentOp` command by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4855

### Fixed Bugs 🐛

- Make Data API work without authentication if requested by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4849

### Enhancements 🛠

- Publish build info, state, and CLI flags in expvar by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4837
- Enforce log messages format in development builds by @noisersup in https://github.com/FerretDB/FerretDB/pull/4754
- Filter out sensitive information from debug archive by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4836
- Improve messages about DocumentDB version by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4832

### Documentation 📄

- Move comment in docs to fix DocCard by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4845
- Update URL to join Slack by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4851
- Recommend using full tags/versions by @Fashander in https://github.com/FerretDB/FerretDB/pull/4834

### Other Changes 🤖

- Change `dbStats` tests TODO links by @noisersup in https://github.com/FerretDB/FerretDB/pull/4823
- Update TODO comment by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4831
- Bump DocumentDB for development by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4839
- Update DocumentDB Docker image by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4853
- Remove old code for tracking operations by @AlekSi in https://github.com/FerretDB/FerretDB/pull/4842
- Remove DocumentDB building from this repo by @chilagrow in https://github.com/FerretDB/FerretDB/pull/4852

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/71?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v2.0.0-rc.2...v2.0.0-rc.5).

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
