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

### Aggregation collection stages

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

### Aggregation database stages

```js
db.aggregate()
```

| Stage                | Status | Comments                                                  |
|----------------------|--------|-----------------------------------------------------------|
| `$changeStream`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1415) |
| `$currentOp`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1444) |
| `$listLocalSessions` | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1426) |
| `$documents`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1419) |

### Aggregation pipeline operators

| Operator                          | Status | Comments                                                  |
|-----------------------------------|--------|-----------------------------------------------------------|
| `$abs`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$accumulator`                    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1458) |
| `$acos`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$acosh`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$add` (arithmetic operator)      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$add` (date operator)            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$addToSet`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$allElementsTrue`                | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$and`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1455) |
| `$anyElementTrue`                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$arrayElemAt`                    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$arrayToObject`                  | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$asin`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$asinh`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$atan`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$atan2`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$atanh`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$avg`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$binarySize`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1459) |
| `$bottom`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$bottomN`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$bsonSize`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1459) |
| `$ceil`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$cmp`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$concat`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$concatArrays`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$cond`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1457) |
| `$convert`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$cos`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$cosh`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$count`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$covariancePop`                  | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$covarianceSamp`                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$dateAdd`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateDiff`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateFromParts`                  | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateSubtract`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateTrunc`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateToParts`                    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateFromString`                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateToString`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dayOfMonth`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dayOfWeek`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dayOfYear`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$degreesToRadians`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$denseRank`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$derivative`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$divide`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$documentNumber`                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$eq`                             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$exp`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$expMovingAvg`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$filter`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$first` (array operator)         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$first`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$firstN`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$firstN`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$floor`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$function`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1458) |
| `$getField`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$gt`                             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$gte`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$hour`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$ifNull`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1457) |
| `$in`                             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$indexOfArray`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$indexOfBytes`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$indexOfCP`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$integral`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$isArray`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$isNumber`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$isoDayOfWeek`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$isoWeek`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$isoWeekYear`                    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$last` (array operator)          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$last`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$lastN`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$lastN`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$let`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$linearFill`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$literal`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$ln`                             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$locf`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$log`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$log10`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$lt`                             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$lte`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$ltrim`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$map`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$max`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$maxN`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$maxN`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$mergeObjects`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$meta`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$min`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$minN`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$minN`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$millisecond`                    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$minute`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$mod`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$month`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$multiply`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$ne`                             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$not`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1455) |
| `$objectToArray`                  | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$or`                             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1455) |
| `$pow`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$push`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$radiansToDegrees`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$rand`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$range`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$rank`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$reduce`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$regexFind`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$regexFindAll`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$regexMatch`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$replaceOne`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$replaceAll`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$reverseArray`                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$round`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$rtrim`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$sampleRate`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$second`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$setDifference`                  | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$setEquals`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$setField`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$setIntersection`                | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$setIsSubset`                    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$setUnion`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$shift`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$size`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$sin`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$sinh`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$slice`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$sortArray`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$split`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$sqrt`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$stdDevPop`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$stdDevSamp`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$strcasecmp`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$strLenBytes`                    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$strLenCP`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$substr`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$substrBytes`                    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$substrCP`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$subtract` (arithmetic operator) | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$subtract` (date operator)       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$sum`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$switch`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1457) |
| `$tan`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$tanh`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$toBool`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$toDate`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$toDecimal`                      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$toDouble`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$toInt`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$toLong`                         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$toObjectId`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$top`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$topN`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$toString`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$toLower`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$toUpper`                        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$trim`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$trunc`                          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$tsIncrement`                    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$tsSecond`                       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$type`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$unsetField`                     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
| `$week`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$year`                           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$zip`                            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
