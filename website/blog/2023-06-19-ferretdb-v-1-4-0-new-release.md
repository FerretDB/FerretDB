---
slug: ferretdb-v-1-4-0-new-release
title: FerretDB v1.4.0 New Release
authors: [alex]
description: >
  We're delighted to announce the latest FerretDB release – v1.4.0, which now includes options for creating unique indexes, and support for more aggregation pipeline stages, among other interesting updates.
image: /img/blog/ferretdb-v1.4.0.jpg
tags: [release]
---

![FerretDB v.1.4.0](/img/blog/ferretdb-v1.4.0.jpg)

We're delighted to announce the latest [FerretDB](https://www.ferretdb.io/) release – v1.4.0, which now includes options for creating unique indexes, and support for more aggregation pipeline stages, among other interesting updates.

<!--truncate-->

Building on the [list of currently supported commands](https://docs.ferretdb.io/reference/supported-commands/), we've been working to improve compatibility and enable more real-world applications to use FerretDB as their open-source MongoDB alternative.
To facilitate this, FerretDB has seen several feature updates, enhancements, and bug fixes since the [first GA announcement](https://blog.ferretdb.io/ferretdb-1-0-ga-opensource-mongodb-alternative/).

All of these have led to a great deal of interest in FerretDB from different communities, most importantly the open source community and contributors willing to help us with these improvements; for example, one of our newest contributors, [@shibasisp](https://github.com/shibasisp), was instrumental in implementing the `$unset`, `$set`, and `$addField` aggregation pipeline stages in this release.
We thank everyone that has contributed to FerretDB – you've all been amazing!

### New Features

One of the notable features in this release is the implementation of the `createIndexes` command to support unique indexes; this should help ensure uniqueness in your indexes so that no two indexed fields have the same values.

You can create a unique index by setting the `unique` option in the `createIndexes` command as `true`.

```js
db.collection.createIndexes({ indexedfield: 1 }, { unique: `true` })
```

Read more about [unique indexes in our documentation](https://docs.ferretdb.io/indexes/#unique-indexes).

Beyond this, we've added `$type` operator support in aggregation `$project` stage.
You can use the `$type` operator in a `$project` stage to return the BSON data type of a specified field in the documents.

Suppose you have the following document:

```json5
[
  {
    _id: 1,
    name: 'John',
    age: 35,
    salary: 5000
  },
  {
    _id: 2,
    name: 'Robert',
    age: 42,
    salary: 7000
  }
]
```

To return the BSON type of the `age` field in the documents, you can use the following query:

```js
db.employees.aggregate([
  {
    $project: {
      name: 1,
      age: 1,
      ageType: { $type: '$age' }
    }
  }
])
```

The output will be:

```json5
[
  { _id: 1, name: 'John', age: 35, ageType: 'int' },
  { _id: 2, name: 'Robert', age: 42, ageType: 'int' }
]
```

In addition to already supported aggregation pipeline stages, we have now added support for the `$unset`, `$set`, and `$addFields` stages.
The `$set` stage modifies the values of existing fields, while the`$addFields` stage adds new fields or updates the values of existing ones in the aggregation pipeline; the `$unset` stage excludes or removes specific fields from documents passed to the next stage of the aggregation process.

:::tip
The `$set` stage can only update the values of existing fields, while the `$addFields` stage can add new fields or update the values of existing ones in a document.
:::

Use the `$addFields` stage to add the `department` and `employmentType` fields to all documents:

```js
db.employees.aggregate([
  {
    $addFields: {
      department: 'HR',
      employmentType: 'Full-time'
    }
  }
])
```

The output will be:

```json5
[
  {
    _id: 1,
    name: 'John',
    age: 35,
    salary: 5000,
    department: 'HR',
    employmentType: 'Full-time'
  },
  {
    _id: 2,
    name: 'Robert',
    age: 42,
    salary: 7000,
    department: 'HR',
    employmentType: 'Full-time'
  }
]
```

Let's use the `$set` stage to update the `department` field:

```js
db.employees.aggregate([
  {
    $set: {
      department: 'Sales'
    }
  }
])
```

The output:

```json5
[
  { _id: 1, name: 'John', age: 35, salary: 5000, department: 'Sales' },
  {
    _id: 2,
    name: 'Robert',
    age: 42,
    salary: 7000,
    department: 'Sales'
  }
]
```

Use the `$unset` stage to remove the `salary` field from documents passed to the next stage in a pipeline:

```js
db.employees.aggregate([
  {
    $unset: 'salary'
  }
])
```

Output:

```json5
[
  { _id: 1, name: 'John', age: 35 },
  { _id: 2, name: 'Robert', age: 42 }
]
```

### Other Changes

To improve our documentation and content, we've included textlint rules for en-dashes and em-dashes so we don't wrongly document them.
We've also updated our contribution guidelines to include description of our current test naming conventions – [have a look at that here](https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md#integration-tests-naming-guidelines).
To learn about other changes in this release, please see the [FerretDB release notes for v1.4.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.4.0).

We're always looking to improve the features of FerretDB, so any suggestions, feedback, or questions you might have would be greatly appreciated.
[Reach out to us here](https://docs.ferretdb.io/#community).

[**Be a part of the FerretDB community - check us out on GitHub**](https://github.com/FerretDB/FerretDB/).
