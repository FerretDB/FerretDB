---
sidebar_position: 1
---

# Introduction

FerretDB (previously MangoDB) was founded to become the de-facto open-source substitute to MongoDB.
FerretDB is an open-source proxy, converting the MongoDB 6.0+ wire protocol queries to SQL -
using PostgreSQL as a database engine.

## Why do we need FerretDB?

MongoDB was originally an eye-opening technology for many of us developers,
empowering us to build applications faster than using relational databases.
In its early days, its ease-to-use and well-documented drivers made MongoDB one of the simplest database solutions available.
However, as time passed, MongoDB abandoned its open-source roots;
changing the license to SSPL - making it unusable for many open source and early-stage commercial projects.

Most MongoDB users do not require any advanced features offered by MongoDB;
however, they need an easy-to-use open-source database solution.
Recognizing this, FerretDB is here to fill that gap.

## Scope and current state

FerretDB will be compatible with MongoDB drivers and will strive to serve as a drop-in replacement for MongoDB 6.0+.

Currently, the project is in its early stages and welcomes all contributors.
See our [public roadmap](https://github.com/orgs/FerretDB/projects/2/views/1)
and [contributing guidelines](https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md/).

### Known differences

1. FerretDB uses the same protocol error names and codes, but the exact error messages may be different in some cases.
2. FerretDB does not support NUL (`\0`) characters in strings.
3. Database and collection names restrictions:

* name cannot start with the reserved prefix `_ferretdb_`.
* name must not include non-latin letters, spaces, dots, dollars or dashes.
* collection name length must be less or equal than 120 symbols, database name length limit is 63 symbols.
* name must not start with a number.
* database name cannot contain capital letters.

If you encounter some other difference in behavior, please [join our community](#community) to report a problem.

## Community

* Website and blog: [https://ferretdb.io](https://ferretdb.io/).
* Twitter: [@ferret_db](https://twitter.com/ferret_db).
* [Slack chat](https://join.slack.com/t/ferretdb/shared_invite/zt-zqe9hj8g-ZcMG3~5Cs5u9uuOPnZB8~A) for quick questions.
* [GitHub Discussions](https://github.com/FerretDB/FerretDB/discussions) for longer topics.
* [GitHub Issues](https://github.com/FerretDB/FerretDB/issues) for bugs and missing features.
* [Open Office House meeting](https://calendar.google.com/event?action=TEMPLATE&tmeid=NjNkdTkyN3VoNW5zdHRiaHZybXFtb2l1OWtfMjAyMTEyMTNUMTgwMDAwWiBjX24zN3RxdW9yZWlsOWIwMm0wNzQwMDA3MjQ0QGc&tmsrc=c_n37tquoreil9b02m0740007244%40group.calendar.google.com&scp=ALL)
  every Monday at 18:00 UTC at [Google Meet](https://meet.google.com/mcb-arhw-qbq).

If you want to contact FerretDB Inc., please use [this form](https://www.ferretdb.io/contact/).
