---
sidebar_position: 6
slug: /supported_shells_and_guis/
---

# Supported Shells and GUIs

<!--
    blah blah blah
-->

| Shell / GUI                            | Feature                                 | Status      | Comments                                                   |
| -------------------------------------- | --------------------------------------- | ----------- | ---------------------------------------------------------- |
| [Studio 3T](https://studio3t.com/)     |                                         | ✅          | Basic tool is fully supported                              |
|                                        | Connecting to FerretDB                  | ✅          |                                                            |
|                                        | IntelliShell                            | ✅          | [Available commands](/website/docs/reference/supported_commands.md)                              |
|                                        | PLAIN authentication mechanism          | ✅          |                                                            |
|                                        | SCRAM authentication mechanisms         | ❌          | [Issue](https://github.com/FerretDB/FerretDB/issues/2012)  |
|                                        | TLS Support                             | ✅          |                                                            |
|                                        | Import/Export JSON                      | ✅          |                                                            |
|                                        | Import/Export CSV                       | ✅          |                                                            |
|                                        | Import/Export BSON                      | ✅          |                                                            |
|                                        | Import/Export archive files             | ✅          |                                                            |
|                                        | Export SQL statements to a `.sql` file  | ✅          |                                                            |
|                                        | `collStats`                             | ❌          | [Issue](https://github.com/FerretDB/FerretDB/issues/1346)  |
| [MongoDB Compass](https://www.mongodb.com/products/compass) |                    | ❌          | Basic tool is fully supported                              |
|                                        | Connecting to FerretDB                  | ✅          |                                                            |
|                                        | PLAIN authentication mechanism          | ✅          |                                                            |
|                                        | SCRAM authentication mechanisms         | ❌          | [Issue](https://github.com/FerretDB/FerretDB/issues/2012)  |
|                                        | TLS Support                             | ✅          |                                                            |
|                                        | [View Documents](https://www.mongodb.com/docs/compass/current/documents/view/)        | ❌    | [Issue](https://github.com/FerretDB/FerretDB/issues/1346) |
|                                        | [Insert Documents](https://www.mongodb.com/docs/compass/current/documents/insert/)    | ❌    | [Issue](https://github.com/FerretDB/FerretDB/issues/1346) |                                                                   |
|                                        | [Clone Documents](https://www.mongodb.com/docs/compass/current/documents/clone/)      | ❌    | [Issue](https://github.com/FerretDB/FerretDB/issues/1346) |
|                                        | [Delete Documents](https://www.mongodb.com/docs/compass/current/documents/delete/)    | ❌    | [Issue](https://github.com/FerretDB/FerretDB/issues/1346) |
|                                        | [Query Your Data](https://www.mongodb.com/docs/compass/current/query/filter/)         | ❌    | [Issue](https://github.com/FerretDB/FerretDB/issues/1346) |
|                                        | [Import and Export Data](https://www.mongodb.com/docs/compass/current/import-export/) | ❌    | [Issue](https://github.com/FerretDB/FerretDB/issues/1346) |
|                                        | [Embedded MongoDB Shell](https://www.mongodb.com/docs/compass/current/embedded-shell/)| ✅    |                                                           |
|                                        | [Aggregation Pipeline Builder](https://www.mongodb.com/docs/compass/current/aggregation-pipeline-builder/)| ❌    | [Issue](https://github.com/FerretDB/FerretDB/issues/9) |

### [MongoDB for VS Code](https://www.mongodb.com/products/vs-code)

### [MongoDB Shell (`mongosh`)](https://www.mongodb.com/docs/mongodb-shell/)

### [Legacy `mongo` Shell](https://www.mongodb.com/docs/v5.0/reference/program/mongo/)

### [DBeaver](https://dbeaver.com/docs/wiki/MongoDB/)

### [NoSQLBooster](https://nosqlbooster.com/)

### [mingo](https://mingo.io/)
