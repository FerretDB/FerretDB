---
slug: ferretdb-v-0-9-3-improved-aggregation-pipeline-support
title: "FerretDB v.0.9.3 - Improved Aggregation Pipeline Support"
authors: [alex]
description: >
  We are thrilled to announce the release of FerretDB v.0.9.3, which includes exciting new features and improvements, and we can't wait for you to try it out
image: /img/blog/ferretdb-v0.9.3.jpg
tags: [release]
---

![FerretDB v0.9.3](/img/blog/ferretdb-v0.9.3.jpg)

We are thrilled to announce the release of [FerretDB v0.9.3](https://github.com/FerretDB/FerretDB/releases/tag/v0.9.3), which includes exciting new features and improvements, and we can't wait for you to try it out.

<!--truncate-->

While this is a minor release, it is quite significant for us at [FerretDB](https://www.ferretdb.io/), because it brings us one step closer to the release of FerretDB GA, which promises a truly open-source document database with MongoDB compatibility, as well as backend engines for [PostgreSQL](https://www.postgresql.org/), [Tigris](https://www.tigrisdata.com/), with one for [SAP HANA](https://www.sap.com/products/technology-platform/hana.html) currently in development, and potentially many others.

FerretDB v0.9.3 builds on our previous release and promises better performance and database experience for FerretDB with enhanced support for new aggregation pipeline stages and more pushdown queries.

But what makes this release truly special is the active participation of the developer community in its development.
We are proud to say that the codebase has been enriched by the contributions from some of the most talented developers from across the globe, contributing their unique perspectives and expertise.
Special appreciation goes to our two new contributors, [@lucboj](https://github.com/lucboj) and [@yu-re-ka](https://github.com/yu-re-ka), for contributing to this release.
Thank you for your continued support.
Your work is greatly appreciated!

So, without further ado, let’s find out the latest changes to this new version and what you can achieve with them.

## New Features

In FerretDB v0.9.3, we’ve enabled the automatic creation of unique `_id` values for each document inserted into a collection.
This new feature ensures that you don’t have duplicate `_id` values in the collection, preventing errors that could lead to the same `_id` values being accidentally inserted twice.

In recent releases, we’ve worked on improving our support for aggregation pipeline stages and pushed more queries to the backend.
FerretDB v0.9.3 improves on this, adding more stages and pushing down more queries, ensuring significantly better database performance.

For instance, we’ve implemented `$sort` and `$group` aggregation pipeline stages.
Suppose you have a collection named `order` that contains the following documents:

```js
{
    _id: 1,
    customer_name: "John Smith",
    order_items: ["shirt", "pants", "shoes"],
    order_total: 23.50
},
{
    _id: 2,
    customer_name: "Mary Johnson",
    order_items: ["dress", "scarf"],
    order_total: 35.20
},
{
    _id: 3,
    customer_name: "Bob Williams",
    order_items: ["hat", "gloves", "socks"],
    order_total: 10.99
}
```

We can use the `$sort` aggregation pipeline stage to sort the documents in ascending or descending order based on a specific field.
For example, to sort the documents in descending order based on the `order_total` field, set the value to `-1`:

```js
db.orders.aggregate([
    { $sort: { order_total: -1 } }
])
```

This will return the documents, sorted in descending order based on the `order_total` field.

You can use the `$group` aggregation stage to group documents together.
For example, the following query groups and counts the total number of documents in the `orders` collection.

```js
db.orders.aggregate([
    {
        $group: {
            _id: null,
            count: { $count: {} }
        }
    }
])
```

As part of our mission to effectively enable more backends to run MongoDB workloads, basic aggregation pipeline support is now available for [Tigris](https://www.tigrisdata.com/).

## Bug Fixes

FerretDB v.0.9.3 has had several enhancements and bug fixes since the last release, indicating our commitment to building a reliable document database for our users.

We’ve fixed dot notation errors for `$max`, `$min`, and `$mul` update operators.
Previously, they were causing issues by failing to return errors when attempting to update paths that were not valid.
The fixes now ensure that update operation will return an error when attempting to update an invalid path.

We’ve also addressed issues with querying an embedded field in an array, especially with the `$pullAll` operator, which was preventing it from removing all instances from an array.
One of our contributors [@yu-re-ka](https://github.com/yu-re-ka) helped fixed `saslStart` for particular clients, which should help improve compatibility with [FastNetMon](https://fastnetmon.com/).

## Other Changes

As part of the changes in this release, we’ve relaxed validation rules to enable the use of collection names that start with `_` and field names with `$` characters.
We’ve also allowed pushdown for boolean, date values, and empty strings to the backend.
Thanks to the efforts of [@lucboj](https://github.com/lucboj) in [this issue](https://github.com/FerretDB/FerretDB/pull/2071), we now have an initial setup for HANA handler, which should enable FerretDB to support [SAP HANA](https://www.sap.com/products/technology-platform/hana.html).

We’ve improved our documentation to document how telemetry sharing works explicitly.
We’ve also updated our supported commands page to reflect all the latest updates and features to FerretDB.
See [here for more information](https://docs.ferretdb.io/reference/supported-commands/) about all the supported commands in FerretDB.

Please read our release note for a [complete list of changes in this release](https://github.com/FerretDB/FerretDB/releases/tag/v0.9.3).

## Looking Ahead: Beyond FerretDB v.0.9.3

We’re getting closer to the release of FerretDB GA, seeing more usages with real-world applications and companies taking advantage of FerretDB's existing capabilities for their projects and applications.
This is fantastic news for us and represents a significant stride in our mission to provide a standard and reliable document database to our users.

Once again, we thank everyone that has been a part of our story, and we look forward to your many contributions, feedback, and enthusiastic support.
Please feel free to [contact us here](https://docs.ferretdb.io/#community).
