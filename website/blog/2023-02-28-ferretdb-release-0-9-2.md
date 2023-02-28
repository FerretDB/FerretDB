---
title: "FerretDB v0.9.2 - Tigris Authentication support"
slug: ferretdb-v-0-9-2
authors: [alex]
description: "We are delighted to announce the release of FerretDB version 0.9.2, which includes Tigris authentication, support for `$listIndexes`, `$addtoSet`, and more."
image: /img/blog/ferretdb_v0.9.2.jpg
tags: [release]
date: 2023-02-28
---

We are delighted to announce the release of FerretDB version 0.9.2, which includes Tigris authentication, support for `listIndexes`, `$addtoSet`, and more.

![FerretDB v0.9.2 - Minor Release](/img/blog/ferretdb_v0.9.2.jpg)

<!--truncate-->

The past few weeks were exciting: our cooperation with the open source community is stronger than ever.
The feedback of developers and service providers makes us confident that we are on the right track with creating a truly open source alternative to MongoDB.

Before we go into all the exciting new features in this release, let’s recap FerretDB and what we’re striving to achieve.
[FerretDB](https://www.ferretdb.io/) is an [open-source alternative to MongoDB](https://blog.ferretdb.io/5-database-alternatives-mongodb-2023/) that converts MongoDB drivers and protocols using PostgreSQL and Tigris as the backend.

We are gradually increasing compatibility with MongoDB – including tools, drivers, and the entire ecosystem, so FerretDB can act as a fully open source drop-in replacement for it.
Be sure to try it out!

Here are some of the notable changes in our previous releases:

* Basic support for aggregation pipelines
* More filter queries pushed down to the backend for Tigris and PostgreSQL
* Client TLS certificate validation
* Support for raspberry pi
* Basic telemetry service

This above list is by no means exhaustive; it’s only a snippet of some of the recent additions.
If you missed our previous releases, see the [complete list of all the updates](https://github.com/FerretDB/FerretDB/releases/) to FerretDB.
We’ll be adding lots of new features as we approach the FerretDB GA.
If you have a feature request, feedback, or discovered a bug, please don’t hesitate to [contact us](https://docs.ferretdb.io/#community).
We’d love to hear from you!

Here are the latest updates on this new version of FerretDB:

## New Features

Coming into this release is support for the `listIndexes` command.
With this command, you can now access a collection’s indexes.

Besides `listindexes` command, we’ve also added support for `$addToSet` update operator, meaning you can easily add elements to an array as long as the element does not previously exist.
Let’s take a look at an example:

```js
db.states.insertOne({
  _id: 1,
  name: "California",
  colors: ["blue", "gold", "red"]
})
```

Use the `$addtoSet` operator to update the array with more elements.

```js
db.states.updateOne(
  { _id: 1 },
  { $addToSet: { colors: "green" } }
)
```

If you try updating the array with an already existing element, there won't be any changes.

```js
db.states.updateOne(
  { _id: 1 },
  { $addToSet: { colors: "red" } }
)
```

This differentiates the `$addtoSet` operator from the `$push` operator which adds the element to the array whether it exists or not.

Another new feature implemented in this release is the support for `$pullAll`.
This features updates an array by removing all entries of the specified element in the specified array.

```js
db.tested.insertOne( { _id: 1, a: [ 1, 0, 2, 3, 3, 5, 0 ] } )
```

Suppose we have the above document and want to update it using the `$pullAll` operator:

```js
db.tested.updateOne(
    { _id: 1 },
    { $pullAll: { scores: [ 0, 3 ] } }
)
```

The document is subsequently purged of all instances of the specified values from the existing array.

Another great addition to our release party is the implementation of authentication for Tigris.

All the newly added features showcase FerretDB’s dedication to creating a database that’s compatible with a wide variety of use cases.

## Bug fixes

In addition to the new releases, we’ve identified and fixed some bugs found in previous FerretDB releases.
For example, we found that Java drivers require support for setting `capped:false` on the `create` command.
See our [release notes](https://github.com/FerretDB/FerretDB/releases/tag/v0.9.2) to find out more.

## Documentation

An important update to our documentation is setting up analytics for our blog and documentation.
Having analytics set up enables us to identify potential areas for improvement and deliver top-notch content and user experience that resonates with our users on our documentation and blog.

In other changes, we’ve added more instructions explaining how evaluation and bitwise query operators work on FerretDB and making extra changes to the overall documentation content.
To learn more about all the changes in the new release, please see [our release notes](https://github.com/FerretDB/FerretDB/releases/tag/v0.9.2).

## What’s next?

In the lead-up to the FerretDB GA release, watch out for more new features as we work on adding more backend filtering queries, aggregation pipeline stages, and enabling more compatibility with real-world use cases.

At FerretDB, we welcome all contributions – code, documentation, blog, etc.
Also, every bit of feedback is invaluable to us, so please feel free to reach out to us on [any of our channels](https://docs.ferretdb.io/#community).
