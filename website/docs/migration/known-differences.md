---
sidebar_position: 4
slug: /migration/known-differences/ # referenced in README.md and redirects
description: Supported commands
---

# Known differences

We don't plan to address those known differences in behavior:

<!--
   Each numbered point above should have a corresponding, numbered test file https://github.com/FerretDB/FerretDB/tree/main/integration/diff_*_test.go
   Bullet subpoints should be in the same file as the parent point.
-->

1. FerretDB uses the same protocol error names and codes as MongoDB,
   but the exact error messages may sometimes be different.
2. FerretDB collection names must be valid UTF-8; MongoDB allows invalid UTF-8 sequences.
   <!-- TODO https://github.com/FerretDB/FerretDB/issues/4879 -->

We consider all other differences in behavior to be problems and want to address them.
Please [join our community](/#community) to report them.

<!--
Use ❌ for features that are not implemented at all.
Use ⚠️ for features implemented with major limitations, or if they are safely ignored.
Use ✅️ otherwise.

See also https://github.com/prettier/prettier/issues/15572
-->

## Wire protocol

### Administrative commands

| Command                   | Status                                                                     |
| ------------------------- | -------------------------------------------------------------------------- |
| `cloneCollectionAsCapped` | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/3631) |
| `collMod`                 | ✅️ Supported                                                              |
| `compact`                 | ✅️ Supported                                                              |
| `convertToCapped`         | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/3631) |
| `create`                  | ✅️ Supported                                                              |
| `createIndexes`           | ✅️ Supported                                                              |
| `currentOp`               | ✅️ Supported                                                              |
| `drop`                    | ✅️ Supported                                                              |
| `dropConnections`         | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1511) |
| `dropDatabase`            | ✅️ Supported                                                              |
| `dropIndexes`             | ✅️ Supported                                                              |
| `getParameter`            | ✅️ Supported                                                              |
| `killCursors`             | ✅️ Supported                                                              |
| `killOp`                  | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1515) |
| `listCollections`         | ✅️ Supported                                                              |
| `listDatabases`           | ✅️ Supported                                                              |
| `listIndexes`             | ✅️ Supported                                                              |
| `logRotate`               | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1959) |
| `reIndex`                 | ✅️ Supported                                                              |
| `renameCollection`        | ✅️ Supported                                                              |
| `setParameter`            | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1518) |
| `shutdown`                | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1519) |

### Aggregation commands

| Command     | Status        |
| ----------- | ------------- |
| `aggregate` | ✅️ Supported |
| `count`     | ✅️ Supported |
| `distinct`  | ✅️ Supported |

### Authentication commands

| Command        | Status                                                                     |
| -------------- | -------------------------------------------------------------------------- |
| `authenticate` | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1731) |
| `logout`       | ✅️ Supported                                                              |
| `saslContinue` | ✅️ Supported                                                              |
| `saslStart`    | ✅️ Supported                                                              |

### Diagnostic commands

| Command                 | Status                                                                     |
| ----------------------- | -------------------------------------------------------------------------- |
| `buildInfo`             | ✅️ Supported                                                              |
| `collStats`             | ✅️ Supported                                                              |
| `connectionStatus`      | ✅️ Supported                                                              |
| `connPoolStats`         | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/4909) |
| `dataSize`              | ✅️ Supported                                                              |
| `dbStats`               | ✅️ Supported                                                              |
| `explain`               | ✅️ Supported                                                              |
| `ferretDebugError`      | ✅️ Supported                                                              |
| `getCmdLineOpts`        | ✅️ Supported                                                              |
| `getLog`                | ✅️ Supported                                                              |
| `hostInfo`              | ✅️ Supported                                                              |
| `listCommands`          | ✅️ Supported                                                              |
| `logApplicationMessage` | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/4969) |
| `ping`                  | ✅️ Supported                                                              |
| `profile`               | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/2398) |
| `serverStatus`          | ✅️ Supported                                                              |
| `validate`              | ✅️ Supported                                                              |
| `whatsmyuri`            | ✅️ Supported                                                              |

### Query commands

| Command         | Status                                                                     |
| --------------- | -------------------------------------------------------------------------- |
| `bulkWrite`     | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/4910) |
| `delete`        | ✅️ Supported                                                              |
| `find`          | ✅️ Supported                                                              |
| `findAndModify` | ✅️ Supported                                                              |
| `getMore`       | ✅️ Supported                                                              |
| `insert`        | ✅️ Supported                                                              |
| `update`        | ✅️ Supported                                                              |

### Role management commands

| Command                    | Status                                                                     |
| -------------------------- | -------------------------------------------------------------------------- |
| `createRole`               | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1528) |
| `dropAllRolesFromDatabase` | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1530) |
| `dropRole`                 | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1529) |
| `grantPrivilegesToRole`    | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1531) |
| `grantRolesToRole`         | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1532) |
| `revokePrivilegesFromRole` | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1534) |
| `revokeRolesFromRole`      | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1535) |
| `rolesInfo`                | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1536) |
| `updateRole`               | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1537) |

### Session commands

| Command                    | Status                                                                           |
| -------------------------- | -------------------------------------------------------------------------------- |
| `abortTransaction`         | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1547)       |
| `commitTransaction`        | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1548)       |
| `endSessions`              | ✅️ Supported                                                                    |
| `killAllSessions`          | ✅️ Supported                                                                    |
| `killAllSessionsByPattern` | [⚠️ Not fully implemented yet](https://github.com/FerretDB/FerretDB/issues/1551) |
| `killSessions`             | ✅️ Supported                                                                    |
| `refreshSessions`          | ✅️ Supported                                                                    |
| `startSession`             | ✅️ Supported                                                                    |

### User management commands

| Command                    | Status                                                                     |
| -------------------------- | -------------------------------------------------------------------------- |
| `createUser`               | ✅️ Supported                                                              |
| `dropAllUsersFromDatabase` | ✅️ Supported                                                              |
| `dropUser`                 | ✅️ Supported                                                              |
| `grantRolesToUser`         | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1494) |
| `revokeRolesFromUser`      | [❌ Not implemented yet](https://github.com/FerretDB/FerretDB/issues/1495) |
| `updateUser`               | ✅️ Supported                                                              |
| `usersInfo`                | ✅️ Supported                                                              |

## Data API

| Path                 | Status        |
| -------------------- | ------------- |
| `/action/aggregate`  | ✅️ Supported |
| `/action/deleteMany` | ✅️ Supported |
| `/action/deleteOne`  | ✅️ Supported |
| `/action/find`       | ✅️ Supported |
| `/action/findOne`    | ✅️ Supported |
| `/action/insertMany` | ✅️ Supported |
| `/action/insertOne`  | ✅️ Supported |
| `/action/updateMany` | ✅️ Supported |
| `/action/updateOne`  | ✅️ Supported |
