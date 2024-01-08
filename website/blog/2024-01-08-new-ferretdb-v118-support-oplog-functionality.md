---
slug: new-ferretdb-v-11180-support-oplog-functionality
title: FerretDB releases v1.18.0!
authors: [alex]
description: >
  We just released FerretDB v1.18.0, which now supports OpLog functionality, capped collections, tailable cursors, among other new features.
image: /img/blog/ferretdb-v1.18.0.jpg
tags: [release]
---

![FerretDB releases v.1.18.0](/img/blog/ferretdb-v1.18.0.jpg)

We are starting the new year with a major feature addition!
The latest version of FerretDB v1.18.0 has just been released with support for basic `OpLog` functionality, along with other exciting features.

`OpLog` functionality has been one of our most requested features, and it would make it possible for developers to build real-time applications using the Meteor framework.
We made this a priority in the last quarter and are super excited that it's now available.
We can't wait to see what you build with FerretDB!

In addition to `OpLog` functionality, we've also added support for capped collections and tailable cursors.
Other new features in the release include support for `createUser`, `dropUser`, `dropAllUsersFromDatabase`, and `UsersInfo`.

FerretDB is on a mission to add MongoDB compatibility to other database backends, including [Postgres](https://www.postgresql.org/) and [SQLite](https://www.sqlite.org/).
All the new features in this release will help us improve compatibility with more applications，and use cases.

## New features

In this release, we've added **support for capped collection**, along with `max` and `size` parameters.

For example, to create a capped collection `testcollection` with a maximum size of 512MB, run:

```js
db.createCollection('testcollection', { capped: true, size: 536870912 })
```

This release also includes **support for tailable cursors**; both `tailable` and `awaitData` parameters are included.

To add a tailable cursor, run:

```js
db.collection.find(
  { $query: {}, $orderby: { $natural: 1 } },
  { tailable: true, awaitData: true }
)
```

Most importantly, we released support for a basic `OpLog` (operations log) functionality.
We've had many requests for this feature especially for Meteor `OpLog` tailing and we're thrilled to have it finally available for our users.

:::note
At the moment, only basic `OpLog` tailing is supported.
Replication is not supported yet.
:::

The functionality is not created by default.
To enable it, manually create a capped collection named `oplog.rs` in the `local` database.

```js
use local
db.createCollection("oplog.rs", {capped: true, size: 536870912})
```

You may also need to set the replica set name using [`--repl-set-name` flag / `FERRETDB_REPL_SET_NAME` environment variable](https://docs.ferretdb.io/configuration/flags/#general).

To query the `OpLog`:

```js
db.oplog.rs.find()
```

To query `OpLog` for all the operations in a particular namespace (`test.foo`), run:

```js
db.oplog.rs.find({ ns: 'test.foo' })
```

If something does not work correctly or you have any question on the `OpLog` functionality, [please inform us here](https://github.com/FerretDB/FerretDB/issues/new?assignees=ferretdb-bot&labels=code%2Fbug%2Cnot+ready&projects=&template=bug.yml).

## Thanks for all the support!

None of this would be possible without all the support we've received in the past year from our growing community and the open source community at large.
In this release, we had two new contributors: [@yonarw](https://github.com/yonarw) and [@sachinpuranik](https://github.com/yonarw).
Thank you for your support!

Last year was an amazing one for everyone at FerretDB, and we can't wait to see what the new year unfolds.
We will continue on our goal to provide an open source MongoDB alternative that enables more database backends run MongoDB workloads.
Work is still ongoing on adding support for [MySQL](https://www.mysql.com/) and [SAP Hana](https://www.sap.com/products/technology-platform/hana.html) backends, and we hope to have more information for you soon.

If you have questions or feedback on FerretDB, [contact us on our community channels](https://docs.ferretdb.io/#community).
