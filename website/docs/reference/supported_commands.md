---
sidebar_position: 1
---

# Supported commands

| Command           | Argument                | Status | Comments                                                  |
|-------------------|-------------------------|--------|-----------------------------------------------------------|
| `listCollections` |                         | ✅     | Basic command is fully supported                          |
|                   | `filter`                | ❌     | [Issue](https://github.com/FerretDB/FerretDB/issues/1405) |
|                   | `nameOnly`              | ❌     | [Issue](https://github.com/FerretDB/FerretDB/issues/301)  |
|                   | `comment`               | ⚠️     | Ignored                                                   |
|                   | `authorizedCollections` | ⚠️     | Ignored                                                   |
| `delete`          |                         | ✅     | Basic command is fully supported                          |
|                   | `deletes`               | ✅     |                                                           |
|                   | `comment`               | ⚠️     | Ignored                                                   |
|                   | `let`                   | ⚠️     | Unimplemented                                             |
|                   | `ordered`               | ✅     |                                                           |
|                   | `writeConcern`          | ⚠️     | Ignored                                                   |
|                   | `q`                     | ✅     |                                                           |
|                   | `limit`                 | ✅     |                                                           |
|                   | `collation`             | ⚠️     | Unimplemented                                             |
|                   | `hint`                  | ❌     | Unimplemented                                             |
| `find`            |                         | ✅     | Basic command is fully supported                          |
|                   | `filter`                | ✅     |                                                           |
|                   | `sort`                  | ✅     |                                                           |
|                   | `projection`            | ✅     |                                                           |
|                   | `hint`                  | ❌     | Ignored                                                   |
|                   | `skip`                  | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/1445) |
|                   | `limit`                 | ✅     |                                                           |
|                   | `batchSize`             | ⚠️     | Unimplemented                                             |
|                   | `singleBatch`           | ⚠️     | Unimplemented                                             |
|                   | `comment`               | ⚠️     | Not implemented in Tigris                                 |
|                   | `maxTimeMS`             | ✅     |                                                           |
|                   | `readConcern`           | ⚠️     | Ignored                                                   |
|                   | `max`                   | ⚠️     | Ignored                                                   |
|                   | `min`                   | ⚠️     | Ignored                                                   |
|                   | `returnKey`             | ⚠️     | Unimplemented                                             |
|                   | `showRecordId`          | ⚠️     | Unimplemented                                             |
|                   | `tailable`              | ❌     | Unimplemented                                             |
|                   | `awaitData`             | ⚠️     | Unimplemented                                             |
|                   | `oplogReplay`           |        | Deprecated since version 4.4.                             |
|                   | `noCursorTimeout`       | ⚠️     | Unimplemented                                             |
|                   | `allowPartialResults`   | ⚠️     | Unimplemented                                             |
|                   | `collation`             | ⚠️     | Unimplemented                                             |
|                   | `allowDiskUse`          | ⚠️     | Unimplemented                                             |
|                   | `let`                   | ⚠️     | Unimplemented                                             |
| `findAndModify`   |                         | ✅     | Basic command is fully supported                          |
|                   | `query`                 | ✅     |                                                           |
|                   | `sort`                  | ✅     |                                                           |
|                   | `remove`                | ✅     |                                                           |
|                   | `update`                | ✅     |                                                           |
|                   | `new`                   | ✅     |                                                           |
|                   | `upsert`                | ✅     |                                                           |
|                   | `bypassDocumentValidation` | ⚠️  | Ignored                                                   |
|                   | `writeConcern`             | ⚠️  | Ignored                                                   |
|                   | `maxTimeMS`                | ✅  |                                                           |
|                   | `collation`                | ⚠️  | Ignored                                                   |
|                   | `arrayFilters`             | ❌  | Unimplemented                                             |
|                   | `hint`                     | ❌  | Ignored                                                   |
|                   | `comment`                  | ⚠️  | Not implemented in Tigris                                 |
|                   | `let`                      | ⚠️  | Unimplemented                                             |
| `getMore`         |                            | ❌  | Unimplemented                                             |
| `insert`          |                            | ✅  | Basic command is fully supported                          |
|                   | `documents`                | ✅  |                                                           |
|                   | `ordered`                  | ❌  | [Issue](https://github.com/FerretDB/FerretDB/issues/940)  |
|                   | `bypassDocumentValidation` | ⚠️  | Ignored                                                   |
|                   | `comment`                  | ⚠️  | Ignored                                                   |
| `resetError`      |                            |     | Removed in MongoDB 5.0.                                   |
| `update`          |                            | ✅  | Basic command is fully supported                          |
|                   | `updates`                  | ✅  |                                                           |
|                   | `ordered`                  | ⚠️  | Ignored                                                   |
|                   | `writeConcern`             | ⚠️  | Ignored                                                   |
|                   | `bypassDocumentValidation` | ⚠️  | Ignored                                                   |
|                   | `comment`                  | ⚠️  | Ignored in Tigris                                         |
|                   | `let`                      | ⚠️  | Unimplemented                                             |
|                   | `q`                        | ✅  |                                                           |
|                   | `u`                        | ✅  | TODO check if u is an array of aggregation pipeline stages|
|                   | `c`                        | ⚠️  | Unimplemented                                             |
|                   | `upsert`                   | ✅  |                                                           |
|                   | `multi`                    | ✅  |                                                           |
|                   | `collation`                | ⚠️  | Unimplemented                                             |
|                   | `arrayFilters`             | ⚠️  | Unimplemented                                             |
|                   | `hint`                     | ⚠️  | Unimplemented                                             |

## Aggregation pipelines

The epic - [Issue](https://github.com/FerretDB/FerretDB/issues/9).

| Command     | Argument             | Status | Comments                                                  |
|-------------|----------------------|--------|-----------------------------------------------------------|
| `aggregate` |                      | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1410) |

### Collection stages

```js
db.collection.aggregate()
```

| Stage                          | Status | Comments                                                  |
|--------------------------------|--------|-----------------------------------------------------------|
| `$addFields`, `$set`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1413) |
| `$bucket`, `$bucketAuto`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1414) |
| `$changeStream`                | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1415) |
| `$collStats`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1416) |
| `$count`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1417) |
| `$densify`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1418) |
| `$documents`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1419) |
| `$facet`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1420) |
| `$fill`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1421) |
| `$geoNear`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1412) |
| `$graphLookup`                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1422) |
| `$group`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1423) |
| `$indexStats`                  | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1424) |
| `$limit`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1425) |
| `$listSessions`                | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1426) |
| `$lookup`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1427) |
| `$match`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1428) |
| `$merge`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1429) |
| `$out`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1430) |
| `$planCacheStats`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1431) |
| `$project`, `$unset`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1432) |
| `$redact`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1433) |
| `$replaceRoot`, `$replaceWith` | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1434) |
| `$sample`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1435) |
| `$search`, `$searchMeta`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1436) |
| `$setWindowFields`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1437) |
| `$skip`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1438) |
| `$sort`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1439) |
| `$sortByCount`                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1440) |
| `$unionWith`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1441) |
| `$unwind`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1442) |
