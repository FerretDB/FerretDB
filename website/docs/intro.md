---
sidebar_position: 1
---

# Introduction

FerretDB (formerly MangoDB) is an open-source proxy that translates MongoDB wire protocol
queries to SQL, with PostgreSQL as the database engine.

Initially built as open-source software, MongoDB was a game-changer for many developers,
enabling them to build fast and robust applications.
Its ease of use and extensive documentation made it a top choice for many developers looking
for an open-source database.
However, all this changed when they switched to an SSPL license,
moving away from their open-source roots.

In light of this, FerretDB was founded to become the true open-source alternative to MongoDB,
making it the go-to choice for most MongoDB users looking for an open-source alternative to MongoDB.
With FerretDB, users can run the same MongoDB protocol queries without needing to learn a new language or command.

## Scope and current state

FerretDB is compatible with MongoDB drivers and can be used as a direct replacement for MongoDB 6.0+.
Currently, the project is still in its early stages of development and not yet suitable for use in production-ready environments. 
See our [public roadmap](https://github.com/orgs/FerretDB/projects/2/views/1)
and [contributing guidelines](CONTRIBUTING.md).

## Known differences

1. FerretDB uses the same protocol error names and codes, but the exact error messages may be different in some cases.
2. FerretDB does not support NULL (\0) characters in strings.
3. For Tigris, FerretDB requires Tigris schema validation for `msg_create`: validator must be set as `$tigrisSchemaString`.
The value must be a JSON string representing JSON schema in [Tigris format](https://docs.tigrisdata.com/overview/schema).

4. Database and collection names restrictions:

* name cannot start with the reserved prefix `_ferretdb_`.
* name must not include non-latin letters, spaces, dots, dollars or dashes.
* collection name length must be less or equal than 120 symbols, database name length limit is 63 symbols.
* name must not start with a number.
* database name cannot contain capital letters.

If you encounter some other difference in behavior, please [join our community](https://github.com/FerretDB/FerretDB#community) to report the problem.

## Community

* Website and blog: [https://ferretdb.io](https://ferretdb.io/).
* Twitter: [@ferret_db](https://twitter.com/ferret_db).
* [Slack chat](https://join.slack.com/t/ferretdb/shared_invite/zt-zqe9hj8g-ZcMG3~5Cs5u9uuOPnZB8~A) for quick questions.
* [GitHub Discussions](https://github.com/FerretDB/FerretDB/discussions) for longer topics.
* [GitHub Issues](https://github.com/FerretDB/FerretDB/issues) for bugs and missing features.
* [Open Office House meeting](https://calendar.google.com/event?action=TEMPLATE&tmeid=NjNkdTkyN3VoNW5zdHRiaHZybXFtb2l1OWtfMjAyMTEyMTNUMTgwMDAwWiBjX24zN3RxdW9yZWlsOWIwMm0wNzQwMDA3MjQ0QGc&tmsrc=c_n37tquoreil9b02m0740007244%40group.calendar.google.com&scp=ALL)
  every Monday at 18:00 UTC at [Google Meet](https://meet.google.com/mcb-arhw-qbq).

If you want to contact FerretDB Inc., please use [this form](https://www.ferretdb.io/contact/).
