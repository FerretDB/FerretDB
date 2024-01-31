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

In recent weeks, we've been working on a new authentication which would make it possible to support `SCRAM-SHA-256` mechanism.
This is a highly requested feature from several potential users eager to use FerretDB in their applications.
Once this is in place we'll be able to support more MongoDB workloads and use cases.

Our goal is to make FerretDB the truly open source MongoDB alternative, and performance is also a key part of that.
What we are working on should result in significant performance improvement for FerretDB in future releases.

## New features

In FerretDB v1.19.0, we've enabled support for index creation on nested fields for SQLite.
Say you have a collection `supply`, you can create an index on a nested field such as `item.name` for the SQLite backend:

```json5
db.supply.createIndex({ "item.name": 1 });
```

This should create an ascending index on the `name` field nested within the `item` object.
Enabling support for this feature on the SQLite backend should offer faster query performance on nested document structures.

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

Many of the changes in recent releases focus on enabling support for new authentication mechanisms, such as `SCRAM-SHA-256`.
Among other things needed to make this possible, FerretDB should manage users by itself and support more user management commands.
So far, we've added basic support for the following user management commands: `createUser`, `dropAllUsersFromDatabase`, `dropUser`, `updateUser`, and `usersInfo`.

For now, they are only accessible for testing purposes by running FerretDB with the `EnableNewAuth` flag/`FERRETDB_TEST_ENABLE_NEW_AUTH` environment variable.

Then you can create a user with the `createUser` command:

```text
ferretdb> db.runCommand({ createUser: "user", pwd: "password", roles: [] });
{ ok: 1 }
ferretdb> db.runCommand({"usersInfo":1})
{
  users: [
    {
      _id: 'ferretdb.user',
      user: 'user',
      db: 'ferretdb',
      roles: [],
      userId: UUID('26e8c6fa-46b1-4f3f-9754-fbfa7a2bf4b8')
    }
  ],
  ok: 1
}
ferretdb> db.dropUser("user");
{ ok: 1 }
```

For other changes in this release, see the [release note here](https://github.com/FerretDB/FerretDB/releases/tag/v1.19.0).

Many thanks to all our contributors, your support means a lot to us, and we value it greatly.
We had many open-source contributors to this release, with [@fadyat](https://github.com/fadyat) making a first contribution – thank you!

If you have any questions or comments about this release or FerretDB, [reach out to us on any of our channels](https://docs.ferretdb.io/#community).
