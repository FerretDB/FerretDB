---
title: "FerretDB v0.9.0 - The Developer Preview"
slug: ferretdb-v-0-9-0-developer-preview
author: Alexander Fashakin
description: "FerretDB 0.9.0 brings with it amazing new features, especially the initial support for aggregation pipelines."
image: /img/blog/developer_preview.jpg
tags: [release]
date: 2023-01-31
---

FerretDB 0.9.0 brings with it amazing new features, especially the initial support for aggregation pipelines.

![FerretDB v0.9.0 - The Developer Preview](/img/blog/developer_preview.jpg)

<!--truncate-->

We just rolled out our first Developer Preview - FerretDB v0.9.0, and we are so excited to show you all the exciting updates now available for you.
This version of [FerretDB](https://www.ferretdb.io) brings us closer to our mission of providing you with a true open-source alternative to MongoDB that supports more real-world use cases.

All this wouldn't be possible without the support of our community and everyone who contributed immensely to this release through their feedback, code, and support.
Special appreciation goes to [@jkoenig134](https://github.com/jkoenig134) for their first contribution to FerretDB.

## What's new

In this Developer Preview, we've added initial support for the aggregation pipelines.
Currently, only the `$match` and `$count` stages have been implemented.

This `$match` stage is used to filter documents that match specified query conditions.

```js
db.collection.aggregate([
    {
        $match: {
            <query>
        }
    }
])
```

On the other hand, the `$count` stage returns the number of documents in the collection that match the specified query conditions.

```js
db.collection.aggregate([
    {
        $count: {
            <query>
        }
    }
])
```

Keep an eye out for other aggregation pipeline stages in future releases.

FerretDB v0.9.0 now includes support for the `$mul` field update operator.
With this operator, you can perform multiplications on specific fields in your documents.
It takes two arguments: the `field` to update and the `value` to multiply the `field` with.

```js
db.collection.update(
    {<query>},
    {
        $mul: {
            <field>: <value>
        }
    }
)
```

Besides that, FerretDB v0.9.0 now supports `$push` array update operator so that you can add elements to the end of an array field in a document.
Here's how you can use it:

```js
db.collection.update(
    {<query>},
    {
        $push: {
            <fieldName>: <value>
        }
    }
)
```

In the example above, `fieldName` represents the name of the array field and `value` represents the value you want to add to the array.
You can add multiple values at once by specifying an array of values.

Also included in the latest release is the support for pushdown for numbers in PostgreSQL.
Pushdown enables quicker queries by retrieving only the necessary data, rather than obtaining all the data.

## Bug Fixes

In this release, we've fixed a few pesky bugs causing issues for our users.
One of these bugs caused wrong error types to be returned when using dot notation with the `$set` operator.

Another bug causing problems for our users was the `$inc` operator panicking for non-existing array indices; that's now fixed.

Finally, we have fixed the `$set` operator to correctly apply comparisons.
Previously, the modified count was not correctly updated when changing to the same value.

## Documentation

We've updated [our documentation](https://docs.ferretdb.io) to be more user-friendly, and this includes routing users directly to the documentation page rather than a landing page.
Besides that, we've also updated our [docker deployment guide](https://docs.ferretdb.io/quickstart_guide/docker/) to be as up-to-date as possible.

For users interested in contributing to our documentation, this new release introduces a [documentation writing guide](https://docs.ferretdb.io/contributing/writing-guide/) which should make it easier to get started.

For other changes in this new release, please see the [release log](https://github.com/FerretDB/FerretDB/releases/tag/v0.9.0).

Our sincere thanks go to all our users, partners, and the entire community for their unwavering support and contributions to FerretDB.

As we keep improving FerretDB to cover all aspects of our users' needs, we hope you can continue providing us with support and feedback for further improvements.

For more information on FerretDB, please [contact us](https://docs.ferretdb.io/#community).
