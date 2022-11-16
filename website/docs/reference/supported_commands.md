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

| Operator            | Status | Comments                                              |
|---------------------|--------|-------------------------------------------------------|
| `$abs`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$accumulator`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$acos`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$acosh`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$add`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$addToSet`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$allElementsTrue`  | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$and`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$anyElementTrue`   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$arrayElemAt`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$arrayToObject`    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$asin`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$asinh`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$atan`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$atan2`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$atanh`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$avg`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$binarySize`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$bottom`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$bottomN`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$bsonSize`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$ceil`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$cmp`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$concat`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$concatArrays`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$cond`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$convert`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$cos`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$cosh`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$count`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$covariancePop`    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$covarianceSamp`   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dateAdd`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dateDiff`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dateFromParts`    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dateSubtract`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dateTrunc`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dateToParts`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dateFromString`   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dateToString`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dayOfMonth`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dayOfWeek`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$dayOfYear`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$degreesToRadians` | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$denseRank`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$derivative`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$divide`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$documentNumber`   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$eq`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$exp`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$expMovingAvg`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$filter`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$first`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$first`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$firstN`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$firstN`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$floor`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$function`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$getField`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$gt`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$gte`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$hour`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$ifNull`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$in`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$indexOfArray`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$indexOfBytes`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$indexOfCP`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$integral`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$isArray`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$isNumber`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$isoDayOfWeek`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$isoWeek`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$isoWeekYear`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$last`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$last`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$lastN`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$lastN`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$let`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$linearFill`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$literal`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$ln`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$locf`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$log`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$log10`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$lt`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$lte`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$ltrim`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$map`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$max`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$maxN`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$maxN`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$mergeObjects`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$meta`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$min`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$minN`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$minN`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$millisecond`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$minute`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$mod`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$month`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$multiply`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$ne`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$not`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$objectToArray`    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$or`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$pow`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$push`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$radiansToDegrees` | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$rand`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$range`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$rank`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$reduce`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$regexFind`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$regexFindAll`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$regexMatch`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$replaceOne`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$replaceAll`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$reverseArray`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$round`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$rtrim`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$sampleRate`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$second`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$setDifference`    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$setEquals`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$setField`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$setIntersection`  | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$setIsSubset`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$setUnion`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$shift`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$size`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$sin`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$sinh`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$slice`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$sortArray`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$split`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$sqrt`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$stdDevPop`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$stdDevSamp`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$strcasecmp`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$strLenBytes`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$strLenCP`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$substr`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$substrBytes`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$substrCP`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$subtract`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$sum`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$switch`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$tan`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$tanh`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$toBool`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$toDate`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$toDecimal`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$toDouble`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$toInt`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$toLong`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$toObjectId`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$top`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$topN`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$toString`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$toLower`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$toUpper`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$trim`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$trunc`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$tsIncrement`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$tsSecond`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$type`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$unsetField`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$week`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$year`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
| `$zip`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/) |
