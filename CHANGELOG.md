# Changelog

## [v1.2.1](https://github.com/FerretDB/FerretDB/releases/tag/v1.2.1) (2023-05-24)

### Fixed Bugs 🐛
* Fix reporting of updates availability by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2653

### Other Changes 🤖
* Return a better error for authentication problems by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2703

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/43?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.2.0...v1.2.1).


## [v1.2.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.2.0) (2023-05-22)

### What's Changed

This release includes a highly experimental and unsupported SQLite backend.
It will be improved in future releases.

### Fixed Bugs 🐛
* Fix compatibility with C# driver by @b1ron in https://github.com/FerretDB/FerretDB/pull/2613
* Fix a bug with unset field sorting by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2638
* Return `int64` values for `dbStats` and `collStats` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2642
* Return command error from `findAndModify` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2646
* Fix index creation on nested fields by @wqhhust in https://github.com/FerretDB/FerretDB/pull/2637

### Enhancements 🛠
* Perform `insertMany` in a single transaction by @raeidish in https://github.com/FerretDB/FerretDB/pull/2532
* Relax PostgreSQL connection checks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2602
* Cleanup `insert` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/2609
* Support dot notation in projection by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2536

### Documentation 📄
* Add FerretDB v1.1.0 release blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/2594
* Update blog post image for 1.1.0 by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2601
* Add documentation for `.rpm` packages by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2604
* Fix a typo in a blog post by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2611
* Fix typo on RPM package file name by @christiano in https://github.com/FerretDB/FerretDB/pull/2628
* Update documentation formatting by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2640
* Add blog post on "Meteor and FerretDB" by @Fashander in https://github.com/FerretDB/FerretDB/pull/2654

### Other Changes 🤖
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2592
* Remove `TODO` comment for closed issue by @adetunjii in https://github.com/FerretDB/FerretDB/pull/2573
* Add experimental integration test flag for pushdown sorting by @noisersup in https://github.com/FerretDB/FerretDB/pull/2595
* Extract handler parameters from corresponding structure by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2513
* Add `shell` subcommands (`mkdir`, `rmdir`) in `envtool` by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2596
* Add basic postcondition checker for errors by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2607
* Add `sqlite` handler stub by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2608
* Make protocol-level crashes easier to understand by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2610
* Simplify `envtool shell` subcommands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2614
* Cleanup old Docker images  by @wqhhust in https://github.com/FerretDB/FerretDB/pull/2533
* Fix exponential backoff minimum duration by @noisersup in https://github.com/FerretDB/FerretDB/pull/2578
* Fix `count`'s `query` parameter by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2622
* Add a README.md file for assertions by @b1ron in https://github.com/FerretDB/FerretDB/pull/2569
* Use `ExtractParameters` in handlers by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2620
* Verify `OP_MSG` message checksum by @adetunjii in https://github.com/FerretDB/FerretDB/pull/2540
* Separate codebase for aggregation `$project` and query `projection` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2631
* Implement `envtool shell read` subcommand by @wqhhust in https://github.com/FerretDB/FerretDB/pull/2626
* Cleanup projection by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2641
* Add common backend interface prototype by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2619
* Add SQLite handler flags by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2651
* Add tests for aggregation expressions with dots in `$group` aggregation stage by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2636
* Implement some SQLite backend commands by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2655
* Fix tests to assert correct error by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2546
* Aggregation expression refactor by @noisersup in https://github.com/FerretDB/FerretDB/pull/2644
* Move common commands to `commoncommands` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2660
* Add basic observability into backend interfaces by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2661
* Implement metadata storage by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2656
* Add `Query` to the common backend interface by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2662
* Implement query request for SQLite backend  by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2665
* Add test case for read in envtools by @wqhhust in https://github.com/FerretDB/FerretDB/pull/2657
* Run integration tests for `sqlite` handler by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2666
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2671
* Create SQLite directory if needed by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2673
* Implement SQLite `update` and `delete` commands by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2674

### New Contributors
* @adetunjii made their first contribution in https://github.com/FerretDB/FerretDB/pull/2573
* @christiano made their first contribution in https://github.com/FerretDB/FerretDB/pull/2628

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/41?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.1.0...v1.2.0).


## [v1.1.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.1.0) (2023-05-09)

### New Features 🎉
* Implement projection fields assignment by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2484
* Implement `$project` pipeline aggregation stage by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2383
* Handle `create` and `drop` commands in Hana handler by @polyal in https://github.com/FerretDB/FerretDB/pull/2458
* Implement `renameCollection` command by @b1ron in https://github.com/FerretDB/FerretDB/pull/2343

### Fixed Bugs 🐛
* Fix `findAndModify` for `$exists` query operator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2385
* Fix `SchemaStats` to return correct data by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2426
* Fix `findAndModify` for `$set` operator setting `_id` by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2507
* Fix `update` for conflicting dot notation paths by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2521
* Fix `$` path errors for sort by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2534
* Fix empty projections panic by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2562
* Fix `runCommand`'s inserts of documents without `_id`s by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2574

### Enhancements 🛠
* Validate `scale` param for `dbStats` and `collStats` correctly by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2418
* Allow database name contain uppercase characters by @syasyayas in https://github.com/FerretDB/FerretDB/pull/2504
* Add identifying Arch Linux version in `hostInfo` command by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2525
* Handle absent `os-release` file by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2541
* Improve handling of `os-release` files by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2553

### Documentation 📄
* Document test script by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2353
* Use `draft` instead of `unlisted` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2372
* Make example docker compose file restart on failure by @noisersup in https://github.com/FerretDB/FerretDB/pull/2376
* Document how to get logs by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2355
* Update writing guide by @Fashander in https://github.com/FerretDB/FerretDB/pull/2373
* Add comments to our documentation workflow by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2390
* Add blogpost: Announcing FerretDB 1.0 GA - a truly Open Source MongoDB alternative by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2397
* Update documentation for index options by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2417
* Add query pushdown documentation by @noisersup in https://github.com/FerretDB/FerretDB/pull/2339
* Update README.md to link to SSPL by @cooljeanius in https://github.com/FerretDB/FerretDB/pull/2420
* Improve documentation for Docker by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2396
* Add more detailed PR guides in CONTRIBUTING.md by @AuruTus in https://github.com/FerretDB/FerretDB/pull/2435
* Remove a few double spaces by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2431
* Add image for a future blog post by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2453
* Add blogpost - Using FerretDB with Studio 3T by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2454
* Fix YAML indentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2455
* Update blog post on Using FerretDB with Studio3T by @Fashander in https://github.com/FerretDB/FerretDB/pull/2457
* Document `createIndexes`, `listIndexes`, and `dropIndexes` commands by @Fashander in https://github.com/FerretDB/FerretDB/pull/2488

### Other Changes 🤖
* Allow setting "package" variable with a testing flag by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2357
* Make it easier to use Docker-related Task targets by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2358
* Do not mark released binaries as dirty by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2371
* Make Docker Compose flags compatible by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2377
* Bump dependencies by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2367
* Fix version.txt generation for git tags by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2388
* Fix types order linter by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2391
* Cleanup deprecated errors by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2411
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2408
* Use parallel tests consistently by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2409
* Compress CI artifacts by @noisersup in https://github.com/FerretDB/FerretDB/pull/2424
* Use exponential backoff with jitter by @j0holo in https://github.com/FerretDB/FerretDB/pull/2419
* Add Mergify rules for blog posts by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2434
* Migrate to `pgx/v5` by @craigpastro in https://github.com/FerretDB/FerretDB/pull/2439
* Make it harder to misuse iterators by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2428
* Update PR template by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2441
* Rename testing flag by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2437
* Fix compilation on riscv64 by @afiskon in https://github.com/FerretDB/FerretDB/pull/2456
* Cleanup exponential backoff with jitter by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2443
* Add workaround for CockroachDB issue by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2464
* Implement blog posts previews by @noisersup in https://github.com/FerretDB/FerretDB/pull/2433
* Introduce integration benchmarks by @noisersup in https://github.com/FerretDB/FerretDB/pull/2381
* Add tests to findAndModify on `$exists` operator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2422
* Bump deps by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2479
* Refactor aggregation by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2463
* Tweak documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2452
* Fix query projection for top level fields by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2386
* Handle `envtool` panic on timeout by @syasyayas in https://github.com/FerretDB/FerretDB/pull/2499
* Enable debugging tracing of SQL queries by @craigpastro in https://github.com/FerretDB/FerretDB/pull/2467
* Update blog file names to match with slug by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2497
* Add benchmark for replacing large document by @noisersup in https://github.com/FerretDB/FerretDB/pull/2482
* Add more documentation-related items to definition of done by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2494
* Return unsupported operator error for `$` projection operator by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2512
* Use `update_available` from Beacon by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2496
* Use iterator in aggregation stages by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2480
* Increase timeout for tests by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2508
* Add `InsertMany` benchmark by @raeidish in https://github.com/FerretDB/FerretDB/pull/2518
* Add coveralls.io integration by @noisersup in https://github.com/FerretDB/FerretDB/pull/2483
* Add linter for checking blog posts by @raeidish in https://github.com/FerretDB/FerretDB/pull/2459
* Add a YAML formatter by @wqhhust in https://github.com/FerretDB/FerretDB/pull/2485
* Fix `collStats` for Tigris by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2520
* Small addition to YAML formatter usage by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2524
* Cleanup of blog post linter for slug by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2526
* Pushdown simplest sorting for `find` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/2506
* Move `handlers/pg/pjson` to `handlers/sjson` by @craigpastro in https://github.com/FerretDB/FerretDB/pull/2531
* Check test database name length in compat test setup by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2527
* Document `not ready` issues label by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2544
* Remove version and name assertions in integration tests by @raeidish in https://github.com/FerretDB/FerretDB/pull/2552
* Add helpers for iterators and generators by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2542
* Do various small cleanups by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2561
* Pushdown simplest sorting for `aggregate` command by @noisersup in https://github.com/FerretDB/FerretDB/pull/2530
* Move handlers parameters to common by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2529
* Use our own Prettier Docker image by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2535
* Improve fuzzing with recorded seed data by @fenogentov in https://github.com/FerretDB/FerretDB/pull/2392
* Add proper CLI to `envtool` - `envtool setup`  subcommand by @kropidlowsky in https://github.com/FerretDB/FerretDB/pull/2570
* Recover from more errors, close connection less often by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2564
* Tweak issue templates and contributing docs by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2572
* Refactor integration benchmarks by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2537
* Do panic in integration tests if connection can't be established by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2577
* Small refactoring by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2575
* Merge `no ci` label into `not ready` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2580

### New Contributors
* @cooljeanius made their first contribution in https://github.com/FerretDB/FerretDB/pull/2420
* @j0holo made their first contribution in https://github.com/FerretDB/FerretDB/pull/2419
* @AuruTus made their first contribution in https://github.com/FerretDB/FerretDB/pull/2435
* @craigpastro made their first contribution in https://github.com/FerretDB/FerretDB/pull/2439
* @afiskon made their first contribution in https://github.com/FerretDB/FerretDB/pull/2456
* @syasyayas made their first contribution in https://github.com/FerretDB/FerretDB/pull/2499
* @raeidish made their first contribution in https://github.com/FerretDB/FerretDB/pull/2518
* @polyal made their first contribution in https://github.com/FerretDB/FerretDB/pull/2458
* @wqhhust made their first contribution in https://github.com/FerretDB/FerretDB/pull/2485

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/40?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v1.0.0...v1.1.0).


## [v1.0.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.0.0) (2023-04-03)

### What's Changed
We are delighted to announce the release of FerretDB 1.0 GA!

### New Features 🎉
* Support `$sum` accumulator of `$group` aggregation by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2292
* Implement `createIndexes` command by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2244
* Add basic `getMore` command by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2309
* Implement `dropIndexes` command by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2313
* Implement `$limit` aggregation pipeline stage by @noisersup in https://github.com/FerretDB/FerretDB/pull/2270
* Add partial support for `collStats`, `dbStats` and `dataSize` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2322
* Implement `$skip` aggregation pipeline stage by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2310
* Implement `$unwind` aggregation pipeline stage by @noisersup in https://github.com/FerretDB/FerretDB/pull/2294
* Support `count` and `storageStats` fields in `$collStats` aggregation pipeline stage by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2338

### Fixed Bugs 🐛
* Fix dot notation negative index errors by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2246
* Apply `skip` before `limit` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2351

### Documentation 📄
* Update supported command for `$sum` aggregation operator by @chilagrow in https://github.com/FerretDB/FerretDB/pull/2318
* Add supported shells and GUIs images by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2323
* Publish FerretDB v0.9.4 blog post by @Fashander in https://github.com/FerretDB/FerretDB/pull/2268
* Use dashes instead of underscores or spaces by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2329
* Update documentation sidebar by @Fashander in https://github.com/FerretDB/FerretDB/pull/2347
* Update FerretDB descriptions by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2281
* Improve flags documentation by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2331
* Describe supported fields for `$collStats` aggregation stage by @rumyantseva in https://github.com/FerretDB/FerretDB/pull/2352

### Other Changes 🤖
* Use iterators for `sort`, `limit`, `skip`, and `projection` by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2254
* Bump dependencies by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2307
* Improve resource tracking by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2319
* Add tests for `find`'s and `count`'s `skip` argument by @w84thesun in https://github.com/FerretDB/FerretDB/pull/2325
* Close iterator properly by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2333
* Improve large numbers initialization in test data by @noisersup in https://github.com/FerretDB/FerretDB/pull/2324
* Ignore `unique` index option for now by @AlekSi in https://github.com/FerretDB/FerretDB/pull/2350

[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/39?closed=1).
[All commits](https://github.com/FerretDB/FerretDB/compare/v0.9.4...v1.0.0).


## Older Releases

See <https://github.com/FerretDB/FerretDB/blob/v1.0.0/CHANGELOG.md>.
