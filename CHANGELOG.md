# Changelog

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
