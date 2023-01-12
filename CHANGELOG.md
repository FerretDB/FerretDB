# Changelog

## [v0.8.0](https://github.com/FerretDB/FerretDB/releases/tag/v0.8.0) (2023-01-02)

### What's Changed

We are pleased to announce our first Beta release!

#### Storage changes for PostgreSQL

We made a few _backward-incompatible_ changes in the way we store data in PostgreSQL to improve FerretDB performance.
In the future, those changes will allow us to use indexes and query collections faster.

To keep your data:

* backup FerretDB databases using `mongodump` or `mongoexport`;
* backup PostgreSQL database using `pg_dump` or other tool (just in case);
* stop FerretDB;
* drop PostgreSQL views for FerretDB databases;
* start FerretDB 0.8;
* restore databases using `mongorestore` or `mongoimport`.

#### Authentication

It is now possible to use the backend's authentication mechanisms in FerretDB.
See [documentation](https://docs.ferretdb.io/security/).

### New Features 🎉
* Support `$min` field update operator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1652
* Support `ordered` argument for `insert` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/1673
* Implement authentication for PostgreSQL by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1725

### Fixed Bugs 🐛
* Fix unset document being updated by invalid value of `$inc` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1685

### Enhancements 🛠
* Update building documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1713

### Documentation 📄
* Add section for comparison and logical query operators by @Fashander in https://github.com/FerretDB/FerretDB/pull/1647
* Add documentation for element query operators by @Fashander in https://github.com/FerretDB/FerretDB/pull/1675
* Add documentation for array query operator by @Fashander in https://github.com/FerretDB/FerretDB/pull/1695
* Enable blog post section by @Fashander in https://github.com/FerretDB/FerretDB/pull/1700

### Other Changes 🤖
* Simplify release procedure by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1657
* Modify `pjson` format by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1620
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1664
* Remove leading space from `SELECT` queries by @noisersup in https://github.com/FerretDB/FerretDB/pull/1665
* Add `InTransactionRetry` helper by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1670
* Add `mongo` test script example by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1600
* Use faster runner instances by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1678
* Update issue templates by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1671
* Bump Tigris version by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1680
* Move update tests to compat tests by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1659
* Add TODO comments by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1687
* Add `saslStart` stub by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1649
* Improve the way of storing data about collections by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1650
* Implement `iterator.Interface` for `types.Document` and `types.Array` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1683
* Improve issue template by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1692
* Remove `$elemMatch` and `$slice` projection operators by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1698
* Add `currentOp` stub by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1708
* Add basic benchmark for query pushdowns by @noisersup in https://github.com/FerretDB/FerretDB/pull/1689
* Enable authentication in PostgreSQL by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1716
* Fix Docker build by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1715
* Minor refactorings of iterators by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1718
* Test `ordered` argument validation by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1719
* Add stub for getting client-specific connection by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1723
* Add compat tests for `InsertOne` in addition to `InsertMany` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1726
* Run `govulncheck` on CI by @noisersup in https://github.com/FerretDB/FerretDB/pull/1729

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/28?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.7.1...v0.8.0).


## [v0.7.1](https://github.com/FerretDB/FerretDB/releases/tag/v0.7.1) (2022-12-19)

### New Features 🎉
* Add basic TLS support by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1586
* Add `validate` command stub by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1645

### Fixed Bugs 🐛
* Fix parsing of `OP_MSG` packets with multiple sections by @b1ron in https://github.com/FerretDB/FerretDB/pull/1611
* Fix parsing of `OP_MSG` packets with multiple sections by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1633
* Fix comparison with unset fields by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1634

### Enhancements 🛠
* Compare documents by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1597

### Documentation 📄
* Infinity values are not allowed in documents by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1622

### Other Changes 🤖
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1609
* Update release checklist by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1621
* Compare unit tests for edge cases by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1624
* Bump Go and other deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1629
* Refactor integration tests setup functions by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1625
* Fix `.deb`/`.rpm` package testing by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1631
* Bump `golang.org/x/net` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1640
* Introduce schema for `pjson` format by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1635
* Use TLS for MongoDB in integration tests by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1623
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1644
* Bump Tigris deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1651
* Remove Incomparable by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1646

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/27?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.7.0...v0.7.1).


## [v0.7.0](https://github.com/FerretDB/FerretDB/releases/tag/v0.7.0) (2022-12-05)

### New Features 🎉
* Add `msg_explain` implementation for Tigris by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1574
* Add `filter` support to `listCollections` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1567

### Fixed Bugs 🐛
* Fix parallel collection inserts for PostgreSQL by @noisersup in https://github.com/FerretDB/FerretDB/pull/1513
* Fix validation for documents with duplicate keys by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1602
* Fix greater and less operators on array value comparison by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1585

### Enhancements 🛠
* Downgrade min wire protocol version to 13 / 5.0 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1571
* Make default telemetry state a bit more clear by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1561
* Add documents validation to `wire` package by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1401
* Allow `-` in database names by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1582
* Support more default `find` parameters by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1588

### Documentation 📄
* Add authentication and role management commands to the docs by @b1ron in https://github.com/FerretDB/FerretDB/pull/1527
* Add section for diagnostic command `buildInfo` and `collStats` by @Fashander in https://github.com/FerretDB/FerretDB/pull/1480
* Add session and free monitoring commands to the docs by @b1ron in https://github.com/FerretDB/FerretDB/pull/1546
* Add Mastodon links by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1555
* Add glossary section to documentation by @Fashander in https://github.com/FerretDB/FerretDB/pull/1583

### Other Changes 🤖
* Simplify array comparison, remove `[]CompareResult` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1499
* Disable CockroachDB on GitHub Actions by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1524
* Add placeholder for testing scripts by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1525
* Add path to record loading error message by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1543
* Move array query tests to compat tests by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1526
* Move `_id` field check to separate function by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1544
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1562
* Implement `listIndexes` command stub by @noisersup in https://github.com/FerretDB/FerretDB/pull/1565
* Do not record partial files by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1566
* Sync label descriptions and issue templates by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1573
* Auto-merge is required now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1563
* Move more query tests to compat and change filter bad value check order by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1560
* Move more tests to compat query tests by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1570
* Fix test names to make them more consistent by @noisersup in https://github.com/FerretDB/FerretDB/pull/1575
* Update documentation for types, add aliases by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1580
* Simplify documents fetching for the `count` operator by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1487
* Simplify documents fetching for `delete` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1584
* Fix `listIndexes` stub by @noisersup in https://github.com/FerretDB/FerretDB/pull/1591
* Move query tests to compat by comparing the error code instead of the error message by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1579
* Add TODOs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1595
* Simplify documents fetching for `find`, `findAndModify`, and `update` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1538
* Improve `iterator.Interface` implementation in `pgdb.queryIterator` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1592
* Refactor `pg`'s `MsgListDatabases` and `pgdb` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1596
* Fix `Conform PR` workflow by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1599
* Improve `envtool` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1598
* Replace `docker-compose` with `docker compose` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1606
* Add test certificates by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1612

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/26?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.6.2...v0.7.0).


## [v0.6.2](https://github.com/FerretDB/FerretDB/releases/tag/v0.6.2) (2022-11-21)

### New Features 🎉
* Provide builds for `linux/arm/v7` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1377
* Implement `enableFreeMonitoring`, `disableFreeMonitoring` and `getFreeMonitoringStatus` commands by @noisersup in https://github.com/FerretDB/FerretDB/pull/1380

### Fixed Bugs 🐛
* Fix `SchemaStats` to return empty stats by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1359
* Fix issues for the Unix listener by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1397

### Enhancements 🛠
* Tweak supported wire protocol versions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1261
* Add more `NewCommandErrorMsgWithArgument` calls by @noisersup in https://github.com/FerretDB/FerretDB/pull/1358
* Use environment variables for configuration by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1383
* Add missing locks when update settings table by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1381

### Documentation 📄
* Make telemetry page visible in documentation sidebar by @Fashander in https://github.com/FerretDB/FerretDB/pull/1393
* Add documentation for dot notation in arrays, objects, and embedded documents by @Fashander in https://github.com/FerretDB/FerretDB/pull/1382
* Start supported commands table by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1406
* Add aggregation section (collection stages) to the docs by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1411
* Add query and write operation commands to the docs by @b1ron in https://github.com/FerretDB/FerretDB/pull/1409
* Add aggregation section (database stages, operators) to the docs by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1448
* Use colored emoji by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1485
* Add update operators to the docs by @b1ron in https://github.com/FerretDB/FerretDB/pull/1481
* Reorganize a list of supported commands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1490
* Add changelog draft by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1479
* Add user management commands to the docs by @b1ron in https://github.com/FerretDB/FerretDB/pull/1489
* Fix information about `delete`'s `comment` argument inside "supported commands" reference by @noisersup in https://github.com/FerretDB/FerretDB/pull/1498
* Add query plan cache commands to the docs by @b1ron in https://github.com/FerretDB/FerretDB/pull/1501
* Add documentation for embedded/nested documents query by @Fashander in https://github.com/FerretDB/FerretDB/pull/1478

### Other Changes 🤖
* Do not cancel in-progress CI runs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1378
* Sync and update golangci-lint configurations by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1205
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1400
* Restructure text, add unestimated tasks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1374
* Ignore website problems for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1404
* Use lowercase directory names by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1408
* Minor telemetry cleanup by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1446
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1477
* Allow duplicates in `bson` documents by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1391
* Implement `debugError` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/1402
* Update some TODOs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1452
* Record integration tests connections by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1482
* Enable more `textlint` rules by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1484
* Use `primitive.Regex` to test that regex `_id` is not allowed by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1486
* Bump Tigris version by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1379
* Use `-` in test collection names by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1325

## New Contributors
* @b1ron made their first contribution in https://github.com/FerretDB/FerretDB/pull/1409

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/25?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.6.1...v0.6.2).


## [v0.6.1](https://github.com/FerretDB/FerretDB/releases/tag/v0.6.1) (2022-11-07)

### Enhancements 🛠
* Deprecate dotted fields in data documents by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1313
* Forbid regex types and arrays in document's `_id` field by @noisersup in https://github.com/FerretDB/FerretDB/pull/1326
* Make users know about telemetry via `startupWarnings` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1336
* Deprecate nested arrays by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1334

### Documentation 📄
* Fix syntax highlighting by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1340
* Add text for empty pages by @Fashander in https://github.com/FerretDB/FerretDB/pull/1338
* Update dark and light theme logo for documentation by @Fashander in https://github.com/FerretDB/FerretDB/pull/1368
* Add documentation for configuring telemetry service by @Fashander in https://github.com/FerretDB/FerretDB/pull/1342

### Other Changes 🤖
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1337
* Switch from `markdownlint-cli` to `markdownlint-cli2` by @codingmickey in https://github.com/FerretDB/FerretDB/pull/1319
* A minor clarification about diff tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1339
* Add stress tests for `SchemaStats` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1318
* Do not compare error strings by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1344
* Add stress tests for settings table and fix simple issues with transactions by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1316
* Cleanup compat tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1345
* Fix ignore patterns for tools by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1349
* Use pre-built textlint image by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1366
* Use pre-built Docusaurus image by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1365
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1367
* Add "_id" string to linter exceptions by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1364
* Remove extra `nolint` directives by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1348
* Setup `lfs-warning` GitHub Action check by @ndkhangvl in https://github.com/FerretDB/FerretDB/pull/1371
* Bump Tigris by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1372
* Remove unsupported NaN and Inf from pjson package documentation by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1373

## New Contributors
* @codingmickey made their first contribution in https://github.com/FerretDB/FerretDB/pull/1319
* @ndkhangvl made their first contribution in https://github.com/FerretDB/FerretDB/pull/1371

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/24?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.6.0...v0.6.1).


## [v0.6.0](https://github.com/FerretDB/FerretDB/releases/tag/v0.6.0) (2022-10-27)

### What's Changed
We are pleased to announce our first Alpha release!

### New Features 🎉
* Support `$max` field update operator by @noisersup in https://github.com/FerretDB/FerretDB/pull/1124
* Migrate FerretDB to Kong by @noisersup in https://github.com/FerretDB/FerretDB/pull/1184
* Make embedded FerretDB's address configurable by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1199
* `tjson`: Support `null` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1005
* Add simple query pushdown for PostgreSQL by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1207
* Run tests on CockroachDB by @noisersup in https://github.com/FerretDB/FerretDB/pull/1260
* Add support for Unix domain sockets by @zhiburt in https://github.com/FerretDB/FerretDB/pull/1214
* Add basic telemetry by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1299
* Deprecate infinity values in data documents by @noisersup in https://github.com/FerretDB/FerretDB/pull/1296
* Explicitly disallow duplicate keys in data documents by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1293

### Fixed Bugs 🐛
* Allow empty document field names by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1196
* Fix test helpers for the `nil` case by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1241
* Fix error messages for invalid `$and`/`$or`/`$nor` arguments by @ronaudinho in https://github.com/FerretDB/FerretDB/pull/1234
* Fix `explain` command by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1294
* Fix `tjson` schema unmarshalling by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1304

### Enhancements 🛠
* Add support for Tigris auth parameters by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1177
* Use single transaction per `msg_insert` request by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1213
* Improve `buildInfo` and `serverStatus` commands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1197
* Add UUID to log messages by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1208
* Add update operators data document fields order test by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1238
* Add UUID to Prometheus metrics if requested by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1240
* Add simplest validation to check data documents by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1246
* Add “metrics” section to `serverStatus` response by @noisersup in https://github.com/FerretDB/FerretDB/pull/1231
* Call data document validation when insert or update documents in Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1290
* Add support for empty command documents by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1277
* Make `_id` field required in data documents by @noisersup in https://github.com/FerretDB/FerretDB/pull/1278
* Add more ways to disable telemetry by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1311
* Allow dashes (`-`) in collection names by @noisersup in https://github.com/FerretDB/FerretDB/pull/1312
* Collect command metrics in telemetry by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1327
* Include info about unimplemented arguments by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1330

### Documentation 📄
* Update introduction documentation by @Fashander in https://github.com/FerretDB/FerretDB/pull/1174
* Add local search plugin by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1178
* Setup documentation search by @Fashander in https://github.com/FerretDB/FerretDB/pull/1180
* DRY known differences documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1181
* Documentation website tweaks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1183
* Documentation for contributors by @Fashander in https://github.com/FerretDB/FerretDB/pull/1194
* Add "CRUD operations" and "Understanding FerretDB" sections by @Fashander in https://github.com/FerretDB/FerretDB/pull/1232
* Add documentation for the `.deb` package usage by @Fashander in https://github.com/FerretDB/FerretDB/pull/1267

### Other Changes 🤖
* Use transactions in more `pgdb` functions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1157
* Add `task` targets for offline work by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1171
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1175
* Fuzz `wire` package with recorded data by @noisersup in https://github.com/FerretDB/FerretDB/pull/1168
* Fix fluky test, refactor it by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1185
* Simplify / unify similar cases by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1187
* Setup Tigris test cases with explicit schemas by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1167
* Migrate envtool to Kong by @noisersup in https://github.com/FerretDB/FerretDB/pull/1190
* Replace `pgxtype.Querier` with `pgx.Tx` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1188
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1195
* Cleanup `pgdb` SQL statements by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1193
* Run linters on integration tests folder by @ravilushqa in https://github.com/FerretDB/FerretDB/pull/1200
* Use codecov upload token by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1204
* Add security scan by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1142
* Use single transaction per `msg_update` request by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1212
* Use Go 1.19.2 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1211
* Fix running `pg` and `tigris` tests in parallel by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1218
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1230
* Use single transaction per `msg_findandmodify` request by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1217
* Improve `task env-data` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1220
* Split `fjson` into `pjson` and `types/fjson` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1219
* Use single transaction for `listDatabases` command by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1237
* Cleanup old validation by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1179
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1245
* Update internal process docs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1249
* Fix flag name by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1255
* Fix CLI flags for Tigris by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1256
* Remove forked `golangci-lint` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1258
* Cleanup types/fjson package by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1254
* Minor handlers refactoring by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1264
* `fjson` and fuzzing cleanup by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1262
* Skip `pjson` fuzzing of invalid documents for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1274
* Add schema-related test cases to `tjson` package by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1247
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1275
* Update docs for the `dummy` handler by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1276
* Fix documentation for linking PRs and issues by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1268
* Add experimental mergify configuration by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1281
* Improve tests cleanup by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1287
* Remove implicit mergify rules by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1288
* Run CockroachDB tests on CI by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1289
* Bump cockroachdb/cockroach from v22.1.8 to v22.1.9 in /build/deps by @dependabot in https://github.com/FerretDB/FerretDB/pull/1285
* Migrate to a newer Tigris version and fix relevant tests by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1239
* Add ability to subscribe to state changes by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1265
* Move tjson into internal/handlers/tigris/tjson by @chilagrow in https://github.com/FerretDB/FerretDB/pull/1291
* Fix a typo in the `types` package docs by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1297
* Disallow usage of old context package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1292
* Disable Unix sockets in tests for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1298
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1309
* Expand `debugError` stub by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1303
* Add comment about `diff` tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1302
* Refactor handler errors by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1322

## New Contributors
* @chilagrow made their first contribution in https://github.com/FerretDB/FerretDB/pull/1254
* @ronaudinho made their first contribution in https://github.com/FerretDB/FerretDB/pull/1234
* @zhiburt made their first contribution in https://github.com/FerretDB/FerretDB/pull/1214

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/22?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.5.4...v0.6.0).


## [v0.5.4](https://github.com/FerretDB/FerretDB/releases/tag/v0.5.4) (2022-09-22)

### Fixed Bugs 🐛
* Add missing `$k` to the schema when creating collection in Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1136

### Documentation 📄
* Remove docusaurus references and update documentation by @Fashander in https://github.com/FerretDB/FerretDB/pull/1130
* Deploy documentation PRs to Vercel by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1131

### Other Changes 🤖
* Add transaction to `msg_drop` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1129
* Add transaction to `pg`'s `msg_listcollections` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1135
* Fix tests for Tigris by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1134
* Use fixed `-test-record` directory in Task targets by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1139
* Fix a typo in Readme by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1141
* Use transaction in more `pgdb` functions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1143
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1158
* Use transaction in more `pgdb` functions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1144
* Refactor `msg_delete` handlers by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1152
* Improve contributing guidelines by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1146
* Update process documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1153
* Update issues and PR templates by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1155
* Fix typo by @si3nloong in https://github.com/FerretDB/FerretDB/pull/1165
* Migrate fuzztool to Kong by @noisersup in https://github.com/FerretDB/FerretDB/pull/1159

## New Contributors
* @si3nloong made their first contribution in https://github.com/FerretDB/FerretDB/pull/1165

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/23?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.5.3...v0.5.4).


## [v0.5.3](https://github.com/FerretDB/FerretDB/releases/tag/v0.5.3) (2022-09-08)

### New Features 🎉
* Add support for updates with replacement objects by @fcoury in https://github.com/FerretDB/FerretDB/pull/791
* Add support for `$update`'s `$set` and `$setOnInsert` operators dot notation by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1008
* Support `$pop` array update operator by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1020
* Add support for `$update`'s `$unset` operators dot notation by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1028
* `tjson`: Implement `regex` by @noisersup in https://github.com/FerretDB/FerretDB/pull/1050
* Implement `MsgDataSize` for Tigris by @polldo in https://github.com/FerretDB/FerretDB/pull/1060
* Support `ordered` argument for `delete` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/1004
* Implement simple query pushdown for Tigris by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1091
* Implement `MsgFindAndModify` for Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1065
* implement `timestamp` type for tigris by @noisersup in https://github.com/FerretDB/FerretDB/pull/1117

### Fixed Bugs 🐛
* Improve `TestCommandsAdministrationServerStatus` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1062
* Fix `ModifiedCount` for updates with an empty replacement document by @nicolascb in https://github.com/FerretDB/FerretDB/pull/1067
* Fix `$inc` `update` operator int64-max issue by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1071
* Handle `findAndModify` and `update` correctly when collection doesn't exist by @noisersup in https://github.com/FerretDB/FerretDB/pull/1087
* Require `limit` parameter in `delete` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/1066

### Enhancements 🛠
* Fix `update` operation for Tigris handler by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1041
* Collect sizes in `MsgListDatabases` for Tigris by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1043

### Documentation 📄
* Add GitHub Pages with documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1100
* Improve contribution guidelines and documentation website by @Fashander in https://github.com/FerretDB/FerretDB/pull/1114
* Fix macOS spelling by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1127

### Other Changes 🤖
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1014
* Fix `Run linters` job in Taskfile.yml by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1022
* Improve and document integration tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1021
* Add missing `MaxTimeMS` support for Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1026
* Expose `delete` problem by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1030
* Bump Tigris version to 1.0.0-alpha.27 by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1032
* Bump postgres from 14.4 to 14.5 in /build/deps by @dependabot in https://github.com/FerretDB/FerretDB/pull/1033
* `tjson`: Implement `datetime` by @noisersup in https://github.com/FerretDB/FerretDB/pull/1027
* `tjson`: Add package documentation for types mapping by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1031
* Rework database and collection creation for Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1038
* Add a few more tests for logical query operators by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1049
* Ensure that database and collection names are unique by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1046
* Implement `MsgDBStats` for Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1047
* Bump Tigris version to 1.0.0-alpha.29 by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1054
* Add tests for update with replacement by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1044
* Small `pgdb` cleanup by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1055
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1059
* Less strict delta for `dataSize` in tests by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1053
* Add `collMod` command stub by @ravilushqa in https://github.com/FerretDB/FerretDB/pull/1037
* Add a linter for Semantic Line Breaks in Markdown files by @GrandShow in https://github.com/FerretDB/FerretDB/pull/998
* Fix data race in test by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1072
* Use `CODECOV_TOKEN` if available by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1073
* Bump dependencies by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1084
* Add integration tests for logical operators by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1085
* Implement `MsgCollStats` for Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1063
* Implement `MsgCreate` for Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/1048
* Use npm lock files for tools by @folex in https://github.com/FerretDB/FerretDB/pull/1093
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1099
* Simplify/sync `delete` a bit by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1104
* Enable `errorlint` for new code by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1105
* Add missing TODO by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1108
* Migrate to MongoDB 6 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1074
* Switch to Go 1.19, bump dependencies by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1123
* Record incoming data for fuzzing by @noisersup in https://github.com/FerretDB/FerretDB/pull/1107
* Add transaction to `msg_drop_database` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/1126

## New Contributors
* @polldo made their first contribution in https://github.com/FerretDB/FerretDB/pull/1060
* @ravilushqa made their first contribution in https://github.com/FerretDB/FerretDB/pull/1037
* @GrandShow made their first contribution in https://github.com/FerretDB/FerretDB/pull/998
* @nicolascb made their first contribution in https://github.com/FerretDB/FerretDB/pull/1067
* @folex made their first contribution in https://github.com/FerretDB/FerretDB/pull/1093
* @Fashander made their first contribution in https://github.com/FerretDB/FerretDB/pull/1114

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/21?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.5.2...v0.5.3).


## [v0.5.2](https://github.com/FerretDB/FerretDB/releases/tag/v0.5.2) (2022-08-09)

### New Features 🎉
* Support `comment` and `$comment` `update`'s arguments by @noisersup in https://github.com/FerretDB/FerretDB/pull/937
* Support `multi` `update`'s argument by @fcoury in https://github.com/FerretDB/FerretDB/pull/790
* Support `comment` and `$comment` `findAndModify`'s argument by @noisersup in https://github.com/FerretDB/FerretDB/pull/958
* Support `comment` and `$comment` `delete`'s arguments by @noisersup in https://github.com/FerretDB/FerretDB/pull/954
* Support `maxTimeMS` argument for `find` and `findAndModify` methods by @DoodgeMatvey in https://github.com/FerretDB/FerretDB/pull/608
* Add support for `update`'s `$inc` operator dot notation by @w84thesun in https://github.com/FerretDB/FerretDB/pull/915

### Fixed Bugs 🐛
* Fix `nModified` count for `update`'s `$set` operator with the same value by @w84thesun in https://github.com/FerretDB/FerretDB/pull/949

### Other Changes 🤖
* `tjson`: Fix schema comparison by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/944
* Make compat error messages better by @AlekSi in https://github.com/FerretDB/FerretDB/pull/950
* Enable `wsl` linter for new and changed code by @AlekSi in https://github.com/FerretDB/FerretDB/pull/856
* Fix some collection names breaking `listDatabases` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/953
* Remove some tests to make next PRs smaller by @AlekSi in https://github.com/FerretDB/FerretDB/pull/959
* Add `SkipForTigris` helper, use it by @AlekSi in https://github.com/FerretDB/FerretDB/pull/960
* Add setup for compatibility tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/961
* Add compatibility tests for `$and` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/963
* Bump igorshubovych/markdownlint-cli from v0.32.0 to v0.32.1 in /build/deps by @dependabot in https://github.com/FerretDB/FerretDB/pull/955
* `tjson`: Check how we support `binary` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/967
* Move logic operators tests to compatibility tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/965
* Add compatibility tests for `$inc` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/964
* Add test case for an empty update path by @AlekSi in https://github.com/FerretDB/FerretDB/pull/976
* Bump golang from 1.18.4 to 1.18.5 by @dependabot in https://github.com/FerretDB/FerretDB/pull/977
* Improve Document's Path API by @w84thesun in https://github.com/FerretDB/FerretDB/pull/973
* `tjson`: Add unit tests for `ObjectID` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/971
* Make linter to enforce our preferred types order in type switch by @fenogentov in https://github.com/FerretDB/FerretDB/pull/654
* Add back `task env-data` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/983
* Insert test data in random order by @AlekSi in https://github.com/FerretDB/FerretDB/pull/862
* `tjson`: Improve ObjectID test by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/992
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/995
* Bump Tigris Docker image to alpha.26 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/997
* Tigris: simplify `ObjectID` and filter usage by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/968
* Add more scalar values to tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/984
* Implement `aggregate` command stub by @AlekSi in https://github.com/FerretDB/FerretDB/pull/981
* Reformat with Go 1.19 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1003
* `tjson`: Cover `document` (`object`) type with tests by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/957
* Add compatibility `delete` test for Tigris by @AlekSi in https://github.com/FerretDB/FerretDB/pull/1002

## New Contributors
* @fcoury made their first contribution in https://github.com/FerretDB/FerretDB/pull/790

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/20?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.5.1...v0.5.2).


## [v0.5.1](https://github.com/FerretDB/FerretDB/releases/tag/v0.5.1) (2022-07-26)

### New Features 🎉
* Validate database names by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/913
* Support `$all` array query operator by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/724
* Support `getLog` diagnostic command by @fenogentov in https://github.com/FerretDB/FerretDB/pull/711
* Implement `MsgCount` for Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/928
* Support `explain` diagnostic command by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/909

### Fixed Bugs 🐛
* Fix edge cases in `drop` and `dropDatabase` handlers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/891
* Fix `ModifyCount` for update operators by @w84thesun in https://github.com/FerretDB/FerretDB/pull/939

### Enhancements 🛠
* Support `gt` `lt` operator comparison for Array type by @ribaraka in https://github.com/FerretDB/FerretDB/pull/819
* Optimize documents fetching / filtering by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/808
* Add test for a database name border case by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/921

### Documentation 📄
* Add a tip to limit concurrent tasks by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/883

### Other Changes 🤖
* Add a few testing helpers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/874
* Add support for `no ci` label by @AlekSi in https://github.com/FerretDB/FerretDB/pull/876
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/880
* Bump golang from 1.18.3 to 1.18.4 by @dependabot in https://github.com/FerretDB/FerretDB/pull/881
* Extract two more helpers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/875
* Set pprof label for client connections by @AlekSi in https://github.com/FerretDB/FerretDB/pull/885
* Cancel request's context when request processed by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/884
* Simplify `dbStats` tests a bit, add TODOs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/886
* Disable logs during test setup by @AlekSi in https://github.com/FerretDB/FerretDB/pull/888
* Use `InsertMany` instead of `InsertOne` in tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/882
* Restart development containers faster by @AlekSi in https://github.com/FerretDB/FerretDB/pull/889
* Cover more logic in transactions by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/887
* Disconnect client in embedded tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/890
* Stop tests on the first data race by @AlekSi in https://github.com/FerretDB/FerretDB/pull/893
* Wait for `Tigris` backend to be ready by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/894
* Handle `42P07` PostgreSQL error to fix the tests by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/895
* Build .rpm and .deb packages by @fenogentov in https://github.com/FerretDB/FerretDB/pull/739
* Add setup for compatibility tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/901
* Extract parameter list into one variable in `QueryDocuments` by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/910
* Add first compatibility tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/863
* Use `v` instead of `value` in tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/916
* Tweak codecov settings by @AlekSi in https://github.com/FerretDB/FerretDB/pull/920
* Remove deprecated functions from `pgdb.Pool` by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/922
* Extract integration tests setup to own package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/923
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/927
* comment `url.Values` to prevent test failing by @noisersup in https://github.com/FerretDB/FerretDB/pull/930
* Add a comment to the setup function about database and collection creation when provider list is empty by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/929
* Bump `golangci-lint`, remove old hacks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/932
* Fix tests for `$all` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/934
* Add Path tests by @w84thesun in https://github.com/FerretDB/FerretDB/pull/936
* Build packages on CI by @AlekSi in https://github.com/FerretDB/FerretDB/pull/938
* Tweak linter settings by @AlekSi in https://github.com/FerretDB/FerretDB/pull/942
* Port and sync unit testing approach from `fjson` to `tjson` by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/935
* Ensure that update operators are in sync by @AlekSi in https://github.com/FerretDB/FerretDB/pull/946

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/19?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.5.0...v0.5.1).


## [v0.5.0](https://github.com/FerretDB/FerretDB/releases/tag/v0.5.0) (2022-07-11)

### What's Changed
This release enables the usage of FerretDB as a Go library.
See [this blog post](https://www.ferretdb.io/0-5-0-release-is-out-embedding-ferretdb-into-go-programs/).

### New Features 🎉
* Support embedded use-case by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/754
* Validate collection names by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/844

### Fixed Bugs 🐛
* Fix embedded usage by @AlekSi in https://github.com/FerretDB/FerretDB/pull/798
* Fix `whatsmyuri` command by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/796
* Handle `null` value for `nameOnly` in `listDatabases` handler by @DoodgeMatvey in https://github.com/FerretDB/FerretDB/pull/738
* pgdb: cover transactions with `inTransaction` function to simplify error handling by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/833

### Enhancements 🛠
* Support all valid collection names by @w84thesun in https://github.com/FerretDB/FerretDB/pull/778
* Remove MongoDB driver "dependency" by @AlekSi in https://github.com/FerretDB/FerretDB/pull/853

### Documentation 📄
* Update contributing docs and PR template according to our best practices by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/779
* Update contributing documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/843
* Document that NUL (`\0`) strings is not supported by @w84thesun in https://github.com/FerretDB/FerretDB/pull/865

### Other Changes 🤖
* Tweak schedule of daily builds by @AlekSi in https://github.com/FerretDB/FerretDB/pull/794
* Do not import `pg` handler explicitly by @AlekSi in https://github.com/FerretDB/FerretDB/pull/799
* Add TODO item by @AlekSi in https://github.com/FerretDB/FerretDB/pull/804
* Fix `task env-pull` target by @AlekSi in https://github.com/FerretDB/FerretDB/pull/810
* Improve contributing documentation for Windows development by @w84thesun in https://github.com/FerretDB/FerretDB/pull/795
* Fix Docker image build by @AlekSi in https://github.com/FerretDB/FerretDB/pull/805
* Make it easier to trigger rebuilds by @AlekSi in https://github.com/FerretDB/FerretDB/pull/815
* Use `github.head_ref` instead of `github.ref` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/814
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/817
* Install QEMU for building Docker images by @AlekSi in https://github.com/FerretDB/FerretDB/pull/820
* Use multi-stage docker build by @AlphaB in https://github.com/FerretDB/FerretDB/pull/605
* Unify `update` tests by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/809
* Add TODOs for all update operators by @AlekSi in https://github.com/FerretDB/FerretDB/pull/832
* Add `env-data` Taskfile target by @AlekSi in https://github.com/FerretDB/FerretDB/pull/834
* Tweak tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/837
* Export integration tests helpers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/838
* Remove old style of `+build` tags where possible by @AlekSi in https://github.com/FerretDB/FerretDB/pull/839
* Export fields by @AlekSi in https://github.com/FerretDB/FerretDB/pull/840
* Update Tigris version by @AlekSi in https://github.com/FerretDB/FerretDB/pull/841
* Create Tigris databases by @AlekSi in https://github.com/FerretDB/FerretDB/pull/842
* Add basic CI for Tigris by @AlekSi in https://github.com/FerretDB/FerretDB/pull/784
* Test with the `main` version of Tigris too by @AlekSi in https://github.com/FerretDB/FerretDB/pull/846
* Wait for Tigris to be fully up by @AlekSi in https://github.com/FerretDB/FerretDB/pull/854
* Fill MongoDB on `task env-data` too by @AlekSi in https://github.com/FerretDB/FerretDB/pull/860
* Add CI job for short tests without environment by @AlekSi in https://github.com/FerretDB/FerretDB/pull/855
* Add TODOs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/852
* Fix `task run` on Windows by @AlekSi in https://github.com/FerretDB/FerretDB/pull/867
* Fix invalid variable names by @AlekSi in https://github.com/FerretDB/FerretDB/pull/868
* Add `ferretdb_` prefix to our custom build tags by @AlekSi in https://github.com/FerretDB/FerretDB/pull/869

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/18?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.4.0...v0.5.0).


## [v0.4.0](https://github.com/FerretDB/FerretDB/releases/tag/v0.4.0) (2022-06-27)

### What's Changed
This release adds preliminary support for the [Tigris](https://www.tigrisdata.com) backend.
We plan to reach parity with our PostgreSQL backend in the next release.

### New Features 🎉
* Support `$setOnInsert` field update operator by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/644
* Support `$unset` field update operator by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/691
* Support `$currentDate` field update operator by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/662
* Support array querying by @w84thesun in https://github.com/FerretDB/FerretDB/pull/618
* Support `$elemMatch` array query operator by @w84thesun in https://github.com/FerretDB/FerretDB/pull/707
* Implement `getFreeMonitoringStatus` stub by @noisersup in https://github.com/FerretDB/FerretDB/pull/751
* Implement `setFreeMonitoring` stub by @noisersup in https://github.com/FerretDB/FerretDB/pull/759
* Implement `tigris` handler by @AlekSi in https://github.com/FerretDB/FerretDB/pull/690

### Fixed Bugs 🐛
* Handle both `buildinfo` and `buildInfo` commands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/688
* Fix a bug with proxy response logs by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/705
* Handle `find`, `count` and `delete` correctly when collection doesn't exist by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/710
* Fix default values for flags by @AlekSi in https://github.com/FerretDB/FerretDB/pull/743
* Fix embedded array query bug by @ribaraka in https://github.com/FerretDB/FerretDB/pull/736

### Enhancements 🛠
* Array comparison substitution by @ribaraka in https://github.com/FerretDB/FerretDB/pull/676
* Build `tigris` handler only if tag is present by @AlekSi in https://github.com/FerretDB/FerretDB/pull/681
* Support getParameter's showDetails, allParameters by @fenogentov in https://github.com/FerretDB/FerretDB/pull/606
* Make log level configurable by @fenogentov in https://github.com/FerretDB/FerretDB/pull/687
* `$currentDate` Timestamp fix `DateTime` seconds and milliseconds bug by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/701

### Documentation 📄
* Be explicit about MongoDB version by @AlekSi in https://github.com/FerretDB/FerretDB/pull/679
* Fix pull request template by @AlekSi in https://github.com/FerretDB/FerretDB/pull/746

### Other Changes 🤖
* Use `"` instead of `'` in all .yml files by @AlekSi in https://github.com/FerretDB/FerretDB/pull/675
* Add empty Tigris handler by @AlekSi in https://github.com/FerretDB/FerretDB/pull/671
* Do not test a global list of databases in parallel by @AlekSi in https://github.com/FerretDB/FerretDB/pull/678
* Enable `revive` linter by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/672
* More tests for dot notation support by @w84thesun in https://github.com/FerretDB/FerretDB/pull/660
* Use circular buffer for zap logs by @fenogentov in https://github.com/FerretDB/FerretDB/pull/585
* Fix build by @AlekSi in https://github.com/FerretDB/FerretDB/pull/703
* Add `tjson` package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/682
* Improve function comment by @AlekSi in https://github.com/FerretDB/FerretDB/pull/712
* Use separate encodings for ObjectID and binary by @AlekSi in https://github.com/FerretDB/FerretDB/pull/713
* Add the default Task target by @AlekSi in https://github.com/FerretDB/FerretDB/pull/716
* Add workaround for Dependabot by @AlekSi in https://github.com/FerretDB/FerretDB/pull/717
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/723
* Always install Go and skip test cache by @AlekSi in https://github.com/FerretDB/FerretDB/pull/718
* Bump mongo from 5.0.8 to 5.0.9 in /build/deps by @dependabot in https://github.com/FerretDB/FerretDB/pull/719
* Better dummy handler errors by @AlekSi in https://github.com/FerretDB/FerretDB/pull/715
* Add `task run-proxy` command by @AlekSi in https://github.com/FerretDB/FerretDB/pull/725
* Add `Min` and `Max` methods to `types.Array` by @ribaraka in https://github.com/FerretDB/FerretDB/pull/726
* Add arrays with `NaN`, `double` and nested empty array to tests' shared data by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/728
* Bump github.com/go-task/task/v3 from 3.12.1 to 3.13.0 in /tools by @dependabot in https://github.com/FerretDB/FerretDB/pull/741
* Disable "free monitoring" to simplify tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/748
* Re-enable `TestStatisticsCommands` tests by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/704
* Fix `lint-golangci-lint` task for Windows systems by @w84thesun in https://github.com/FerretDB/FerretDB/pull/752
* Remove outdated comment by @AlekSi in https://github.com/FerretDB/FerretDB/pull/755
* Skip `-race` flag on Windows by @w84thesun in https://github.com/FerretDB/FerretDB/pull/753
* Fix fluky test by @AlekSi in https://github.com/FerretDB/FerretDB/pull/757
* `tjson` improvements by @AlekSi in https://github.com/FerretDB/FerretDB/pull/760
* Unify similar code in `pg` handler by @AlekSi in https://github.com/FerretDB/FerretDB/pull/762
* Add Tigris environment by @AlekSi in https://github.com/FerretDB/FerretDB/pull/761
* Bump postgres from 14.3 to 14.4 in /build/deps by @dependabot in https://github.com/FerretDB/FerretDB/pull/768
* Use forked `golangci-lint` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/758
* Update `conform-pr` action by @AlekSi in https://github.com/FerretDB/FerretDB/pull/783
* Drop `test_db` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/788
* Add `task init-clean` target by @AlekSi in https://github.com/FerretDB/FerretDB/pull/756
* Add `godoc` to tools by @AlekSi in https://github.com/FerretDB/FerretDB/pull/789

## New Contributors
* @noisersup made their first contribution in https://github.com/FerretDB/FerretDB/pull/751

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/17?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.3.0...v0.4.0).


## [v0.3.0](https://github.com/FerretDB/FerretDB/releases/tag/v0.3.0) (2022-06-01)

### New Features 🎉
* Support `findAndModify` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/548
* Support `$inc` field `update` operator by @w84thesun in https://github.com/FerretDB/FerretDB/pull/596
* Support `$set` field update operator by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/634

### Fixed Bugs 🐛
* Improve negative zero handling by @AlekSi in https://github.com/FerretDB/FerretDB/pull/613

### Enhancements 🛠
* Added support for sorting scalar data types by @ribaraka in https://github.com/FerretDB/FerretDB/pull/607

### Other Changes 🤖
* Better `-0` handling in tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/616
* Bump github.com/golangci/golangci-lint from 1.46.1 to 1.46.2 in /tools by @dependabot in https://github.com/FerretDB/FerretDB/pull/617
* Bump PostgreSQL and MongoDB versions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/599
* Rename `OP_*` constants to `OpCode*` constants  by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/620
* Bump gopkg.in/yaml.v3 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/646
* Bump gopkg.in/yaml.v3 in tools by @AlekSi in https://github.com/FerretDB/FerretDB/pull/648
* Make `Path` type by @w84thesun in https://github.com/FerretDB/FerretDB/pull/635
* Fix incorrect test for `$mod` operator by @fenogentov in https://github.com/FerretDB/FerretDB/pull/645
* Skip test on all ARM64 OSes by @AlekSi in https://github.com/FerretDB/FerretDB/pull/652
* Add more visibility for the router/proxy error log levels by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/650
* Update CODEOWNERS by @AlekSi in https://github.com/FerretDB/FerretDB/pull/655
* Sync dummy and pg handlers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/641
* Panic on unexpected order values by @AlekSi in https://github.com/FerretDB/FerretDB/pull/668
* Add some comments to the functions and variables by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/619
* Remove dead code by @AlekSi in https://github.com/FerretDB/FerretDB/pull/669

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/15?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.2.1...v0.3.0).


## [v0.2.1](https://github.com/FerretDB/FerretDB/releases/tag/v0.2.1) (2022-05-17)

### New Features 🎉
* Support `$slice` projection query operator by @GinGin3203 in https://github.com/FerretDB/FerretDB/pull/518
* Support `$comment` query operator by @ribaraka in https://github.com/FerretDB/FerretDB/pull/563
* Support basic `connectionStatus` diagnostic command by @fenogentov in https://github.com/FerretDB/FerretDB/pull/572
* Support `$regex` evaluation query operator by @w84thesun in https://github.com/FerretDB/FerretDB/pull/588

### Enhancements 🛠
* Support querying documents by @w84thesun in https://github.com/FerretDB/FerretDB/pull/573
* Improve comparison of arrays and documents by @ribaraka in https://github.com/FerretDB/FerretDB/pull/589
* Support `getParameter`'s parameters by @fenogentov in https://github.com/FerretDB/FerretDB/pull/535
* Add stubs to make VSCode plugin work by @AlekSi in https://github.com/FerretDB/FerretDB/pull/603

### Documentation 📄
* Add conform CI workflow, improve docs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/566
* Update CONTRIBUTING.md with typo fix and a tiny correction by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/574
* Add note about forks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/575

### Other Changes 🤖
* Bump go.mongodb.org/mongo-driver from 1.9.0 to 1.9.1 in /integration by @dependabot in https://github.com/FerretDB/FerretDB/pull/555
* Add missing `//nolint` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/556
* Set the handler to use via a command-line flag and remove debug handlers from interface by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/534
* Add tests for `RemoveByPath` by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/549
* Add `altMessage` to `AssertEqualError` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/550
* Add documentation for values comparision by @AlekSi in https://github.com/FerretDB/FerretDB/pull/559
* Add `debug` and `panic` msg handlers to `Command` map by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/561
* Add `RemoveByPath` for `Array` and `CompositeTypeInterface` by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/560
* Bump docker/login-action from 1 to 2 by @dependabot in https://github.com/FerretDB/FerretDB/pull/565
* Bump pgx version by @AlekSi in https://github.com/FerretDB/FerretDB/pull/570
* Use `float64(x)` for `ok` everywhere by @AlekSi in https://github.com/FerretDB/FerretDB/pull/577
* Improve `AssertEqualAltError` documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/578
* Remove `types.MustNewDocument` in some places by @AlekSi in https://github.com/FerretDB/FerretDB/pull/579
* Remove `MustNewDocument` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/581
* Remove `MustNewArray` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/582
* Remove `MustConvertDocument` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/583
* Enable `staticcheck` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/580
* Enable `gosimple` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/584
* Change the way linters work by @AlekSi in https://github.com/FerretDB/FerretDB/pull/586
* Merge `BigNumbersData` into `Scalars` by @AlphaB in https://github.com/FerretDB/FerretDB/pull/595
* Set `GOLANGCI_LINT_CACHE` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/597
* Increase `golangci-lint` timeout by @AlekSi in https://github.com/FerretDB/FerretDB/pull/598
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/604

## New Contributors
* @rumyantseva made their first contribution in https://github.com/FerretDB/FerretDB/pull/574
* @AlphaB made their first contribution in https://github.com/FerretDB/FerretDB/pull/595

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/16?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.2.0...v0.2.1).


## [v0.2.0](https://github.com/FerretDB/FerretDB/releases/tag/v0.2.0) (2022-05-04)

### What's Changed
This release implements all required functionality to support [CLA Assistant](https://github.com/cla-assistant/cla-assistant).
More details will be available shortly in [our blog](https://www.ferretdb.io/blog/).

### New Features 🎉
* Add support for `$nin` operator by @ribaraka in https://github.com/FerretDB/FerretDB/pull/459
* Support querying with dot notation for documents by @GinGin3203 in https://github.com/FerretDB/FerretDB/pull/483
* Add support for `$ne` operator by @ribaraka in https://github.com/FerretDB/FerretDB/pull/464
* Add basic `findAndModify` implementation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/501
* Add support for `$in` operator by @ribaraka in https://github.com/FerretDB/FerretDB/pull/499

### Fixed Bugs 🐛
* Fix large numbers comparision by @DoodgeMatvey in https://github.com/FerretDB/FerretDB/pull/466
* Fix panic on receiving a filter query with unknown operator by @GinGin3203 in https://github.com/FerretDB/FerretDB/pull/517
* Fix bitwise operators by @w84thesun in https://github.com/FerretDB/FerretDB/pull/488

### Enhancements 🛠
* Return better errors for unimplemented operations by @AlekSi in https://github.com/FerretDB/FerretDB/pull/504
* Implement `nameOnly` for `listDatabases` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/524
* Improve `hostInfo` command's `os` response by @DoodgeMatvey in https://github.com/FerretDB/FerretDB/pull/509

### Documentation 📄
* Mention force pushes by @AlekSi in https://github.com/FerretDB/FerretDB/pull/500
* Update guidelines by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/496
* Document `task env-pull` target by @AlekSi in https://github.com/FerretDB/FerretDB/pull/528

### Other Changes 🤖
* Stabilize tests by always sorting results by @AlekSi in https://github.com/FerretDB/FerretDB/pull/490
* Skip one test for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/494
* Bump MongoDB version by @AlekSi in https://github.com/FerretDB/FerretDB/pull/495
* Use `goimports` to group imports on `task fmt` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/498
* Make default Docker arguments a bit more useful by @AlekSi in https://github.com/FerretDB/FerretDB/pull/502
* Export helpers that will be used in other package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/505
* Correctly override `FUZZTIME` on CI by @AlekSi in https://github.com/FerretDB/FerretDB/pull/506
* Pass context to PostgreSQL pool by @AlekSi in https://github.com/FerretDB/FerretDB/pull/507
* Bump dependencies by @AlekSi in https://github.com/FerretDB/FerretDB/pull/514
* Remove `Array.Subslice` method by @AlekSi in https://github.com/FerretDB/FerretDB/pull/515
* Remove `types.CString` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/529
* Make test helpers harder to misuse by @AlekSi in https://github.com/FerretDB/FerretDB/pull/530
* Move existing comparision code to `types` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/531
* Extract common interface for handlers by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/521
* Move all handler test to integration tests by @w84thesun in https://github.com/FerretDB/FerretDB/pull/523
* Use `nil` errors instead of empty values by @fenogentov in https://github.com/FerretDB/FerretDB/pull/542
* Delete old tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/543
* Add tests for `sort` and `find` parameters type by @w84thesun in https://github.com/FerretDB/FerretDB/pull/544

## New Contributors
* @DoodgeMatvey made their first contribution in https://github.com/FerretDB/FerretDB/pull/466

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/14?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.1.1...v0.2.0).


## [v0.1.1](https://github.com/FerretDB/FerretDB/releases/tag/v0.1.1) (2022-04-19)

### New Features 🎉
* Support `$gt` comparision operator by @ribaraka in https://github.com/FerretDB/FerretDB/pull/330
* Support `$exists` element query operator by @w84thesun in https://github.com/FerretDB/FerretDB/pull/446
* Add basic `upsert` support by @AlekSi in https://github.com/FerretDB/FerretDB/pull/473
* Add `$type` operator by @w84thesun in https://github.com/FerretDB/FerretDB/pull/453
* Support `$mod` evaluation query operator by @fenogentov in https://github.com/FerretDB/FerretDB/pull/440
* Support logical operators by @w84thesun in https://github.com/FerretDB/FerretDB/pull/465

### Enhancements 🛠
* Ping database in some commands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/435
* Ensure that `_id` fields are always the first by @AlekSi in https://github.com/FerretDB/FerretDB/pull/476

### Documentation 📄
* Improve contributing guidelines by @AlekSi in https://github.com/FerretDB/FerretDB/pull/480

### Other Changes 🤖
* Integration tests improvements by @AlekSi in https://github.com/FerretDB/FerretDB/pull/441
* Add test stub for bitwise operators by @AlekSi in https://github.com/FerretDB/FerretDB/pull/443
* Add tests for collections `create` and `drop` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/444
* Add tests for more diagnostic commands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/448
* Transfer existing comparison tests by @ribaraka in https://github.com/FerretDB/FerretDB/pull/445
* Move `getParameter` tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/450
* Improve `envtool` diagnostics by @w84thesun in https://github.com/FerretDB/FerretDB/pull/426
* Fix Postgres port check for `envtool` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/451
* Add support for `$gte`, `$lt`, `$lte` operators by @ribaraka in https://github.com/FerretDB/FerretDB/pull/452
* Add tests for `null` values by @w84thesun in https://github.com/FerretDB/FerretDB/pull/458
* Bump actions/upload-artifact from 2 to 3 by @dependabot in https://github.com/FerretDB/FerretDB/pull/460
* Update tests for the latest mongo-driver by @AlekSi in https://github.com/FerretDB/FerretDB/pull/463
* Rearrange tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/467
* Do not invoke Dance tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/468
* Minor unification of tests style by @AlekSi in https://github.com/FerretDB/FerretDB/pull/469
* Add helpers that generate ObjectID by @AlekSi in https://github.com/FerretDB/FerretDB/pull/474
* Add deep copy helpers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/475
* Allow the usage of proxy/diff mode in tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/477
* Bump codecov/codecov-action from 2 to 3 by @dependabot in https://github.com/FerretDB/FerretDB/pull/461
* Composite data type find handling by @ribaraka in https://github.com/FerretDB/FerretDB/pull/471
* Fix failing tests by @w84thesun in https://github.com/FerretDB/FerretDB/pull/482
* Rename `q` to `filter` in tests by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/484
* Supress linter warning by @AlekSi in https://github.com/FerretDB/FerretDB/pull/485

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/11?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.1.0...v0.1.1).


## [v0.1.0](https://github.com/FerretDB/FerretDB/releases/tag/v0.1.0) (2022-04-04)

### What's Changed
In this release, we made a big change in the way FerretDB fetches data from PostgreSQL.

Previously, we generated a single SQL query that extensively used json/jsonb PostgreSQL functions for each incoming MongoDB request, then converted fetched data.
All the filtering was performed by PostgreSQL.
Unfortunately, the semantics of those functions do not match MongoDB behavior in edge cases like comparison or sorting of different types.
That resulted in a difference in behavior between FerretDB and MongoDB, and that is a problem we wanted to fix.

So starting from this release we fetch more data from PostgreSQL and perform filtering on the FerretDB side.
This allows us to match MongoDB behavior in all cases.
Of course, that also greatly reduces performance.
We plan to address it in future releases by pushing down parts of filtering queries that can be made fully compatible with MongoDB.
For example, a simple query like `db.collection.find({_id: 'some-id-value'})` can be converted to SQL `WHERE` condition relatively easy and be compatible even with weird values like IEEE 754 NaNs, infinities, etc.

In short, we want FerretDB to be compatible with MongoDB first and fast second, and we are still working towards the first goal.

### New Features 🎉
* Implement `$bitsAllClear` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/394
* Support `$elemMatch` projection query operator by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/383
* Support all bitwise query operators by @w84thesun in https://github.com/FerretDB/FerretDB/pull/400
* Support `$eq` comparision operator by @ribaraka in https://github.com/FerretDB/FerretDB/pull/309
### Fixed Bugs 🐛
* Fix a few issues found by fuzzing by @AlekSi in https://github.com/FerretDB/FerretDB/pull/345
* More fixes for bugs found by fuzzing by @AlekSi in https://github.com/FerretDB/FerretDB/pull/346
* Commands are case-sensitive by @AlekSi in https://github.com/FerretDB/FerretDB/pull/369
* Make updates work by @AlekSi in https://github.com/FerretDB/FerretDB/pull/385
* Handle any number type for `limit` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/399
* Fix numbers comparision by @ribaraka in https://github.com/FerretDB/FerretDB/pull/356
* Fix finding zero documents with `FindOne` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/409
* Fix `sort` for arrays and documents by @AlekSi in https://github.com/FerretDB/FerretDB/pull/424
* Fix pgdb helpers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/425
### Enhancements 🛠
* Update `insert` command's help by @narqo in https://github.com/FerretDB/FerretDB/pull/321
* Return correct error codes for projections by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/384
* Add SortDocuments by @w84thesun in https://github.com/FerretDB/FerretDB/pull/378
### Documentation 📄
* Add Docker badge by @AlekSi in https://github.com/FerretDB/FerretDB/pull/305
* Tweak Markdown linter a bit by @AlekSi in https://github.com/FerretDB/FerretDB/pull/393
### Other Changes 🤖
* Remove Docker volumes on `make env-down` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/315
* Update deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/320
* Build static binaries by @AlekSi in https://github.com/FerretDB/FerretDB/pull/322
* Integrate with dance PRs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/324
* Bump Docker images by @AlekSi in https://github.com/FerretDB/FerretDB/pull/325
* Bump github.com/golangci/golangci-lint from 1.44.0 to 1.44.2 in /tools by @dependabot in https://github.com/FerretDB/FerretDB/pull/327
* Bump mvdan.cc/gofumpt from 0.2.1 to 0.3.0 in /tools by @dependabot in https://github.com/FerretDB/FerretDB/pull/329
* Bump actions/checkout from 2 to 3 by @dependabot in https://github.com/FerretDB/FerretDB/pull/333
* Various small cleanups by @AlekSi in https://github.com/FerretDB/FerretDB/pull/334
* Rewrite `generate.sh` in Go by @w84thesun in https://github.com/FerretDB/FerretDB/pull/338
* Add helper for getting required parameters by @AlekSi in https://github.com/FerretDB/FerretDB/pull/339
* Use safe type assertions for inputs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/341
* Fix seed fuzz corpus collection by @AlekSi in https://github.com/FerretDB/FerretDB/pull/340
* Add fuzzing tests for handlers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/328
* Add `bin/task` to tools by @AlekSi in https://github.com/FerretDB/FerretDB/pull/349
* Add more checks for Go versions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/350
* Improve Windows tooling by @w84thesun in https://github.com/FerretDB/FerretDB/pull/348
* Add assertions for BSON values comparision by @AlekSi in https://github.com/FerretDB/FerretDB/pull/352
* Replace `Makefile` with `Taskfile` by @w84thesun in https://github.com/FerretDB/FerretDB/pull/358
* Fix Taskfile by @w84thesun in https://github.com/FerretDB/FerretDB/pull/365
* Remove OS-specific Taskfiles, cleanup by @AlekSi in https://github.com/FerretDB/FerretDB/pull/366
* Remove SQL storage by @w84thesun in https://github.com/FerretDB/FerretDB/pull/367
* Use square brackets for nicer logs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/373
* Fix build tags by @AlekSi in https://github.com/FerretDB/FerretDB/pull/374
* Add converter from types.Regex to regexp.Regexp by @AlekSi in https://github.com/FerretDB/FerretDB/pull/375
* Log test failures for updates and deletes by @AlekSi in https://github.com/FerretDB/FerretDB/pull/376
* Filter documents using Go code by @AlekSi in https://github.com/FerretDB/FerretDB/pull/370
* Projection: `<field>: <1 or true>` and `<field>: <0 or false>` by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/377
* Fix small issues after rewrite by @AlekSi in https://github.com/FerretDB/FerretDB/pull/380
* Projection: `<field>: <1 or true>` and `<field>: <0 or false>`: error messages formatting  by @seeforschauer in https://github.com/FerretDB/FerretDB/pull/382
* Bump dependecnies by @AlekSi in https://github.com/FerretDB/FerretDB/pull/387
* Fix some fluky tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/351
* Minor CI and build tweaks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/390
* Add Markdown linter by @fenogentov in https://github.com/FerretDB/FerretDB/pull/386
* Do not cache modules by @AlekSi in https://github.com/FerretDB/FerretDB/pull/392
* Fix more fluky tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/391
* Fix the last fluky test by @AlekSi in https://github.com/FerretDB/FerretDB/pull/395
* Allow access to actual listener's address by @AlekSi in https://github.com/FerretDB/FerretDB/pull/397
* Add a new way to write integration tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/389
* Move `internal/pg` to `internal/handlers/pg/pgdb` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/401
* Move `handlers/jsonb1` to `handlers/pg` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/402
* Move handler to `pg` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/403
* Move tests back for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/404
* Use `testutil.AssertEqual` helper by @AlekSi in https://github.com/FerretDB/FerretDB/pull/407
* Move `$size` tests to integration tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/410
* Improve logging in integration tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/412
* Tweak `hello`/`ismaster`/`isMaster` responses by @AlekSi in https://github.com/FerretDB/FerretDB/pull/418
* Fix named loggers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/427
* Add tests for `getLog` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/421
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/430

## New Contributors
* @narqo made their first contribution in https://github.com/FerretDB/FerretDB/pull/321
* @w84thesun made their first contribution in https://github.com/FerretDB/FerretDB/pull/338
* @seeforschauer made their first contribution in https://github.com/FerretDB/FerretDB/pull/377
* @fenogentov made their first contribution in https://github.com/FerretDB/FerretDB/pull/386
* @ribaraka made their first contribution in https://github.com/FerretDB/FerretDB/pull/356

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/12?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.0.6...v0.1.0).


## [v0.0.6](https://github.com/FerretDB/FerretDB/releases/tag/v0.0.6) (2022-02-10)

### New Features 🎉
* Support projections by @ekalinin in https://github.com/FerretDB/FerretDB/pull/212
* Support `dbStats` by @ekalinin in https://github.com/FerretDB/FerretDB/pull/232
* Support `dataSize` by @ekalinin in https://github.com/FerretDB/FerretDB/pull/246
* Implement `listCommands` by @OpenSauce in https://github.com/FerretDB/FerretDB/pull/203
* Support `serverStatus` by @ekalinin in https://github.com/FerretDB/FerretDB/pull/289
* Add more metrics by @AlekSi in https://github.com/FerretDB/FerretDB/pull/298
* Implement `$size` query operator by @taaraora in https://github.com/FerretDB/FerretDB/pull/296
### Fixed Bugs 🐛
* Forbid short document keys like `$k` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/234
* Fix benchmarks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/236
* Move handler tests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/239
* Fix and enable fuzzing by @AlekSi in https://github.com/FerretDB/FerretDB/pull/240
* Make `db.collection.stats()` & `.dataSize()` work from `mongosh` by @ekalinin in https://github.com/FerretDB/FerretDB/pull/243
* fix: remove amd-v2 limit by @Junnplus in https://github.com/FerretDB/FerretDB/pull/282
* Catch concurrent schema/table creation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/283
* Ignore some parameters by @AlekSi in https://github.com/FerretDB/FerretDB/pull/310
### Enhancements 🛠
* Add `buildEnvironment` and `debug` to `buildInfo` command by @GinGin3203 in https://github.com/FerretDB/FerretDB/pull/218
* Add helper for checking for unimplemented fields by @AlekSi in https://github.com/FerretDB/FerretDB/pull/267
* Ignore `authorizedXXX` parameters for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/311
### Documentation 📄
* Update documentation about `fjson` package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/262
* Update tutorial ,add depends_on in docker-compose by @muyouming in https://github.com/FerretDB/FerretDB/pull/275
### Other Changes 🤖
* Bump github.com/reviewdog/reviewdog from 0.13.0 to 0.13.1 in /tools by @dependabot in https://github.com/FerretDB/FerretDB/pull/222
* Fix Docker workflow by @AlekSi in https://github.com/FerretDB/FerretDB/pull/225
* Extract `fjson` package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/207
* Fix test for `collstats` by @ekalinin in https://github.com/FerretDB/FerretDB/pull/233
* Bump go.uber.org/zap from 1.19.1 to 1.20.0 by @dependabot in https://github.com/FerretDB/FerretDB/pull/241
* Use generics for CompositeType by @AlekSi in https://github.com/FerretDB/FerretDB/pull/245
* Enable go-consistent by @AlekSi in https://github.com/FerretDB/FerretDB/pull/248
* Unexport `fjson` types by @AlekSi in https://github.com/FerretDB/FerretDB/pull/231
* Remove JSON methods from bson package by @AlekSi in https://github.com/FerretDB/FerretDB/pull/259
* Fix `make gen` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/264
* Add fuzztool by @ferretdb-bot in https://github.com/FerretDB/FerretDB/pull/56
* Use FerretDB/github-actions/linters by @AlekSi in https://github.com/FerretDB/FerretDB/pull/265
* Make PRs from forks work by @AlekSi in https://github.com/FerretDB/FerretDB/pull/266
* Use `types.Null` instead of `nil` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/268
* Add `make fuzz-corpus` target by @AlekSi in https://github.com/FerretDB/FerretDB/pull/279
* Pass Documents by pointer by @AlekSi in https://github.com/FerretDB/FerretDB/pull/272
* Unexport some `bson` types by @AlekSi in https://github.com/FerretDB/FerretDB/pull/280
* Rename receivers by @AlekSi in https://github.com/FerretDB/FerretDB/pull/284
* Bump github.com/prometheus/client_golang from 1.11.0 to 1.12.0 by @dependabot in https://github.com/FerretDB/FerretDB/pull/285
* Introduce generics for types by @AlekSi in https://github.com/FerretDB/FerretDB/pull/287
* Fix some typos and style by @ekalinin in https://github.com/FerretDB/FerretDB/pull/286
* Add Docker workflow stub by @AlekSi in https://github.com/FerretDB/FerretDB/pull/288
* Split Docker Build and Push by @AlekSi in https://github.com/FerretDB/FerretDB/pull/290
* Securely build and push Docker images by @AlekSi in https://github.com/FerretDB/FerretDB/pull/292
* Update golangci-lint by @AlekSi in https://github.com/FerretDB/FerretDB/pull/294
* Make fuzztool less verbose by @AlekSi in https://github.com/FerretDB/FerretDB/pull/295
* Fix compilation with the latest go tip by @AlekSi in https://github.com/FerretDB/FerretDB/pull/300
* Use `values` MongoDB database by @AlekSi in https://github.com/FerretDB/FerretDB/pull/299
* Spend less time fuzzing pull requests by @AlekSi in https://github.com/FerretDB/FerretDB/pull/302
* Detect matching PR by @AlekSi in https://github.com/FerretDB/FerretDB/pull/303
* Add detection action by @AlekSi in https://github.com/FerretDB/FerretDB/pull/304
* Remove extra allocation by @peakle in https://github.com/FerretDB/FerretDB/pull/307
* Micro fixes: type assert order, strings.Split -> strings.Cut by @peakle in https://github.com/FerretDB/FerretDB/pull/308
* Bump github.com/prometheus/client_golang from 1.12.0 to 1.12.1 by @dependabot in https://github.com/FerretDB/FerretDB/pull/306

## New Contributors
* @Junnplus made their first contribution in https://github.com/FerretDB/FerretDB/pull/282
* @muyouming made their first contribution in https://github.com/FerretDB/FerretDB/pull/275
* @peakle made their first contribution in https://github.com/FerretDB/FerretDB/pull/307
* @taaraora made their first contribution in https://github.com/FerretDB/FerretDB/pull/296

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/10?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.0.5...v0.0.6).


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
* Accept $ and `.` in object field names by @AlekSi in https://github.com/FerretDB/FerretDB/pull/127
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
* Use composite GitHub Action for Go setup in https://github.com/FerretDB/FerretDB/pull/126
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

* A new name in https://github.com/FerretDB/FerretDB/discussions/100
* Added support for databases sizes in `listDatabases` command ([#61](https://github.com/FerretDB/FerretDB/issues/61), thanks to [Leigh](https://github.com/OpenSauce)).

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/3?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.0.3...v0.0.4).


## [v0.0.3](https://github.com/FerretDB/FerretDB/releases/tag/v0.0.3) (2021-11-19)

* Added support for `$regex` evaluation query operator in https://github.com/FerretDB/FerretDB/issues/28
* Fixed handling of Infinity, -Infinity, NaN BSON Double values in https://github.com/FerretDB/FerretDB/issues/29
* Improved documentation for contributors (thanks to [Leigh](https://github.com/OpenSauce)).

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/2?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.0.2...v0.0.3).


## [v0.0.2](https://github.com/FerretDB/FerretDB/releases/tag/v0.0.2) (2021-11-13)

* Added support for comparison query operators: `$eq`,`$gt`,`$gte`,`$in`,`$lt`,`$lte`,`$ne`,`$nin` in https://github.com/FerretDB/FerretDB/issues/26
* Added support for logical query operators: `$and`, `$not`, `$nor`, `$or` in https://github.com/FerretDB/FerretDB/issues/27

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/1?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.0.1...v0.0.2).


## [v0.0.1](https://github.com/FerretDB/FerretDB/releases/tag/v0.0.1) (2021-11-01)

* Initial public release!
