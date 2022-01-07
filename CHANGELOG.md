# Changelog

## [v0.0.5](https://github.com/FerretDB/FerretDB/releases/tag/v0.0.5) (2022-01-04)

### New Features 🎉
* Add basic metrics by @AlekSi in https://github.com/FerretDB/FerretDB/pull/108
* Implement `serverStatus` command by @jyz0309 in https://github.com/FerretDB/FerretDB/pull/116
* Implement `dropDatabase` command by @radmirnovii in https://github.com/FerretDB/FerretDB/pull/117
* Support count function by @thuan1412 in https://github.com/FerretDB/FerretDB/pull/97
* Implement `getParameter` command by @jyz0309 in https://github.com/FerretDB/FerretDB/pull/142
* Support `limit` parameter in delete by @OpenSauce in https://github.com/FerretDB/FerretDB/pull/141
* Implement basic `create` command by @ekalinin in https://github.com/FerretDB/FerretDB/pull/184
* Build Docker image with GitHub Actions by @pboros in https://github.com/FerretDB/FerretDB/pull/189
* Automatically create databases by @AlekSi in https://github.com/FerretDB/FerretDB/pull/185
* Support `hello` command by @AlekSi in https://github.com/FerretDB/FerretDB/pull/195
* Add stub for `createindexes` command by @AlekSi in https://github.com/FerretDB/FerretDB/pull/196
* Support basic `hostInfo` command by @ekalinin in https://github.com/FerretDB/FerretDB/pull/188
* Support `collStats` command by @ekalinin in https://github.com/FerretDB/FerretDB/pull/206
### Fixed Bugs 🐛
* Accept $ and . in object field names by @AlekSi in https://github.com/FerretDB/FerretDB/pull/127
* Make `checkConnection` less strict for common UTF8 localizations by @klokar in https://github.com/FerretDB/FerretDB/pull/135
* Wait for PostgreSQL on `make env-up` by @agneum in https://github.com/FerretDB/FerretDB/pull/149
* Fix build info parsing by @AlekSi in https://github.com/FerretDB/FerretDB/pull/205
* Fix GetLog & add missed test for it by @ekalinin in https://github.com/FerretDB/FerretDB/pull/211
### Enhancements 🛠
* Return version in `serverStatus` command by @AlekSi in https://github.com/FerretDB/FerretDB/pull/121
* Improve output of buildInfo command by @GinGin3203 in https://github.com/FerretDB/FerretDB/pull/204
### Documentation 📄
* CONTRIBUTING.md: fix typo & add clonning section by @ekalinin in https://github.com/FerretDB/FerretDB/pull/114
* CONTRIBUTING.md: fix "/user/.../" -> "/usr/.../" by @GinGin3203 in https://github.com/FerretDB/FerretDB/pull/137
* Add community links by @AlekSi in https://github.com/FerretDB/FerretDB/pull/180
### Other Changes 🤖
* Add convention for Decimal128 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/103
* Bump github.com/jackc/pgx/v4 from 4.14.0 to 4.14.1 by @dependabot in https://github.com/FerretDB/FerretDB/pull/99
* Build multi-arch Docker images by @AlekSi in https://github.com/FerretDB/FerretDB/pull/107
* Verify modules on `make init` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/123
* Enable go-consistent linter by @AlekSi in https://github.com/FerretDB/FerretDB/pull/124
* Use composite GitHub Action for Go setup. (#122) by @klokar in https://github.com/FerretDB/FerretDB/pull/126
* Use shared setup-go action by @AlekSi in https://github.com/FerretDB/FerretDB/pull/131
* Add an option to use read-only user in tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/132
* Refactor handler tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/136
* Bump MongoDB and test_db versions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/139
* Remove old hack by @AlekSi in https://github.com/FerretDB/FerretDB/pull/144
* Enable goheader linter by @AlekSi in https://github.com/FerretDB/FerretDB/pull/145
* Cleanups and fixes by @AlekSi in https://github.com/FerretDB/FerretDB/pull/146
* Use `any` instead of `interface{}` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/147
* Tweak storage by @AlekSi in https://github.com/FerretDB/FerretDB/pull/148
* Add helpers for accessing objects by paths by @AlekSi in https://github.com/FerretDB/FerretDB/pull/140
* Bump mvdan.cc/gofumpt from 0.2.0 to 0.2.1 in /tools by @dependabot in https://github.com/FerretDB/FerretDB/pull/186
* Add and use schema and table helpers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/191
* Refactor / cleanup tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/192
* Add missed test for `buildInfo` command by @ekalinin in https://github.com/FerretDB/FerretDB/pull/187
* Refactor slice / `types.Array` type by @AlekSi in https://github.com/FerretDB/FerretDB/pull/202
* Bump golang.org/x/text from 0.3.6 to 0.3.7 by @dependabot in https://github.com/FerretDB/FerretDB/pull/208
* Setup changelog generation by @ekalinin in https://github.com/FerretDB/FerretDB/pull/209
* Build containers for branches as well by @pboros in https://github.com/FerretDB/FerretDB/pull/213
* Container builds for PRs and tags by @pboros in https://github.com/FerretDB/FerretDB/pull/215
* Use our own action for extracting Docker tag by @AlekSi in https://github.com/FerretDB/FerretDB/pull/219

## New Contributors
* @AlekSi made their first contribution in https://github.com/FerretDB/FerretDB/pull/103
* @ekalinin made their first contribution in https://github.com/FerretDB/FerretDB/pull/114
* @jyz0309 made their first contribution in https://github.com/FerretDB/FerretDB/pull/116
* @radmirnovii made their first contribution in https://github.com/FerretDB/FerretDB/pull/117
* @klokar made their first contribution in https://github.com/FerretDB/FerretDB/pull/126
* @GinGin3203 made their first contribution in https://github.com/FerretDB/FerretDB/pull/137
* @agneum made their first contribution in https://github.com/FerretDB/FerretDB/pull/149
* @pboros made their first contribution in https://github.com/FerretDB/FerretDB/pull/189

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/4?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.0.4...v0.0.5).


## [v0.0.4](https://github.com/FerretDB/FerretDB/releases/tag/v0.0.4) (2021-12-01)

* A new name! ([see here](https://github.com/FerretDB/FerretDB/discussions/100))
* Added support for databases sizes in `listDatabases` command ([#61](https://github.com/FerretDB/FerretDB/issues/61), thanks to [Leigh](https://github.com/OpenSauce)).

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/3?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.0.3...v0.0.4).


## [v0.0.3](https://github.com/FerretDB/FerretDB/releases/tag/v0.0.3) (2021-11-19)

* Added support for `$regex` evaluation query operator. ([#28](https://github.com/FerretDB/FerretDB/issues/28))
* Fixed handling of Infinity, -Infinity, NaN BSON Double values. ([#29](https://github.com/FerretDB/FerretDB/issues/29))
* Improved documentation for contributors (thanks to [Leigh](https://github.com/OpenSauce)).

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/2?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.0.2...v0.0.3).


## [v0.0.2](https://github.com/FerretDB/FerretDB/releases/tag/v0.0.2) (2021-11-13)

* Added support for comparison query operators: `$eq`,`$gt`,`$gte`,`$in`,`$lt`,`$lte`,`$ne`,`$nin`. ([#26](https://github.com/FerretDB/FerretDB/issues/26))
* Added support for logical query operators: `$and`, `$not`, `$nor`, `$or`. ([#27](https://github.com/FerretDB/FerretDB/issues/27))

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/1?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.0.1...v0.0.2).


## [v0.0.1](https://github.com/FerretDB/FerretDB/releases/tag/v0.0.1) (2021-11-01)

* Initial public release!
