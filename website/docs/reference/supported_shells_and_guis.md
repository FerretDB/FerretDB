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
|                                        | IntelliShell                            | ✅          | [Supported commands](/website/docs/reference/supported_commands.md)                              |
|                                        | PLAIN authentication mechanism          | ✅          |                                                            |
|                                        | SCRAM authentication mechanisms         | ❌          | [Issue](https://github.com/FerretDB/FerretDB/issues/2012)  |
|                                        | TLS Support                             | ✅          |                                                            |
|                                        | Import/Export JSON                      | ✅          |                                                            |
|                                        | Import/Export CSV                       | ✅          |                                                            |
|                                        | Import/Export BSON                      | ✅          |                                                            |
|                                        | Import/Export archive files             | ✅          |                                                            |
|                                        | Export SQL statements to a `.sql` file  | ✅          |                                                            |
|                                        | `dbstats`                            | ❌          | [Issue](https://github.com/FerretDB/FerretDB/issues/1346)  |
| [MongoDB Compass](https://www.mongodb.com/products/compass) |                    | ❌          |                           |
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
| [MongoDB Shell (`mongosh`)](https://www.mongodb.com/docs/mongodb-shell/)|          | ✅    | [Supported commands](/website/docs/reference/supported_commands.md) |
|                                        | PLAIN authentication mechanism          | ✅          |                                                            |
|                                        | SCRAM authentication mechanisms         | ❌          | [Issue](https://github.com/FerretDB/FerretDB/issues/2012)  |
|                                        | TLS Support                             | ✅          |                                                            |
| [Legacy `mongo` Shell](https://www.mongodb.com/docs/v5.0/reference/program/mongo/)|          | ✅    | [Supported commands](/website/docs/reference/supported_commands.md) |
|                                        | PLAIN authentication mechanism          | ✅          |                                                            |
|                                        | SCRAM authentication mechanisms         | ❌          | [Issue](https://github.com/FerretDB/FerretDB/issues/2012)  |
|                                        | TLS Support                             | ✅          |                                                            |
| [MongoDB for VS Code](https://www.mongodb.com/products/vs-code) |          | ✅    | [Supported commands](/website/docs/reference/supported_commands.md) |
|                                        | PLAIN authentication mechanism          | ✅          |                                                            |
|                                        | SCRAM authentication mechanisms         | ❌          | [Issue](https://github.com/FerretDB/FerretDB/issues/2012)  |
|                                        | TLS Support                             | ✅          |
| [DBeaver](https://dbeaver.com/docs/wiki/MongoDB/)|          | ✅    | Basic tool is fully supported |
|                                        | PLAIN authentication mechanism          | ✅          |                                                            |
|                                        | SCRAM authentication mechanisms         | ❌          | [Issue](https://github.com/FerretDB/FerretDB/issues/2012)  |
|                                        | TLS Support                             | ✅          |
|                                        | [Data export/import](https://dbeaver.com/docs/wiki/Data-transfer/)                    | ✅          |                                                            |
|                                        | [Browsing Mongo collections](https://dbeaver.com/docs/wiki/MongoDB/#browsing-mongo-collections)                    | ✅          |                                                            |
|                                        | [Executing JavaScript](https://dbeaver.com/docs/wiki/MongoDB/#executing-javascript)                    | ✅          | [Supported commands](/website/docs/reference/supported_commands.md)                                                           |
|                                        | [Executing SQL](https://dbeaver.com/docs/wiki/MongoDB/#executing-sql)                    | ✅          |                                                            |                                                  |
|                                        | Database Statistics (`dbstats`) | ❌    | [Issue](https://github.com/FerretDB/FerretDB/issues/1346) |

### [NoSQLBooster](https://nosqlbooster.com/)

### [mingo](https://mingo.io/)
