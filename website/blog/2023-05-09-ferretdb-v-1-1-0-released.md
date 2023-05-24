---
slug: ferretdb-v-1-1-0-released
title: "FerretDB v1.1.0. Released"
authors: [alex]
description: >
  It is our pleasure to announce the release of [FerretDB](https://www.ferretdb.io/) version 1.1.0, which includes the addition of `renameCollection`, support for projection field assignments, and the `$project` pipeline aggregation stage, as well as `create` and `drop` commands in SAP HANA handler.
image: /img/blog/ferretdb-1-1-0-release.png
tags: [release]
---

![FerretDB v.1.1.0](/img/blog/ferretdb-1-1-0-release.png)

It is our pleasure to announce the release of [FerretDB](https://www.ferretdb.io/) version 1.1.0, which includes the addition of `renameCollection`, support for projection field assignments, and the `$project` pipeline aggregation stage, as well as `create` and `drop` commands in SAP HANA handler.

<!--truncate-->

Last month, in the middle of April, we released the [FerretDB 1.0. GA](https://blog.ferretdb.io/ferretdb-1-0-ga-opensource-mongodb-alternative/) to overwhelming success, which has seen us featured and mentioned in several blog posts, podcasts, webinars, and events.
Since then, we’ve had genuinely amazing support from the community as well as a host of new contributors, including [@cooljeanius](https://github.com/cooljeanius), [@j0holo](https://github.com/j0holo), [@AuruTus](https://github.com/AuruTus), [@craigpastro](https://github.com/craigpastro), [@afiskon](https://github.com/afiskon), [@syasyayas](https://github.com/syasyayas), [@raeidish](https://github.com/raeidish), [@polyal](https://github.com/polyal), and [@wqhhust](https://github.com/wqhhust).

We thank you all!
Your enthusiasm and passion for FerretDB reinforce our belief in the need for a truly open-source document database alternative to MongoDB.

While this is not a major release, we have some exciting updates and fixes for you.
Let’s find out!

## New features

In this release, we’ve added `renameCollection` command, which would enable users to rename an existing FerretDB collection.

Say you have an `inventory` collection below:

```js
{
  "_id": 1,
  name: "ABC Electronics",
  location: "123 Main Street",
  category: "Electronics",
  inventory: [
    {
      product: "Laptop",
      price: 1200,
      quantity: 10
    }
  ]
}
```

You can access the `renameCollection` command through the `db.collection.renameCollection()` method within the same database in mongosh, as shown below for a current `inventory` collection.

```js
db.inventory.renameCollection("store")
```

Note that `writeConcern`, `comment`, and `dropTarget` arguments are not currently implemented.

In addition to the currently available aggregation pipeline stages, we now support `$project` stage, which will enable you to reshape and refine the output of your queries specifying new fields, including or excluding existing fields, and also rearranging the structure of the documents.

You can include only specific fields from the output documents, such as `category` and `inventory` in a `$project` stage, as shown below.

```js
db.store.aggregate([
  {
    $project: {
      category: 1,
      inventory: 1,
    },
  },
])
```

The output document looks like this:

```sh
[
  {
    _id: 1,
    category: "Electronics",
    inventory: [{ product: "Laptop", price: 1200, quantity: 10 }],
  },
]
```

This outputs the fields specified, together with the default `_id`.
You can suppress the default `_id` by setting it as `0`.

In the new release, we have added support for field projections assignment.
With this feature, users can now specify which fields to retrieve from the database and assign new values to them in a single query.
For instance, if we have a `users` collection as shown below:

```sh
[
  {
    _id: 1,
    name: "John",
    age: 30,
    email: "john@example.com",
  },
  {
    _id: 2,
    name: "Jane",
    age: 25,
    email: "jane@example.com",
  },
]
```

Suppose we want to retrieve the documents from the `users` collection but only include the name field while assigning a new value of 'Anonymous' to it.

```js
db.users.find({}, { name: "Anonymous" })
```

The query will return:

```sh
[
  { _id: 1, name: "Anonymous" },
  { _id: 2, name: "Anonymous" },
]
```

Also, thanks to one of our contributors, [@polyal](https://github.com/polyal), we now support `create` and `drop` commands in SAP HANA handler.

## Bug fixes

In addition to the new features, we have fixed some of the discovered bugs in the previous release.
For example, in the previous release, there was a bug when using `findandModify` for `$exists` query operations and when it shouldn't allow `$upsert` on existing `_id`.

Another bug was discovered when using multiple update operators, such as `$set` and `$min` on the same document path.
Normally, it should return an error stating that there is a conflict, which should prevent the update operation, but it didn’t.
This bug has now been resolved.

Aside from that, we’ve also resolved a bug that occurs when attempting to use dot notation in sorting, especially when using a sort criteria like `{"v.foo", 1}`.

## Documentation

For those interested in contributing to FerretDB, we’ve also updated our PR guide in [CONTRIBUTING.md](https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md), with more details on squash and push, and other information related to PR management.

In our documentation, you can now discover ways to get Docker and binary executable logs from FerretDB.
[See here for more](https://docs.ferretdb.io/configuration/logging/#docker-logs).
We’ve also documented `createIndexes`, `listIndexes`, and `dropIndexes` commands and how to use them in FerretDB.

## Conclusion

Of course, there are several other changes in this release, especially community contributions, and you can find a full list of them here in the [FerretDB version 1.1.0 release notes](https://github.com/FerretDB/FerretDB/releases/tag/v1.1.0).
We appreciate every single word of support, bug discovery, code contributions, and feedback from the community.
Your continuous support showcases the strength and belief in open source.

Like always, we look forward to your feedback and comments on this new release.
So if you have any questions or find any bugs or suggestions for new features or future improvements, [please feel free to reach out to us](https://docs.ferretdb.io/#community)!
