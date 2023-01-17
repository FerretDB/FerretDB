---
sidebar_position: 1
---

# Supported commands

## Query commands

| Command         | Argument                   | Status | Comments                                                   |
| --------------- | -------------------------- | ------ | ---------------------------------------------------------- |
| `delete`        |                            | ✅      | Basic command is fully supported                           |
|                 | `deletes`                  | ✅      |                                                            |
|                 | `comment`                  | ⚠️      | Ignored in Tigris                                          |
|                 | `let`                      | ⚠️      | Unimplemented                                              |
|                 | `ordered`                  | ✅      |                                                            |
|                 | `writeConcern`             | ⚠️      | Ignored                                                    |
|                 | `q`                        | ✅      |                                                            |
|                 | `limit`                    | ✅      |                                                            |
|                 | `collation`                | ⚠️      | Unimplemented                                              |
|                 | `hint`                     | ❌      | Unimplemented                                              |
| `find`          |                            | ✅      | Basic command is fully supported                           |
|                 | `filter`                   | ✅      |                                                            |
|                 | `sort`                     | ✅      |                                                            |
|                 | `projection`               | ✅      | Basic projections with fields is supported                 |
|                 | `hint`                     | ❌      | Ignored                                                    |
|                 | `skip`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1445)  |
|                 | `limit`                    | ✅      |                                                            |
|                 | `batchSize`                | ⚠️      | Unimplemented                                              |
|                 | `singleBatch`              | ⚠️      | Unimplemented                                              |
|                 | `comment`                  | ⚠️      | Not implemented in Tigris                                  |
|                 | `maxTimeMS`                | ✅      |                                                            |
|                 | `readConcern`              | ⚠️      | Ignored                                                    |
|                 | `max`                      | ⚠️      | Ignored                                                    |
|                 | `min`                      | ⚠️      | Ignored                                                    |
|                 | `returnKey`                | ⚠️      | Unimplemented                                              |
|                 | `showRecordId`             | ⚠️      | Unimplemented                                              |
|                 | `tailable`                 | ❌      | Unimplemented                                              |
|                 | `awaitData`                | ⚠️      | Unimplemented                                              |
|                 | `noCursorTimeout`          | ⚠️      | Unimplemented                                              |
|                 | `allowPartialResults`      | ⚠️      | Unimplemented                                              |
|                 | `collation`                | ⚠️      | Unimplemented                                              |
|                 | `allowDiskUse`             | ⚠️      | Unimplemented                                              |
|                 | `let`                      | ⚠️      | Unimplemented                                              |
| `findAndModify` |                            | ✅      | Basic command is fully supported                           |
|                 | `query`                    | ✅      |                                                            |
|                 | `sort`                     | ✅      |                                                            |
|                 | `remove`                   | ✅      |                                                            |
|                 | `update`                   | ✅      |                                                            |
|                 | `new`                      | ✅      |                                                            |
|                 | `upsert`                   | ✅      |                                                            |
|                 | `bypassDocumentValidation` | ⚠️      | Ignored                                                    |
|                 | `writeConcern`             | ⚠️      | Ignored                                                    |
|                 | `maxTimeMS`                | ✅      |                                                            |
|                 | `collation`                | ⚠️      | Ignored                                                    |
|                 | `arrayFilters`             | ❌      | Unimplemented                                              |
|                 | `hint`                     | ❌      | Ignored                                                    |
|                 | `comment`                  | ⚠️      | Not implemented in Tigris                                  |
|                 | `let`                      | ⚠️      | Unimplemented                                              |
| `getMore`       |                            | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1733)  |
| `insert`        |                            | ✅      | Basic command is fully supported                           |
|                 | `documents`                | ✅      |                                                            |
|                 | `ordered`                  | ✅      |                                                            |
|                 | `bypassDocumentValidation` | ⚠️      | Ignored                                                    |
|                 | `comment`                  | ⚠️      | Ignored                                                    |
| `update`        |                            | ✅      | Basic command is fully supported                           |
|                 | `updates`                  | ✅      |                                                            |
|                 | `ordered`                  | ⚠️      | Ignored                                                    |
|                 | `writeConcern`             | ⚠️      | Ignored                                                    |
|                 | `bypassDocumentValidation` | ⚠️      | Ignored                                                    |
|                 | `comment`                  | ⚠️      | Ignored in Tigris                                          |
|                 | `let`                      | ⚠️      | Unimplemented                                              |
|                 | `q`                        | ✅      |                                                            |
|                 | `u`                        | ✅      | TODO check if u is an array of aggregation pipeline stages |
|                 | `c`                        | ⚠️      | Unimplemented                                              |
|                 | `upsert`                   | ✅      |                                                            |
|                 | `multi`                    | ✅      |                                                            |
|                 | `collation`                | ⚠️      | Unimplemented                                              |
|                 | `arrayFilters`             | ⚠️      | Unimplemented                                              |
|                 | `hint`                     | ⚠️      | Unimplemented                                              |

### Update Operators

The following operators and modifiers are available in the `update` and `findAndModify` commands.

| Operator          | Modifier    | Status | Comments                                                 |
|-------------------|-------------|--------|----------------------------------------------------------|
| `$currentDate`    |             | ✅      |                                                          |
| `$inc`            |             | ✅      |                                                          |
| `$min`            |             | ✅      |                                                          |
| `$max`            |             | ✅      |                                                          |
| `$mul`            |             | ✅      |                                                          |
| `$rename`         |             | ✅      |                                                          |
| `$set`            |             | ✅      |                                                          |
| `$setOnInsert`    |             | ✅      |                                                          |
| `$unset`          |             | ✅      |                                                          |
| `$`               |             | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/822) |
| `$[]`             |             | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/823) |
| `$[<identifier>]` |             | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/824) |
| `$addToSet`       |             | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/825) |
| `$pop`            |             | ✅      |                                                          |
| `$pull`           |             | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/826) |
| `$push`           |             | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/503) |
| `$pullAll`        |             | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/827) |
|                   | `$each`     | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/828) |
|                   | `$position` | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/829) |
|                   | `$slice`    | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/830) |
|                   | `$sort`     | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/831) |
|                   | `$bit`      | ⚠️     | [Issue](https://github.com/FerretDB/FerretDB/issues/821) |

### Projection Operators

The following operators are available in the `find` command `projection` argument.

| Operator     | Status | Comments                                                  |
| ------------ | ------ | --------------------------------------------------------- |
| `$`          | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1709) |
| `$elemMatch` | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1710) |
| `$meta`      | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1712) |
| `$slice`     | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1711) |

## Query Plan Cache Commands

Related epic - [Issue](https://github.com/FerretDB/FerretDB/issues/78).

| Command                 | Argument     | Status | Comments                                                  |
| ----------------------- | ------------ | ------ | --------------------------------------------------------- |
| `planCacheClear`        |              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1502) |
|                         | `query`      | ⚠️      |                                                           |
|                         | `projection` | ⚠️      |                                                           |
|                         | `sort`       | ⚠️      |                                                           |
|                         | `comment`    | ⚠️      |                                                           |
| `planCacheClearFilters` |              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1503) |
|                         | `query`      | ⚠️      |                                                           |
|                         | `sort`       | ⚠️      |                                                           |
|                         | `projection` | ⚠️      |                                                           |
|                         | `collation`  | ⚠️      |                                                           |
|                         | `comment`    | ⚠️      |                                                           |
| `planCacheListFilters`  |              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1504) |
|                         | `comment`    | ⚠️      |                                                           |
| `planCacheSetFilter`    |              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1505) |
|                         | `query`      | ⚠️      |                                                           |
|                         | `sort`       | ⚠️      |                                                           |
|                         | `projection` | ⚠️      |                                                           |
|                         | `collation`  | ⚠️      |                                                           |
|                         | `indexes`    | ⚠️      |                                                           |
|                         | `comment`    | ⚠️      |                                                           |

## Free Monitoring Commands

| Command                   | Argument            | Status | Comments                                               |
| ------------------------- | ------------------- | ------ | ------------------------------------------------------ |
| `setFreeMonitoring`       |                     | ✅      | [Telemetry reporting](/telemetry/)                     |
|                           | `action: "enable"`  | ✅      | [`--telemetry=enable`](/telemetry/#enable-telemetry)   |
|                           | `action: "disable"` | ✅      | [`--telemetry=disable`](/telemetry/#disable-telemetry) |
| `getFreeMonitoringStatus` |                     | ✅      |                                                        |

## Database Operations

### User Management Commands

| Command                    | Argument                         | Status | Comments                                                  |
| -------------------------- | -------------------------------- | ------ | --------------------------------------------------------- |
| `createUser`               |                                  | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1491) |
|                            | `pwd`                            | ⚠️      |                                                           |
|                            | `customData`                     | ⚠️      |                                                           |
|                            | `roles`                          | ⚠️      |                                                           |
|                            | `digestPassword`                 | ⚠️      |                                                           |
|                            | `writeConcern`                   | ⚠️      |                                                           |
|                            | `authenticationRestrictions`     | ⚠️      |                                                           |
|                            | `mechanisms`                     | ⚠️      |                                                           |
|                            | `digestPassword`                 | ⚠️      |                                                           |
|                            | `comment`                        | ⚠️      |                                                           |
| `dropAllUsersFromDatabase` |                                  | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1492) |
|                            | `writeConcern`                   | ⚠️      |                                                           |
|                            | `comment`                        | ⚠️      |                                                           |
| `dropUser`                 |                                  | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1493) |
|                            | `writeConcern`                   | ⚠️      |                                                           |
|                            | `comment`                        | ⚠️      |                                                           |
| `grantRolesToUser`         |                                  | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1494) |
|                            | `writeConcern`                   | ⚠️      |                                                           |
|                            | `comment`                        | ⚠️      |                                                           |
| `revokeRolesFromUser`      |                                  | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1495) |
|                            | `roles`                          | ⚠️      |                                                           |
|                            | `writeConcern`                   | ⚠️      |                                                           |
|                            | `comment`                        | ⚠️      |                                                           |
| `updateUser`               |                                  | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1496) |
|                            | `pwd`                            | ⚠️      |                                                           |
|                            | `customData`                     | ⚠️      |                                                           |
|                            | `roles`                          | ⚠️      |                                                           |
|                            | `digestPassword`                 | ⚠️      |                                                           |
|                            | `writeConcern`                   | ⚠️      |                                                           |
|                            | `authenticationRestrictions`     | ⚠️      |                                                           |
|                            | `mechanisms`                     | ⚠️      |                                                           |
|                            | `digestPassword`                 | ⚠️      |                                                           |
|                            | `comment`                        | ⚠️      |                                                           |
| `usersInfo`                |                                  | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1497) |
|                            | `showCredentials`                | ⚠️      |                                                           |
|                            | `showCustomData`                 | ⚠️      |                                                           |
|                            | `showPrivileges`                 | ⚠️      |                                                           |
|                            | `showAuthenticationRestrictions` | ⚠️      |                                                           |
|                            | `filter`                         | ⚠️      |                                                           |
|                            | `comment`                        | ⚠️      |                                                           |

### Authentication Commands

| Command        | Argument | Status | Comments                                                  |
| -------------- | -------- | ------ | --------------------------------------------------------- |
| `authenticate` |          | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1731) |

### Role Management Commands

| Command                    | Argument                     | Status | Comments                                                  |
| -------------------------- | ---------------------------- | ------ | --------------------------------------------------------- |
| `createRole`               |                              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1528) |
|                            | `privileges`                 | ⚠️      |                                                           |
|                            | `roles`                      | ⚠️      |                                                           |
|                            | `authenticationRestrictions` | ⚠️      |                                                           |
|                            | `writeConcern`               | ⚠️      |                                                           |
|                            | `comment`                    | ⚠️      |                                                           |
| `dropRole`                 |                              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1529) |
|                            | `writeConcern`               | ⚠️      |                                                           |
|                            | `comment`                    | ⚠️      |                                                           |
| `dropAllRolesFromDatabase` |                              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1530) |
|                            | `writeConcern`               | ⚠️      |                                                           |
|                            | `comment`                    | ⚠️      |                                                           |
| `grantPrivilegesToRole`    |                              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1531) |
|                            | `privileges`                 | ⚠️      |                                                           |
|                            | `writeConcern`               | ⚠️      |                                                           |
|                            | `comment`                    | ⚠️      |                                                           |
| `grantRolesToRole`         |                              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1532) |
|                            | `roles`                      | ⚠️      |                                                           |
|                            | `writeConcern`               | ⚠️      |                                                           |
|                            | `comment`                    | ⚠️      |                                                           |
| `invalidateUserCache`      |                              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1533) |
| `revokePrivilegesFromRole` |                              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1534) |
|                            | `privileges`                 | ⚠️      |                                                           |
|                            | `writeConcern`               | ⚠️      |                                                           |
|                            | `comment`                    | ⚠️      |                                                           |
| `revokeRolesFromRole`      |                              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1535) |
|                            | `roles`                      | ⚠️      |                                                           |
|                            | `writeConcern`               | ⚠️      |                                                           |
|                            | `comment`                    | ⚠️      |                                                           |
| `rolesInfo`                |                              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1536) |
|                            | `showPrivileges`             | ⚠️      |                                                           |
|                            | `showBuiltinRoles`           | ⚠️      |                                                           |
|                            | `comment`                    | ⚠️      |                                                           |
| `updateRole`               |                              | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1537) |
|                            | `privileges`                 | ⚠️      |                                                           |
|                            | `roles`                      | ⚠️      |                                                           |
|                            | `authenticationRestrictions` | ⚠️      |                                                           |
|                            | `writeConcern`               | ⚠️      |                                                           |
|                            | `comment`                    | ⚠️      |                                                           |

## Session Commands

Related epic - [Issue](https://github.com/FerretDB/FerretDB/issues/8)

Related epic - [Issue](https://github.com/FerretDB/FerretDB/issues/153)

| Command                    | Argument       | Status | Comments                                                  |
| -------------------------- | -------------- | ------ | --------------------------------------------------------- |
| `abortTransaction`         |                | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1547) |
|                            | `txnNumber`    | ⚠️      |                                                           |
|                            | `writeConcern` | ⚠️      |                                                           |
|                            | `autocommit`   | ⚠️      |                                                           |
|                            | `comment`      | ⚠️      |                                                           |
| `commitTransaction`        |                | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1548) |
|                            | `txnNumber`    | ⚠️      |                                                           |
|                            | `writeConcern` | ⚠️      |                                                           |
|                            | `autocommit`   | ⚠️      |                                                           |
|                            | `comment`      | ⚠️      |                                                           |
| `endSessions`              |                | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1549) |
| `killAllSessions`          |                | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1550) |
| `killAllSessionsByPattern` |                | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1551) |
| `killSessions`             |                | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1552) |
| `refreshSessions`          |                | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1553) |
| `startSession`             |                | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1554) |

## Aggregation pipelines

The epic - [Issue](https://github.com/FerretDB/FerretDB/issues/9).

| Command     | Argument | Status | Comments                                                  |
|-------------|----------|--------|-----------------------------------------------------------|
| `aggregate` |          | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1410) |
| `count`     |          | ✅      |                                                           |
| `distinct`  |          | ✅      |                                                           |

<!-- markdownlint-disable MD001 MD033 -->
<!-- That's the simplest way to remove those sections from the right menu. -->

<details>
<summary>Stages and operators</summary>

#### Aggregation collection stages

```js
db.collection.aggregate()
```

| Stage                          | Status | Comments                                                  |
| ------------------------------ | ------ | --------------------------------------------------------- |
| `$addFields`, `$set`           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1413) |
| `$bucket`, `$bucketAuto`       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1414) |
| `$changeStream`                | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1415) |
| `$collStats`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1416) |
| `$count`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1417) |
| `$densify`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1418) |
| `$documents`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1419) |
| `$facet`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1420) |
| `$fill`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1421) |
| `$geoNear`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1412) |
| `$graphLookup`                 | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1422) |
| `$group`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1423) |
| `$indexStats`                  | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1424) |
| `$limit`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1425) |
| `$listSessions`                | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1426) |
| `$lookup`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1427) |
| `$match`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1428) |
| `$merge`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1429) |
| `$out`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1430) |
| `$planCacheStats`              | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1431) |
| `$project`, `$unset`           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1432) |
| `$redact`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1433) |
| `$replaceRoot`, `$replaceWith` | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1434) |
| `$sample`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1435) |
| `$search`, `$searchMeta`       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1436) |
| `$setWindowFields`             | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1437) |
| `$skip`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1438) |
| `$sort`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1439) |
| `$sortByCount`                 | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1440) |
| `$unionWith`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1441) |
| `$unwind`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1442) |

#### Aggregation database stages

```js
db.aggregate()
```

| Stage                | Status | Comments                                                  |
| -------------------- | ------ | --------------------------------------------------------- |
| `$changeStream`      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1415) |
| `$currentOp`         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1444) |
| `$listLocalSessions` | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1426) |
| `$documents`         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1419) |

#### Aggregation pipeline operators

| Operator                          | Status | Comments                                                  |
| --------------------------------- | ------ | --------------------------------------------------------- |
| `$abs`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$accumulator`                    | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$acos`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$acosh`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$add` (arithmetic operator)      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$add` (date operator)            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$addToSet`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$allElementsTrue`                | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1462) |
| `$and`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1455) |
| `$anyElementTrue`                 | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1462) |
| `$arrayElemAt`                    | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$arrayToObject`                  | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$asin`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$asinh`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$atan`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$atan2`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$atanh`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$avg`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$binarySize`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1459) |
| `$bottom`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$bottomN`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$bsonSize`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1459) |
| `$ceil`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$cmp`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$concat`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$concatArrays`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$cond`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1457) |
| `$convert`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1466) |
| `$cos`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$cosh`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$count`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$covariancePop`                  | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$covarianceSamp`                 | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$dateAdd`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateDiff`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateFromParts`                  | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateSubtract`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateTrunc`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateToParts`                    | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateFromString`                 | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dateToString`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dayOfMonth`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dayOfWeek`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$dayOfYear`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$degreesToRadians`               | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$denseRank`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$derivative`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$divide`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$documentNumber`                 | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$eq`                             | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$exp`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$expMovingAvg`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$filter`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$first` (array operator)         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$first` (accumulator)            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$firstN`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$floor`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$function`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1458) |
| `$getField`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1471) |
| `$gt`                             | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$gte`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$hour`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$ifNull`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1457) |
| `$in`                             | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$indexOfArray`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$indexOfBytes`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$indexOfCP`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$integral`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$isArray`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$isNumber`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1466) |
| `$isoDayOfWeek`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$isoWeek`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$isoWeekYear`                    | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$last` (array operator)          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$last`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$lastN`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$let`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1469) |
| `$linearFill`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$literal`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1470) |
| `$ln`                             | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$locf`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$log`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$log10`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$lt`                             | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$lte`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$ltrim`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$map`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$max`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$maxN`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$mergeObjects`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$meta`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$min`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$minN`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$millisecond`                    | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$minute`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$mod`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$month`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$multiply`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$ne`                             | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1456) |
| `$not`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1455) |
| `$objectToArray`                  | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1461) |
| `$or`                             | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1455) |
| `$pow`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$push`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$radiansToDegrees`               | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$rand`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/541)  |
| `$range`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$rank`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$reduce`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$regexFind`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$regexFindAll`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$regexMatch`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$replaceOne`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$replaceAll`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$reverseArray`                   | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$round`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$rtrim`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$sampleRate`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1472) |
| `$second`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$setDifference`                  | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1462) |
| `$setEquals`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1462) |
| `$setField`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1461) |
| `$setIntersection`                | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1462) |
| `$setIsSubset`                    | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1462) |
| `$setUnion`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1462) |
| `$shift`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1468) |
| `$size`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$sin`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$sinh`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$slice`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$sortArray`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |
| `$split`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$sqrt`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$stdDevPop`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$stdDevSamp`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$strcasecmp`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$strLenBytes`                    | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$strLenCP`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$substr`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$substrBytes`                    | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$substrCP`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$subtract` (arithmetic operator) | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$subtract` (date operator)       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$sum`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$switch`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1457) |
| `$tan`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$tanh`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1465) |
| `$toBool`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1466) |
| `$toDate`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$toDecimal`                      | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1466) |
| `$toDouble`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1466) |
| `$toInt`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1466) |
| `$toLong`                         | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1466) |
| `$toObjectId`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1466) |
| `$top`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$topN`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1467) |
| `$toString`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1466) |
| `$toLower`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$toUpper`                        | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$trim`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1463) |
| `$trunc`                          | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1453) |
| `$tsIncrement`                    | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1464) |
| `$tsSecond`                       | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1464) |
| `$type`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1466) |
| `$unsetField`                     | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1461) |
| `$week`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$year`                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1460) |
| `$zip`                            | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1454) |

</details>

<!-- markdownlint-enable MD001 MD033 -->

## Administration commands

| Command                           | Argument / Option              | Property                  | Status | Comments                                                  |
| --------------------------------- | ------------------------------ | ------------------------- | ------ | --------------------------------------------------------- |
| `listCollections`                 |                                |                           | ✅      | Basic command is fully supported                          |
|                                   | `filter`                       |                           | ✅      |                                                           |
|                                   | `nameOnly`                     |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/301)  |
|                                   | `comment`                      |                           | ⚠️      | Ignored                                                   |
|                                   | `authorizedCollections`        |                           | ⚠️      | Ignored                                                   |
| `cloneCollectionAsCapped`         |                                |                           | ❌      |                                                           |
|                                   | `toCollection`                 |                           | ⚠️      |                                                           |
|                                   | `size`                         |                           | ⚠️      |                                                           |
|                                   | `writeConcern`                 |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `collMod`                         |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1510) |
|                                   | `index`                        |                           | ⚠️      |                                                           |
|                                   |                                | `keyPattern`              | ⚠️      |                                                           |
|                                   |                                | `name`                    | ⚠️      |                                                           |
|                                   |                                | `expireAfterSeconds`      | ⚠️      |                                                           |
|                                   |                                | `hidden`                  | ⚠️      |                                                           |
|                                   |                                | `prepareUnique`           | ⚠️      |                                                           |
|                                   |                                | `unique`                  | ⚠️      |                                                           |
|                                   | `validator`                    |                           | ⚠️      |                                                           |
|                                   |                                | `validationLevel`         | ⚠️      |                                                           |
|                                   |                                | `validationAction`        | ⚠️      |                                                           |
|                                   | `viewOn` (Views)               |                           | ⚠️      |                                                           |
|                                   | `pipeline` (Views)             |                           | ⚠️      |                                                           |
|                                   | `cappedSize`                   |                           | ⚠️      |                                                           |
|                                   | `cappedMax`                    |                           | ⚠️      |                                                           |
|                                   | `changeStreamPreAndPostImages` |                           | ⚠️      |                                                           |
| `compact`                         |                                |                           | ❌      |                                                           |
|                                   | `force`                        |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `compactStructuredEncryptionData` |                                |                           | ❌      |                                                           |
|                                   | `compactionTokens`             |                           | ⚠️      |                                                           |
| `convertToCapped`                 |                                |                           | ❌      |                                                           |
|                                   | `size`                         |                           | ⚠️      |                                                           |
|                                   | `writeConcern`                 |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `create`                          |                                |                           | ✅      | Basic command is fully supported                          |
|                                   | `capped`                       |                           | ⚠️      | Unimplemented                                             |
|                                   | `timeseries`                   |                           | ⚠️      | Unimplemented                                             |
|                                   |                                |                           | ⚠️      |                                                           |
|                                   |                                | `timeField`               | ⚠️      |                                                           |
|                                   |                                | `metaField`               | ⚠️      |                                                           |
|                                   |                                | `granularity`             | ⚠️      |                                                           |
|                                   | `expireAfterSeconds`           |                           | ⚠️      | Unimplemented                                             |
|                                   | `clusteredIndex`               |                           | ⚠️      |                                                           |
|                                   | `changeStreamPreAndPostImages` |                           | ⚠️      |                                                           |
|                                   | `autoIndexId`                  |                           | ⚠️      | Ingored                                                   |
|                                   | `size`                         |                           | ⚠️      | Unimplemented                                             |
|                                   | `max`                          |                           | ⚠️      | Unimplemented                                             |
|                                   | `storageEngine`                |                           | ⚠️      | Ingored                                                   |
|                                   | `validator`                    |                           | ⚠️      | Not implemented in PostgreSQL                             |
|                                   | `validationLevel`              |                           | ⚠️      | Unimplemented                                             |
|                                   | `validationAction`             |                           | ⚠️      | Unimplemented                                             |
|                                   | `indexOptionDefaults`          |                           | ⚠️      | Ingored                                                   |
|                                   | `viewOn`                       |                           | ⚠️      | Unimplemented                                             |
|                                   | `pipeline`                     |                           | ⚠️      | Unimplemented                                             |
|                                   | `collation`                    |                           | ⚠️      | Unimplemented                                             |
|                                   | `writeConcern`                 |                           | ⚠️      | Ingored                                                   |
|                                   | `encryptedFields`              |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      | Ingored                                                   |
| `createIndexes`                   |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1509) |
|                                   | `indexes`                      |                           | ⚠️      |                                                           |
|                                   |                                | `key`                     | ⚠️      |                                                           |
|                                   |                                | `name`                    | ⚠️      |                                                           |
|                                   |                                | `background`              | ⚠️      |                                                           |
|                                   |                                | `unique`                  | ⚠️      |                                                           |
|                                   |                                | `partialFilterExpression` | ⚠️      |                                                           |
|                                   |                                | `sparse`                  | ⚠️      |                                                           |
|                                   |                                | `expireAfterSeconds`      | ⚠️      |                                                           |
|                                   |                                | `hidden`                  | ⚠️      |                                                           |
|                                   |                                | `storageEngine`           | ⚠️      |                                                           |
|                                   |                                | `weights`                 | ⚠️      |                                                           |
|                                   |                                | `default_language`        | ⚠️      |                                                           |
|                                   |                                | `language_override`       | ⚠️      |                                                           |
|                                   |                                | `textIndexVersion`        | ⚠️      |                                                           |
|                                   |                                | `2dsphereIndexVersion`    | ⚠️      |                                                           |
|                                   |                                | `bits`                    | ⚠️      |                                                           |
|                                   |                                | `min`                     | ⚠️      |                                                           |
|                                   |                                | `max`                     | ⚠️      |                                                           |
|                                   |                                | `bucketSize`              | ⚠️      |                                                           |
|                                   |                                | `collation`               | ⚠️      |                                                           |
|                                   |                                | `wildcardProjection`      | ⚠️      |                                                           |
|                                   | `writeConcern`                 |                           | ⚠️      |                                                           |
|                                   | `commitQuorum`                 |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `currentOp`                       |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/161)  |
|                                   | `$ownOps`                      |                           | ⚠️      |                                                           |
|                                   | `$all`                         |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `drop`                            |                                |                           | ✅      | Basic command is fully supported                          |
|                                   | `writeConcern`                 |                           | ⚠️      | Ingored                                                   |
|                                   | `comment`                      |                           | ⚠️      | Ingored                                                   |
| `dropDatabase`                    |                                |                           | ✅      | Basic command is fully supported                          |
|                                   | `writeConcern`                 |                           | ⚠️      | Ingored                                                   |
|                                   | `comment`                      |                           | ⚠️      | Ingored                                                   |
| `dropConnections`                 |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1511) |
|                                   | `hostAndPort`                  |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `dropIndexes`                     |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1512) |
|                                   | `index`                        |                           | ⚠️      |                                                           |
|                                   | `writeConcern`                 |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `filemd5`                         |                                |                           | ❌      |                                                           |
| `fsync`                           |                                |                           | ❌      |                                                           |
| `fsyncUnlock`                     |                                |                           | ❌      |                                                           |
|                                   | `lock`                         |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `getDefaultRWConcern`             |                                |                           | ❌      |                                                           |
|                                   | `inMemory`                     |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `getClusterParameter`             |                                |                           | ❌      |                                                           |
| `getParameter`                    |                                |                           | ❌      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `killCursors`                     |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1514) |
|                                   | `cursors`                      |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `killOp`                          |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1515) |
|                                   | `op`                           |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `listCollections`                 |                                |                           | ✅      | Basic command is fully supported                          |
|                                   | `filter`                       |                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/1405) |
|                                   | `nameOnly`                     |                           | ⚠️      | [Issue](https://github.com/FerretDB/FerretDB/issues/301)  |
|                                   | `authorizedCollections`        |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `listDatabases`                   |                                |                           | ✅      | Basic command is fully supported                          |
|                                   | `filter`                       |                           | ✅      |                                                           |
|                                   | `nameOnly`                     |                           | ✅      |                                                           |
|                                   | `authorizedDatabases`          |                           | ⚠️      | Ingored                                                   |
|                                   | `comment`                      |                           | ⚠️      | Ingored                                                   |
| `listIndexes`                     |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/278)  |
|                                   | `cursor.batchSize`             |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `logRotate`                       |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/278)  |
|                                   | `<target>`                     |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `reIndex`                         |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1516) |
| `renameCollection`                |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1517) |
|                                   | `to`                           |                           | ⚠️      |                                                           |
|                                   | `dropTarget`                   |                           | ⚠️      |                                                           |
|                                   | `writeConcern`                 |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `rotateCertificates`              |                                |                           | ❌      |                                                           |
| `setFeatureCompatibilityVersion`  |                                |                           | ❌      |                                                           |
| `setIndexCommitQuorum`            |                                |                           | ❌      |                                                           |
|                                   | `setIndexCommitQuorum`         |                           | ⚠️      |                                                           |
|                                   | `indexNames`                   |                           | ⚠️      |                                                           |
|                                   | `commitQuorum`                 |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `setParameter`                    |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1518) |
| `setDefaultRWConcern`             |                                |                           | ❌      |                                                           |
|                                   | `defaultReadConcern`           |                           | ⚠️      |                                                           |
|                                   | `defaultWriteConcern`          |                           | ⚠️      |                                                           |
|                                   | `writeConcern`                 |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |
| `shutdown`                        |                                |                           | ❌      | [Issue](https://github.com/FerretDB/FerretDB/issues/1519) |
|                                   | `force`                        |                           | ⚠️      |                                                           |
|                                   | `timeoutSecs`                  |                           | ⚠️      |                                                           |
|                                   | `comment`                      |                           | ⚠️      |                                                           |

## Diagnostic commands

| Command              | Argument         | Status | Comments                         |
| -------------------- | ---------------- | ------ | -------------------------------- |
| `buildInfo`          |                  | ✅      | Basic command is fully supported |
| `collStats`          |                  | ✅      | Basic command is fully supported |
|                      | `collStats`      | ✅      |                                  |
|                      | `scale`          | ✅      |                                  |
| `connPoolStats`      |                  | ❌      | Unimplemented                    |
| `connectionStatus`   |                  | ✅      | Basic command is fully supported |
|                      | `showPrivileges` | ✅      |                                  |
| `dataSize`           |                  | ✅      | Basic command is fully supported |
|                      | `keyPattern`     | ⚠️      | Unimplemented                    |
|                      | `min`            | ⚠️      | Unimplemented                    |
|                      | `max`            | ⚠️      | Unimplemented                    |
|                      | `estimate`       | ⚠️      | Ignored                          |
| `dbHash`             |                  | ❌      | Unimplemented                    |
|                      | `collection`     | ⚠️      |                                  |
| `dbStats`            |                  | ✅      | Basic command is fully supported |
|                      | `scale`          | ✅      |                                  |
|                      | `freeStorage`    | ⚠️      | Unimplemented                    |
| `driverOIDTest`      |                  | ⚠️      | Unimplemented                    |
| `explain`            |                  | ✅      | Basic command is fully supported |
|                      | `verbosity`      | ⚠️      | Ignored                          |
|                      | `comment`        | ⚠️      | Unimplemented                    |
| `features`           |                  | ❌      | Unimplemented                    |
| `getCmdLineOpts`     |                  | ✅      | Basic command is fully supported |
| `getLog`             |                  | ✅      | Basic command is fully supported |
| `hostInfo`           |                  | ✅      | Basic command is fully supported |
| `_isSelf`            |                  | ❌      | Unimplemented                    |
| `listCommands`       |                  | ✅      | Basic command is fully supported |
| `lockInfo`           |                  | ❌      | Unimplemented                    |
| `netstat`            |                  | ❌      | Unimplemented                    |
| `ping`               |                  | ✅      | Basic command is fully supported |
| `profile`            |                  | ❌      | Unimplemented                    |
|                      | `slowms`         | ⚠️      |                                  |
|                      | `sampleRate`     | ⚠️      |                                  |
|                      | `filter`         | ⚠️      |                                  |
| `serverStatus`       |                  | ✅      | Basic command is fully supported |
| `shardConnPoolStats` |                  | ❌      | Unimplemented                    |
| `top`                |                  | ❌      | Unimplemented                    |
| `validate`           |                  | ❌      | Unimplemented                    |
|                      | `full`           | ⚠️      |                                  |
|                      | `repair`         | ⚠️      |                                  |
|                      | `metadata`       | ⚠️      |                                  |
| `validateDBMetadata` |                  | ❌      | Unimplemented                    |
|                      | `apiParameters`  | ⚠️      |                                  |
|                      | `db`             | ⚠️      |                                  |
|                      | `collections`    | ⚠️      |                                  |
| `whatsmyuri`         |                  | ✅      | Basic command is fully supported |
