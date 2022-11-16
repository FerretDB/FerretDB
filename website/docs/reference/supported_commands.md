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


## Aggregation pipelines

The epic - [Issue](https://github.com/FerretDB/FerretDB/issues/9).

| Command     | Argument             | Status | Comments                                                  |
|-------------|----------------------|--------|-----------------------------------------------------------|
| `aggregate` |                      | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1410) |
|             | `$abs`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$accumulator`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$acos`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$acosh`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$add`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$addToSet`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$allElementsTrue`   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$and`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$anyElementTrue`    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$arrayElemAt`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$arrayToObject`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$asin`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$asinh`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$atan`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$atan2`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$atanh`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$avg`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$binarySize`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$bsonSize`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$ceil`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$cmp`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$concat`            | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$concatArrays`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$cond`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$convert`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$cos`               | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$cosh`              | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$count`             | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$covariancePop`     | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$covarianceSamp`    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dateAdd`           | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dateDiff`          | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dateFromPart`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dateFromString`    | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dateSubtract`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dateToParts`       | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dateToString`      | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dateTrunc`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dayOfMonth`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dayOfWeek`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$dayOfYear`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$degreesToRadians ` | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$denseRank`         | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$derivative`        | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | `$divide`                  | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                   | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |

|              | `$geoNear` | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/1412) |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |

|             | `$rand`                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/541)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |
|             | ``                 | ⚠      | [Issue](https://github.com/FerretDB/FerretDB/issues/)     |

| `count`     |            | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1410) |
| `distinct`  |            | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1410) |
| `mapReduce` |            | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1410) |

