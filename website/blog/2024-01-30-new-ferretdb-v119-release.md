---
slug: new-ferretdb-v119-release
title: FerretDB releases v1.19.0
authors: [alex]
description: >
  We just released FerretDB v1.19.0, which includes support for creating an index on nested fields in SQLite and some important bug fixes.
image: /img/blog/ferretdb-v1.19.0.jpg
tags: [release]
---

![FerretDB releases v1.19.0](/img/blog/ferretdb-v1.19.0.jpg)

We just released FerretDB v1.19.0, which includes support for creating an index on nested fields in SQLite and some important bug fixes.

In recent weeks, we've been working on improving authentication and enabling support for more user management commands.

Our goal is to make FerretDB the truly open source MongoDB alternative, and performance is also a key part of that.
We have been focusing on improving the performance of FerretDB, and should have some exciting updates and improvements in future releases.

## New features

In FerretDB v1.19.0, we've enabled support for index creation on nested fields for SQLite.
Say you have a collection `supply`, you can create an index on a nested field such as `item.name` for the SQLite backend:

```json5
db.supply.createIndex({ "item.name": 1 });
```

This should create an ascending index on the `name` field nested within the `item` object.

## Bug fixes

This release fixes some of the bugs in previous versions.
For instance, we've fixed the issue with `maxTimeMS` for `getMore` command, which is important for tailable cursors.
`upsert` with `$setOnInsert` was also not working, and that has been fixed.

We've also fixed the validation process for creating duplicate `_id` index.
Normally, such operations should not cause an "Index already exists with a different name" and we're glad that has been resolved.

## Documentation update

In [our last release (v1.18.0)](https://blog.ferretdb.io/new-ferretdb-v118-support-oplog-functionality/), we added support for OpLog functionality, and we documented its usage and how you can set it up for your application.
With the addition of OpLog, users can start building real-time applications with FerretDB using the Meteor framework – [See the docs here](https://docs.ferretdb.io/configuration/oplog-support/).

We also fixed an issue with search queries on the documentation.
Before, our documentation search wasn't able to handle queries that involved operators like `$eq`, and this is finally fixed.

## Other changes

Many of the recent changes have gone into enabling better support for authentication on FerretDB.
To make that possible, we've recently added support for more user management commands, such as the `updateUser` command.

For other changes in this release, see the [release note here](https://github.com/FerretDB/FerretDB/releases/tag/v1.19.0).

Many thanks to all our contributors, your support means a lot to us, and we value it greatly.
We had many open-source contributors to this release, with [@fadyat](https://github.com/fadyat) making a first contribution – thank you!

If you have any questions or comments about this release or FerretDB, [reach out to us on any of our channels](https://docs.ferretdb.io/#community).
