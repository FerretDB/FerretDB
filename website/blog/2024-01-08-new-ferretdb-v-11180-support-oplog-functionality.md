---
slug: new-ferretdb-v-11180-support-oplog-functionality
title: FerretDB releases v1.18.0!
authors: [alex]
description: >
  We just released FerretDB v1.18.0, which now supports `OpLog` functionality, capped collections, tailable cursors, among other new features.
image: /img/blog/ferretdb-v1.18.0.jpg
tags: [release]
---

![FerretDB releases v.1.18.0](/img/blog/ferretdb-v1.18.0.jpg)

We are starting the new year with a bang! The latest version of FerretDB v1.18.0 has just been released with support for `OpLog`, along with other exciting features.

One of the most requested features on FerretDB is the support for `OpLog`, which would make it possible for developers to build real-time applications using the Meteor framework. We made this a priority in the last quarter and are super excited that it’s now available. Along with this addition, we included support for capped collections and tailable cursors too.

Other new features in the release include support for `createUser`, `dropUser`, `dropAllUsersFromDatabase`, UsersInfo, and many others.

## New features

In this release, we’ve added **support for capped collection**, along with `max` and `size` parameters.

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

Most importantly, we released support for a basic `OpLog` (operations log) functionality. We've had many requests for this feature especially for Meteor `OpLog` tailing and we're excited to see it finally available.

:::note
At the moment, only basic `OpLog` tailing is supported. Replication is not supported yet.
:::

The functionality is not created by default. To enable it, manually create a capped collection named `oplog.rs` in the `local` database.

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

If something does not work correctly on `OpLog` functionality, [please contact us here](https://github.com/FerretDB/FerretDB/issues/new?assignees=ferretdb-bot&labels=code%2Fbug%2Cnot+ready&projects=&template=bug.yml).

The release also includes support for `createUser`, `dropUser`, `dropAllUsersFromDatabase`, and `UsersInfo` commands. These commands will be useful in enabling compatibility with more applications.

## Thanks for all the support!

None of this would be possible without all the support we’ve received in the past year from our growing community and the open source community at large. We had numerous contributors, users, partners, and well-wishers rooting for us. Last year was truly a delight for everyone at FerretDB, and we can't wait to see what the new year unfolds. We will continue our mission to enable to more enable more database backends to run MongoDB workloads.
For instance, work is still ongoing on adding support for MySQL and SAP Hana backends, and we hope to have more information for you soon.

In this release, we had two new contributors: [@yonarw](https://github.com/yonarw) and [@sachinpuranik](https://github.com/yonarw). Thank you for your support!

To everyone in the FerretDB community, Happy New Year!
