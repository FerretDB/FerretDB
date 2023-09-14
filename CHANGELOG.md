# Changelog

<!-- markdownlint-disable MD024 MD034 -->

## [v1.10.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.10.0) (2023-09-14)

### What's Changed

With this release, the SQLite backend support is officially out of beta,
on par with our PostgreSQL backend, and fully supported!

### New Features 🎉

- Implement `aggregate` for SQLite by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3256
- Implement `collStats` for SQLite by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3295
- Implement `createIndexes` for SQLite by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/3304
- Implement `dbStats` for SQLite by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3270
- Implement `distinct` for SQLite by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3265
- Implement `dropIndexes` for SQLite by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/3329
- Implement `explain` command for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3264
- Implement `findAndModify` for SQLite by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3302
- Implement `getLog` for SQLite by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/3279
- Implement `listDatabases` for SQLite by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/3269
- Implement `listIndexes` for SQLite by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/3301
- Implement `renameCollection` for SQLite by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3321
- Implement `serverStatus` and `dataSize` commands for SQLite by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3316
- Support `_id` implicit filter for `ObjectID` in SQLite by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3330
- Support `$bit` bitwise update operator by @adetunjii in https://github.com/FerretDB/FerretDB/pull/3266
- Support `ordered` `insert`s for SQLite by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3223
- Build Docker images for `arm/v6` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3357

### Enhancements 🛠

- Make `delete`s atomic for SQLite by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3291
- Make `update`s atomic for SQLite by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3296
- Do not change `search_path` parameter by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3299

### Documentation 📄

- Cleanup `$bit` update operator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3315
- Document how to test for compatibility by @b1ron in https://github.com/FerretDB/FerretDB/pull/3268
- Update blog writing guide documentation by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3209
- Update category links in writing guide by @Fashander in https://github.com/FerretDB/FerretDB/pull/3323
- Update deb.md - minor grammar correction by @athkishore in https://github.com/FerretDB/FerretDB/pull/3289
- Update the writing guide by @Fashander in https://github.com/FerretDB/FerretDB/pull/3311

### Other Changes 🤖

- Add ability to freeze `*types.Document` and `*types.Array` by @KrishnaSindhur in https://github.com/FerretDB/FerretDB/pull/3253
- Add backend decorators and OpLog stub by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3303
- Add backend interface for `collStats` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3294
- Add backend interface for `dbStats` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3267
- Add more tests by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3336
- Add new PostgreSQL backend stub by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3319
- Add tests for accessing aggregation variable `$$ROOT` field by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3254
- Add tests for validation bug by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3286
- Add transactions to `fsql` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3278
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3284
- Clean-up `*types.Timestamp` a bit by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3305
- Do not `ConsumeValues` in the `$group` aggregation stage by @adetunjii in https://github.com/FerretDB/FerretDB/pull/3344
- Expand architecture docs, add comments by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3288
- Fix params handling for `dropIndexes` implementation for SQLite by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/3350
- Make registry return full collection info by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3292
- Remove `Database.Close` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3327
- Remove duplicated `$expr` tests by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3255
- Return correct response if unique index violation happened on SQLite backend by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/3353
- Simplify and deprecate `commonerrors.WriteErrors` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3258
- Skip tests for `enable` `setFreeMonitoring` for MongoDB by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3318
- Tweak MongoDB initialization process by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3307
- Update TODO comments by @noisersup in https://github.com/FerretDB/FerretDB/pull/3262
- Use `pkgsite` instead of `godoc` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3326
- Use Go 1.21 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3324

### New Contributors

- @athkishore made their first contribution in https://github.com/FerretDB/FerretDB/pull/3289

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/51?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.9.0...v1.10.0).

## [v1.9.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.9.0) (2023-08-28)

### Enhancements 🛠

- Add more metrics for `*sql.DB` by @slavabobik in https://github.com/FerretDB/FerretDB/pull/3230

### Documentation 📄

- Add blog post for FerretDB v1.8.0 by @Fashander in https://github.com/FerretDB/FerretDB/pull/3198
- Fix typos in documentation by @pratikmota in https://github.com/FerretDB/FerretDB/pull/3217
- Make the writing guide accessible but unlisted by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3221
- Add blogpost on Leafcloud by @Fashander in https://github.com/FerretDB/FerretDB/pull/3153
- Add Postgres Ibiza event blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/3210
- Add Civo Navigate event blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/3201

### Other Changes 🤖

- Configure repo settings with files by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3208
- Update `go-hdb` to v1.4.1 by @aenkya in https://github.com/FerretDB/FerretDB/pull/3213
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3215
- Add another stress test for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3195
- Improve building with test coverage information by @durgakiran in https://github.com/FerretDB/FerretDB/pull/3059
- Fix concurrent SQLite tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3222
- Refactor aggregation operators by @noisersup in https://github.com/FerretDB/FerretDB/pull/3188
- Add stubs for `renameCollection` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3233
- Update issue links by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3234
- Add stubs for `explain` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3236
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3248
- Add linter for issue comments by @KrishnaSindhur in https://github.com/FerretDB/FerretDB/pull/3154
- Simplify `commonerrors` package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3227
- Publish Docker images on quay.io by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3250
- Refactor aggregation accumulators by @noisersup in https://github.com/FerretDB/FerretDB/pull/3203
- Add new PostgreSQL backend stub by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3251
- Cleanup SQLite tests by @noisersup in https://github.com/FerretDB/FerretDB/pull/3246

### New Contributors

- @aenkya made their first contribution in https://github.com/FerretDB/FerretDB/pull/3213
- @pratikmota made their first contribution in https://github.com/FerretDB/FerretDB/pull/3217
- @durgakiran made their first contribution in https://github.com/FerretDB/FerretDB/pull/3059
- @slavabobik made their first contribution in https://github.com/FerretDB/FerretDB/pull/3230

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/50?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.8.0...v1.9.0).

## [v1.8.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.8.0) (2023-08-14)

### New Features 🎉

- Implement `$group` stage `_id` expression by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3138
- Implement `$expr` evaluation query operator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3163

### Fixed Bugs 🐛

- Do not return immutable `_id` error from `findAndModify` for upserting same `_id` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3171

### Enhancements 🛠

- Cache SQLite tables metadata by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3124
- Use lock for SQLite metadata by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3146

### Other Changes 🤖

- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3142
- Improve MongoDB/FerretDB error checking in tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3143
- Expect some `aggregate` and `insert` tests to fail for SQLite by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3147
- Make administration command integration tests pass for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3152
- Bump deps, including Go by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3160
- Make aggregate stats integration tests pass for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3151
- Simplify tests a bit by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3149
- Make `distinct` and `explain` command integration tests pass for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3159
- Use one implementation for finding path values by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3087
- Make aggregate documents compat tests pass for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3150
- Make `query` integration tests pass for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3182
- Make `findAndModify` integration tests pass for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3173
- Make index integration tests pass for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3185
- Add tests for `$$ROOT` aggregation expression variable by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3180
- Make `getMore` integration tests pass for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3174
- Make `update` integration tests pass for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/3184
- Add tests for `$$ROOT` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3187

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/49?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.7.0...v1.8.0).

## [v1.7.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.7.0) (2023-07-31)

### New Features 🎉

- Implement `$sum` aggregation standard operator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3063

### Fixed Bugs 🐛

- Fix `PLAIN` auth with C# driver by @b1ron in https://github.com/FerretDB/FerretDB/pull/3012

### Enhancements 🛠

- Add validating max nested document/array depth by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2882
- Validate database and collection names for SQLite handler by @noisersup in https://github.com/FerretDB/FerretDB/pull/2868
- Add basic metrics, logging and tracing for SQLite by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3123
- Tweak and document SQLite URI parameters by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3128

### Documentation 📄

- Add blog post for FerretDB v1.6.0 by @Fashander in https://github.com/FerretDB/FerretDB/pull/3058
- Update changelog by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3072
- Update blog post for FerretDB v1.6.0 by @Fashander in https://github.com/FerretDB/FerretDB/pull/3073
- Tweak documentation and blog by @Fashander in https://github.com/FerretDB/FerretDB/pull/2992
- Add blog post on "Community matters: fireside chat with Artem Ervits, CockroachDB" by @Fashander in https://github.com/FerretDB/FerretDB/pull/3066
- Update Blog Post by @Fashander in https://github.com/FerretDB/FerretDB/pull/3086
- Update tags formatting in writing guide by @Fashander in https://github.com/FerretDB/FerretDB/pull/3097
- Add blog post on "Using Mingo with FerretDB" by @Fashander in https://github.com/FerretDB/FerretDB/pull/3074
- Simplify `checkdocs` linter by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3104
- Update MongoDB comparision blog post by @ptrfarkas in https://github.com/FerretDB/FerretDB/pull/3117
- Update MongoDB comparision blog post by @ptrfarkas in https://github.com/FerretDB/FerretDB/pull/3119
- Add blog post on Grafana Monitoring for FerretDB by @Fashander in https://github.com/FerretDB/FerretDB/pull/3106

### Other Changes 🤖

- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3064
- Mark some tests as failing for SQLite by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3051
- Improve sjson package fuzzing by @quasilyte in https://github.com/FerretDB/FerretDB/pull/3071
- Merges fuzztool into envtool by @Aditya1404Sal in https://github.com/FerretDB/FerretDB/pull/2645
- Do not import `commonerrors` in tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3081
- Remove dead code by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3093
- Allow to change SQLite URI in tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3092
- Replace test doubles with constants by @noisersup in https://github.com/FerretDB/FerretDB/pull/3024
- Improve `checkdocs` linter by @KrishnaSindhur in https://github.com/FerretDB/FerretDB/pull/3095
- Add daily progress principle to `PROCESS.md` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/3098
- Support `_id` aggregation operators for `$group` stage by @noisersup in https://github.com/FerretDB/FerretDB/pull/3096
- Bump the tools group in /tools with 1 update by @dependabot in https://github.com/FerretDB/FerretDB/pull/3109
- Backport v1.6.1 fixes by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3107
- Support recursive operator calls for `$sum` aggregation accumulator by @noisersup in https://github.com/FerretDB/FerretDB/pull/3116

### New Contributors

- @Aditya1404Sal made their first contribution in https://github.com/FerretDB/FerretDB/pull/2645
- @KrishnaSindhur made their first contribution in https://github.com/FerretDB/FerretDB/pull/3095
- @ptrfarkas made their first contribution in https://github.com/FerretDB/FerretDB/pull/3117

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/47?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.6.1...v1.7.0).

## [v1.6.1](https://github.com/FerretDB/FerretDB/releases/tag/v1.6.1) (2023-07-26)

### Fixed Bugs 🐛

- Fix pushdown for `find` with `filter` and `limit` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3114

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/48?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.6.0...v1.6.1).

## [v1.6.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.6.0) (2023-07-17)

### New Features 🎉

- Implement `killCursors` command by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2939
- Implement `ping` command for SQLite by @noisersup in https://github.com/FerretDB/FerretDB/pull/2965
- Implement `getParameter` method for SQLite by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2985

### Fixed Bugs 🐛

- Ignore `lsid` field in all commands by @b1ron in https://github.com/FerretDB/FerretDB/pull/3010
- Allow `$set` operator to update `_id` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3009
- Apply pushdown for `limit` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2993
- Fix `update` with query operator for `upsert` option by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3028

### Enhancements 🛠

- Add integration tests for `maxTimeMS` in `find`, `aggregate` and `getMore` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2953
- Remove double decoding in unmarshalSingleValue by @quasilyte in https://github.com/FerretDB/FerretDB/pull/3018
- Ignore `count.fields` argument by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3048

### Documentation 📄

- Add blog post on FerretDB release v1.5.0 by @Fashander in https://github.com/FerretDB/FerretDB/pull/2958
- Mention SQLite in README.md by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2968
- Add blog post about using NoSQLBooster with FerretDB by @Fashander in https://github.com/FerretDB/FerretDB/pull/2962
- Update blog post image by @Fashander in https://github.com/FerretDB/FerretDB/pull/3029
- Add a note about setting the stable API version by @b1ron in https://github.com/FerretDB/FerretDB/pull/3035
- Add blog post on "How to run FerretDB on top of StackGres" by @Fashander in https://github.com/FerretDB/FerretDB/pull/2869
- Fix blog post formatting by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3047
- Update database naming restrictions by @b1ron in https://github.com/FerretDB/FerretDB/pull/3042

### Other Changes 🤖

- Move `find` and `aggregation` cursor integration tests to `getMore` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2952
- Make a copy of the `testing.TB` interface by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2987
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2998
- Remove Tigris from documentation and builds by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2999
- Remove Tigris code by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3001
- Remove Tigris from tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3002
- Crush PNG files to make them smaller by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3020
- Update issue URL by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3021
- Move `testutil.TB` to `testtb.TB` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3022
- Move `logout` to `commoncommands` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3019
- Make `task all` run only unit tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3023
- Update closed issue links by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3027
- Unskip `findAndModify` `$set` integration test for `_id` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3025
- Expect `renameCollection` tests failures by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3007
- Fix `killCursors` edge case by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3030
- Fix error checking in backend contracts by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3031
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3034
- Remove `Type()` interface from aggregation stage by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3045
- Remove fixed issue link and clean up integration test provider setup by @chilagrow in https://github.com/FerretDB/FerretDB/pull/3052
- Prepare v1.6.0 release by @AlekSi in https://github.com/FerretDB/FerretDB/pull/3056

### New Contributors

- @quasilyte made their first contribution in https://github.com/FerretDB/FerretDB/pull/3018

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/46?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.5.0...v1.6.0).

## [v1.5.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.5.0) (2023-07-03)

### What's Changed

This release provides beta-level support for the SQLite backend.
There is some missing functionality, but it is ready for early adopters.

This release provides improved cursor support, enabling commands like `find` and `aggregate` to return large data sets much more effectively.

Tigris data users: Please note that this is the last release of FerretDB which includes support for the Tigris backend.
Starting from FerretDB v1.6.0, Tigris will not be supported.
If you wish to use Tigris, please do not update FerretDB beyond v1.5.0.
This and earlier versions of FerretDB with Tigris support will still be available on GitHub.

### New Features 🎉

- Implement `count` for SQLite by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2865
- Enable cursor support for PostgreSQL and SQLite by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2864

### Enhancements 🛠

- Support `find` `singleBatch` and validate `getMore` parameters by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2855
- Support cursors for aggregation pipelines by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2861
- Fix collection name starting with dot validation by @noisersup in https://github.com/FerretDB/FerretDB/pull/2912
- Improve validation for `createIndexes` and `dropIndexes` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2884
- Use cursors in `find` command by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2933

### Documentation 📄

- Add blogpost on FerretDB v1.4.0 by @Fashander in https://github.com/FerretDB/FerretDB/pull/2858
- Add blog post on "Meet FerretDB at Percona University in Casablanca and Belgrade" by @Fashander in https://github.com/FerretDB/FerretDB/pull/2870
- Update supported commands by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2876
- Add blog post "FerretDB Demo: Launch and Test a Database in Minutes" by @Fashander in https://github.com/FerretDB/FerretDB/pull/2851
- Fix Github link for Dance repository by @Matthieu68857 in https://github.com/FerretDB/FerretDB/pull/2887
- Add blog post on "How to Configure FerretDB to work on Percona Distribution for PostgreSQL" by @Fashander in https://github.com/FerretDB/FerretDB/pull/2911
- Update incorrect blog post image by @Fashander in https://github.com/FerretDB/FerretDB/pull/2920
- Crush PNG images by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2931

### Other Changes 🤖

- Add more validation and tests for `$unset` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2853
- Make it easier to debug GitHub Actions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2860
- Unify tests for indexes by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2866
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2875
- Fix fuzzing corpus collection by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2879
- Add basic tests for iterators by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2880
- Implement basic `insert` support for SAP HANA by @polyal in https://github.com/FerretDB/FerretDB/pull/2732
- Update contributing docs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2828
- Improve `wire` and `sjson` fuzzing by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2883
- Add operators support for `$addFields` by @noisersup in https://github.com/FerretDB/FerretDB/pull/2850
- Unskip test that passes now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2885
- Tweak contributing guidelines by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2886
- Add handler's metrics registration by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2895
- Clean-up some code and comments by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2904
- Fix cancelation signals propagation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2908
- Bump deps, add permissions monitoring by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2930
- Fix integration tests after bumping deps by @noisersup in https://github.com/FerretDB/FerretDB/pull/2934
- Update benchmark to use cursors by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2932
- Set `minWireVersion` to 0 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2937
- Test `getMore` integration test using one connection pool by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2878
- Add better metrics for connections by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2938
- Use cursors with iterator in `aggregate` command by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2929
- Implement proper response for `createIndexes` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2936
- Re-implement `DELETE` for SQLite backend by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2907
- Validate database names for SQLite handler by @noisersup in https://github.com/FerretDB/FerretDB/pull/2924
- Add `insert` documents type validation by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2946
- Convert SQLite directory to URI by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2922
- Do not break fuzzing initialization by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2951

### New Contributors

- @Matthieu68857 made their first contribution in https://github.com/FerretDB/FerretDB/pull/2887

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/45?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.4.0...v1.5.0).

## [v1.4.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.4.0) (2023-06-19)

### New Features 🎉

- Implement `$type` aggregation operator by @noisersup in https://github.com/FerretDB/FerretDB/pull/2789
- Implement `$unset` aggregation pipeline stage by @shibasisp in https://github.com/FerretDB/FerretDB/pull/2676
- Implement simple `$addFields/$set` aggregation pipeline stages by @shibasisp in https://github.com/FerretDB/FerretDB/pull/2783
- Implement `createIndexes` for unique indexes by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2814

### Documentation 📄

- Add blog post for FerretDB v1.3.0 by @Fashander in https://github.com/FerretDB/FerretDB/pull/2791
- Add `release` tag to release blog post by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2792
- Add textlint rules for en dashes and em dashes by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2823
- Add Blog Post on Document Databases by @Fashander in https://github.com/FerretDB/FerretDB/pull/2204
- Add user documentation about unique index creation by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2856

### Other Changes 🤖

- Make `testutil.Logger` easier to use by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2790
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2798
- Refactor SQLite handler by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2731
- Merge test workflows to fix coverage calculation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2801
- Improve `testDistinctCompat` by @noisersup in https://github.com/FerretDB/FerretDB/pull/2782
- Use iterator in `$sum` aggregation accumulator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2799
- Bump Go to 1.20.5 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2810
- Fix free monitoring tests for MongoDB 6.0.6 by @jeremyphua in https://github.com/FerretDB/FerretDB/pull/2784
- Bump MongoDB to 6.0.6 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2727
- Bump MongoDB Go driver by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2817
- Implement `envtool tests shard` command by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2773
- Check error message in non compat integration tests by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2806
- Shard integration tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2820
- Describe current test naming conventions in the contributing guidelines by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2821
- Add tests for `find`/`getMore` `batchSize` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2825
- Add more test cases for index validation by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2752
- Fix running single test with `task` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2832
- Refactor `getWholeParamStrict` and `GetScaleParam` functions by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2831
- Prevent tests deadlock when backend is down by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2846
- Fix `unimplemented-non-default` tag usages by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2848
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2849
- Add more tests for `$set` and `$addFields` aggregation stages by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2844
- Improve benchmarks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2833
- Handle `$type` aggregation operator errors properly by @noisersup in https://github.com/FerretDB/FerretDB/pull/2829

### New Contributors

- @shibasisp made their first contribution in https://github.com/FerretDB/FerretDB/pull/2676

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/44?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.3.0...v1.4.0).

## [v1.3.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.3.0) (2023-06-05)

### New Features 🎉

- Implement positional operator in projection by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2688
- Implement `logout` command by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2639

### Fixed Bugs 🐛

- Fix reporting of updates availability by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2653
- Fix `.deb` and `.rpm` package versions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2725
- Allow query to be type null in `distinct` command by @b1ron in https://github.com/FerretDB/FerretDB/pull/2658
- Fix path collisions for multiple update operators by @noisersup in https://github.com/FerretDB/FerretDB/pull/2713

### Enhancements 🛠

- Fix `_id` formatting in update error messages by @noisersup in https://github.com/FerretDB/FerretDB/pull/2711

### Documentation 📄

- Add release blog post for FerretDB version 1.2.0 by @Fashander in https://github.com/FerretDB/FerretDB/pull/2686
- Update `$project` in Supported Commands by @Fashander in https://github.com/FerretDB/FerretDB/pull/2710
- Add formatter for markdown tables by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2693
- Reformat and lint more documentation files by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2740
- Document aggregation operations by @Fashander in https://github.com/FerretDB/FerretDB/pull/2672
- Improve authentication documentation by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2737

### Other Changes 🤖

- Refactor `gitBinaryMaskParam` function by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2634
- Add `distinct` command errors test by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2687
- Clarify what's left in handling OP_MSG checksum by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2677
- Return a better error for authentication problems by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2703
- Aggregation operators refactor by @noisersup in https://github.com/FerretDB/FerretDB/pull/2664
- Implement `envtool version` command by @jeremyphua in https://github.com/FerretDB/FerretDB/pull/2714
- Make `go test -list=.` work by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2718
- Include Hana in integration tests by @polyal in https://github.com/FerretDB/FerretDB/pull/2715
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2702
- Add `logout` test for all backend by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2726
- Fix telemetry reporter logging by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2707
- Add supported aggregations to the `buildInfo` output by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2716
- Add aggregation operator tests by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2724
- Add more consistency to table tests' field names by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2717
- Don't use `sjson.GetTypeOfValue` where it shouldn't be used by @noisersup in https://github.com/FerretDB/FerretDB/pull/2728
- Unify test file names by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2709
- Make `testFindAndModifyCompat` work with `compatTestCaseResultType` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2739
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2745
- Call `ListSpecifications` driver's method in tests to check indexes by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2746
- Simplify `CountIterator` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2759
- Check for `nil` values in iterators explicitly by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2758
- Trigger GC to run finalizers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2771
- Update `golangci-lint` config by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2772
- Remove the need to call `DeepCopy` in some places by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2774
- Clean-up `lazyerrors`, use them in more places by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2770
- Replace document slices with iterators by @noisersup in https://github.com/FerretDB/FerretDB/pull/2730
- Fix `findAndModify` tests for MongoDB 6.0.6 by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2779
- Implement a few command stubs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2777
- Add more handler tests by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2769
- Remove `findAndModify` integration tests with `$` prefixed key for MongoDB 6.0.6 compatibility by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2785

### New Contributors

- @jeremyphua made their first contribution in https://github.com/FerretDB/FerretDB/pull/2714

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/42?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.2.1...v1.3.0).

## [v1.2.1](https://github.com/FerretDB/FerretDB/releases/tag/v1.2.1) (2023-05-24)

### Fixed Bugs 🐛

- Fix reporting of updates availability by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2653

### Other Changes 🤖

- Return a better error for authentication problems by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2703

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/43?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.2.0...v1.2.1).

## [v1.2.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.2.0) (2023-05-22)

### What's Changed

This release includes a highly experimental and unsupported SQLite backend.
It will be improved in future releases.

### Fixed Bugs 🐛

- Fix compatibility with C# driver by @b1ron in https://github.com/FerretDB/FerretDB/pull/2613
- Fix a bug with unset field sorting by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2638
- Return `int64` values for `dbStats` and `collStats` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2642
- Return command error from `findAndModify` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2646
- Fix index creation on nested fields by @wqhhust in https://github.com/FerretDB/FerretDB/pull/2637

### Enhancements 🛠

- Perform `insertMany` in a single transaction by @raeidish in https://github.com/FerretDB/FerretDB/pull/2532
- Relax PostgreSQL connection checks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2602
- Cleanup `insert` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/2609
- Support dot notation in projection by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2536

### Documentation 📄

- Add FerretDB v1.1.0 release blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/2594
- Update blog post image for 1.1.0 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2601
- Add documentation for `.rpm` packages by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2604
- Fix a typo in a blog post by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2611
- Fix typo on RPM package file name by @christiano in https://github.com/FerretDB/FerretDB/pull/2628
- Update documentation formatting by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2640
- Add blog post on "Meteor and FerretDB" by @Fashander in https://github.com/FerretDB/FerretDB/pull/2654

### Other Changes 🤖

- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2592
- Remove `TODO` comment for closed issue by @adetunjii in https://github.com/FerretDB/FerretDB/pull/2573
- Add experimental integration test flag for pushdown sorting by @noisersup in https://github.com/FerretDB/FerretDB/pull/2595
- Extract handler parameters from corresponding structure by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2513
- Add `shell` subcommands (`mkdir`, `rmdir`) in `envtool` by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2596
- Add basic postcondition checker for errors by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2607
- Add `sqlite` handler stub by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2608
- Make protocol-level crashes easier to understand by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2610
- Simplify `envtool shell` subcommands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2614
- Cleanup old Docker images by @wqhhust in https://github.com/FerretDB/FerretDB/pull/2533
- Fix exponential backoff minimum duration by @noisersup in https://github.com/FerretDB/FerretDB/pull/2578
- Fix `count`'s `query` parameter by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2622
- Add a README.md file for assertions by @b1ron in https://github.com/FerretDB/FerretDB/pull/2569
- Use `ExtractParameters` in handlers by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2620
- Verify `OP_MSG` message checksum by @adetunjii in https://github.com/FerretDB/FerretDB/pull/2540
- Separate codebase for aggregation `$project` and query `projection` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2631
- Implement `envtool shell read` subcommand by @wqhhust in https://github.com/FerretDB/FerretDB/pull/2626
- Cleanup projection by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2641
- Add common backend interface prototype by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2619
- Add SQLite handler flags by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2651
- Add tests for aggregation expressions with dots in `$group` aggregation stage by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2636
- Implement some SQLite backend commands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2655
- Fix tests to assert correct error by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2546
- Aggregation expression refactor by @noisersup in https://github.com/FerretDB/FerretDB/pull/2644
- Move common commands to `commoncommands` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2660
- Add basic observability into backend interfaces by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2661
- Implement metadata storage by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2656
- Add `Query` to the common backend interface by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2662
- Implement query request for SQLite backend by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2665
- Add test case for read in envtools by @wqhhust in https://github.com/FerretDB/FerretDB/pull/2657
- Run integration tests for `sqlite` handler by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2666
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2671
- Create SQLite directory if needed by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2673
- Implement SQLite `update` and `delete` commands by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2674

### New Contributors

- @adetunjii made their first contribution in https://github.com/FerretDB/FerretDB/pull/2573
- @christiano made their first contribution in https://github.com/FerretDB/FerretDB/pull/2628

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/41?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.1.0...v1.2.0).

## [v1.1.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.1.0) (2023-05-09)

### New Features 🎉

- Implement projection fields assignment by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2484
- Implement `$project` pipeline aggregation stage by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2383
- Handle `create` and `drop` commands in Hana handler by @polyal in https://github.com/FerretDB/FerretDB/pull/2458
- Implement `renameCollection` command by @b1ron in https://github.com/FerretDB/FerretDB/pull/2343

### Fixed Bugs 🐛

- Fix `findAndModify` for `$exists` query operator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2385
- Fix `SchemaStats` to return correct data by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2426
- Fix `findAndModify` for `$set` operator setting `_id` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2507
- Fix `update` for conflicting dot notation paths by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2521
- Fix `$` path errors for sort by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2534
- Fix empty projections panic by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2562
- Fix `runCommand`'s inserts of documents without `_id`s by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2574

### Enhancements 🛠

- Validate `scale` param for `dbStats` and `collStats` correctly by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2418
- Allow database name contain uppercase characters by @syasyayas in https://github.com/FerretDB/FerretDB/pull/2504
- Add identifying Arch Linux version in `hostInfo` command by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2525
- Handle absent `os-release` file by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2541
- Improve handling of `os-release` files by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2553

### Documentation 📄

- Document test script by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2353
- Use `draft` instead of `unlisted` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2372
- Make example docker compose file restart on failure by @noisersup in https://github.com/FerretDB/FerretDB/pull/2376
- Document how to get logs by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2355
- Update writing guide by @Fashander in https://github.com/FerretDB/FerretDB/pull/2373
- Add comments to our documentation workflow by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2390
- Add blogpost: Announcing FerretDB 1.0 GA - a truly Open Source MongoDB alternative by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2397
- Update documentation for index options by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2417
- Add query pushdown documentation by @noisersup in https://github.com/FerretDB/FerretDB/pull/2339
- Update README.md to link to SSPL by @cooljeanius in https://github.com/FerretDB/FerretDB/pull/2420
- Improve documentation for Docker by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2396
- Add more detailed PR guides in CONTRIBUTING.md by @AuruTus in https://github.com/FerretDB/FerretDB/pull/2435
- Remove a few double spaces by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2431
- Add image for a future blog post by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2453
- Add blogpost - Using FerretDB with Studio 3T by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2454
- Fix YAML indentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2455
- Update blog post on Using FerretDB with Studio3T by @Fashander in https://github.com/FerretDB/FerretDB/pull/2457
- Document `createIndexes`, `listIndexes`, and `dropIndexes` commands by @Fashander in https://github.com/FerretDB/FerretDB/pull/2488

### Other Changes 🤖

- Allow setting "package" variable with a testing flag by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2357
- Make it easier to use Docker-related Task targets by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2358
- Do not mark released binaries as dirty by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2371
- Make Docker Compose flags compatible by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2377
- Bump dependencies by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2367
- Fix version.txt generation for git tags by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2388
- Fix types order linter by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2391
- Cleanup deprecated errors by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2411
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2408
- Use parallel tests consistently by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2409
- Compress CI artifacts by @noisersup in https://github.com/FerretDB/FerretDB/pull/2424
- Use exponential backoff with jitter by @j0holo in https://github.com/FerretDB/FerretDB/pull/2419
- Add Mergify rules for blog posts by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2434
- Migrate to `pgx/v5` by @craigpastro in https://github.com/FerretDB/FerretDB/pull/2439
- Make it harder to misuse iterators by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2428
- Update PR template by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2441
- Rename testing flag by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2437
- Fix compilation on riscv64 by @afiskon in https://github.com/FerretDB/FerretDB/pull/2456
- Cleanup exponential backoff with jitter by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2443
- Add workaround for CockroachDB issue by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2464
- Implement blog posts previews by @noisersup in https://github.com/FerretDB/FerretDB/pull/2433
- Introduce integration benchmarks by @noisersup in https://github.com/FerretDB/FerretDB/pull/2381
- Add tests to findAndModify on `$exists` operator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2422
- Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2479
- Refactor aggregation by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2463
- Tweak documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2452
- Fix query projection for top level fields by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2386
- Handle `envtool` panic on timeout by @syasyayas in https://github.com/FerretDB/FerretDB/pull/2499
- Enable debugging tracing of SQL queries by @craigpastro in https://github.com/FerretDB/FerretDB/pull/2467
- Update blog file names to match with slug by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2497
- Add benchmark for replacing large document by @noisersup in https://github.com/FerretDB/FerretDB/pull/2482
- Add more documentation-related items to definition of done by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2494
- Return unsupported operator error for `$` projection operator by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2512
- Use `update_available` from Beacon by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2496
- Use iterator in aggregation stages by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2480
- Increase timeout for tests by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2508
- Add `InsertMany` benchmark by @raeidish in https://github.com/FerretDB/FerretDB/pull/2518
- Add coveralls.io integration by @noisersup in https://github.com/FerretDB/FerretDB/pull/2483
- Add linter for checking blog posts by @raeidish in https://github.com/FerretDB/FerretDB/pull/2459
- Add a YAML formatter by @wqhhust in https://github.com/FerretDB/FerretDB/pull/2485
- Fix `collStats` for Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2520
- Small addition to YAML formatter usage by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2524
- Cleanup of blog post linter for slug by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2526
- Pushdown simplest sorting for `find` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/2506
- Move `handlers/pg/pjson` to `handlers/sjson` by @craigpastro in https://github.com/FerretDB/FerretDB/pull/2531
- Check test database name length in compat test setup by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2527
- Document `not ready` issues label by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2544
- Remove version and name assertions in integration tests by @raeidish in https://github.com/FerretDB/FerretDB/pull/2552
- Add helpers for iterators and generators by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2542
- Do various small cleanups by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2561
- Pushdown simplest sorting for `aggregate` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/2530
- Move handlers parameters to common by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2529
- Use our own Prettier Docker image by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2535
- Improve fuzzing with recorded seed data by @fenogentov in https://github.com/FerretDB/FerretDB/pull/2392
- Add proper CLI to `envtool` - `envtool setup` subcommand by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2570
- Recover from more errors, close connection less often by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2564
- Tweak issue templates and contributing docs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2572
- Refactor integration benchmarks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2537
- Do panic in integration tests if connection can't be established by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2577
- Small refactoring by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2575
- Merge `no ci` label into `not ready` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2580

### New Contributors

- @cooljeanius made their first contribution in https://github.com/FerretDB/FerretDB/pull/2420
- @j0holo made their first contribution in https://github.com/FerretDB/FerretDB/pull/2419
- @AuruTus made their first contribution in https://github.com/FerretDB/FerretDB/pull/2435
- @craigpastro made their first contribution in https://github.com/FerretDB/FerretDB/pull/2439
- @afiskon made their first contribution in https://github.com/FerretDB/FerretDB/pull/2456
- @syasyayas made their first contribution in https://github.com/FerretDB/FerretDB/pull/2499
- @raeidish made their first contribution in https://github.com/FerretDB/FerretDB/pull/2518
- @polyal made their first contribution in https://github.com/FerretDB/FerretDB/pull/2458
- @wqhhust made their first contribution in https://github.com/FerretDB/FerretDB/pull/2485

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/40?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.0.0...v1.1.0).

## [v1.0.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.0.0) (2023-04-03)

### What's Changed

We are delighted to announce the release of FerretDB 1.0 GA!

### New Features 🎉

- Support `$sum` accumulator of `$group` aggregation by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2292
- Implement `createIndexes` command by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2244
- Add basic `getMore` command by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2309
- Implement `dropIndexes` command by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2313
- Implement `$limit` aggregation pipeline stage by @noisersup in https://github.com/FerretDB/FerretDB/pull/2270
- Add partial support for `collStats`, `dbStats` and `dataSize` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2322
- Implement `$skip` aggregation pipeline stage by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2310
- Implement `$unwind` aggregation pipeline stage by @noisersup in https://github.com/FerretDB/FerretDB/pull/2294
- Support `count` and `storageStats` fields in `$collStats` aggregation pipeline stage by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2338

### Fixed Bugs 🐛

- Fix dot notation negative index errors by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2246
- Apply `skip` before `limit` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2351

### Documentation 📄

- Update supported command for `$sum` aggregation operator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2318
- Add supported shells and GUIs images by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2323
- Publish FerretDB v0.9.4 blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/2268
- Use dashes instead of underscores or spaces by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2329
- Update documentation sidebar by @Fashander in https://github.com/FerretDB/FerretDB/pull/2347
- Update FerretDB descriptions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2281
- Improve flags documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2331
- Describe supported fields for `$collStats` aggregation stage by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2352

### Other Changes 🤖

- Use iterators for `sort`, `limit`, `skip`, and `projection` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2254
- Bump dependencies by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2307
- Improve resource tracking by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2319
- Add tests for `find`'s and `count`'s `skip` argument by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2325
- Close iterator properly by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2333
- Improve large numbers initialization in test data by @noisersup in https://github.com/FerretDB/FerretDB/pull/2324
- Ignore `unique` index option for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2350

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/39?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.9.4...v1.0.0).

## Older Releases

See <https://github.com/FerretDB/FerretDB/blob/v1.0.0/CHANGELOG.md>.
