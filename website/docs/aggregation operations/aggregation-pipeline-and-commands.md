---
sidebar_position: 1
slug: /aggregation-pipeline-and-commands/
---

# Aggregation pipeline and commands

Aggregation operations involve performing various operations on a large number of data records, such as data grouping, sorting, restructuring, or modifying.
These operations pass through one or more stages, which make up a pipeline.

![aggregation stages](../../static/img/docs/aggregation-stages.jpg)

Each stage acts upon the returned documents of the previous stage, starting with the input documents.
As shown above, the documents pass through the pipeline with the result of the previous stage acting as input for the next stage, going from `$match` => `$group` => `$sort` stage.

For example, say you have the following documents in a `sales` collection:

```js
[
  { "_id": 1, "category": "Electronics", "price": 1000 },
  { "_id": 2, "category": "Electronics", "price": 800 },
  { "_id": 3, "category": "Clothing", "price": 30 },
  { "_id": 4, "category": "Clothing", "price": 50 },
  { "_id": 5, "category": "Home", "price": 1500 },
  { "_id": 6, "category": "Home", "price": 1200 },
  { "_id": 7, "category": "Books", "price": 20 },
  { "_id": 8, "category": "Books", "price": 40 }
]
```

A typical aggregation pipeline would look like this:

```js
db.sales.aggregate([
  { $match: { category: { $ne: "Electronics" } } },
  {
    $group: {
      _id: "$category",
      totalPrice: { $sum: "$price" },
      productCount: { $sum: 1 }
    }
  },
  { $sort: { totalPrice: -1 } }
])
```

In the pipeline, the complex query is broken down into separate stages where the record goes through a series of transformations until it finally produces the desired result.
First, the `$match` stage filters out all documents where the `category` field is not `Electronics`.
Then, the `$group` stage groups the documents by their `category` and calculates the total price and product count for each of those category.
Finally, the `$sort` stage sorts the documents by the `totalPrice` field in descending order.

So the above aggregation pipeline operation would return the following result:

```sh
[
  { _id: 'Home', totalPrice: 2700, productCount: 2 },
  { _id: 'Clothing', totalPrice: 80, productCount: 2 },
  { _id: 'Books', totalPrice: 60, productCount: 2 }
]
```

This section of the documentation will focus on [aggregation commands](#aggregation-commands), [aggregation stages](../aggregation-stages), and aggregation operators.

## Aggregation commands

Aggregation commands are top-level commands used for aggregating data, and are typically either via the `aggregate`, `count`, or `distinct` command.

:::note
Aggregation pipeline stages are only available for the `aggregate` command, and not for the `count` or `distinct` command.
:::

### `aggregate`

The `aggregate` command is used for performing aggregation operations on a collection.
It lets you specify aggregation operations in a pipeline consisting of one or more stages and operators for transforming and analyzing data, such as grouping, filtering, sorting, projecting, and calculating aggregates.

```js
// Aggregation pipeline to perform aggregation operations on a collection
db.collection.aggregate([
  // Stage 1: Matching documents based on a specific field and value
  { $match: { field: value } },
  // Stage 2: Grouping documents by the "category" field and calculating the sum of the "quantity" field
  { $group: { _id: "$category", total: { $sum: "$quantity" } } }
])
```

### `count`

The `count` command displays the number of documents returned by a specific query.
Returns a `count` as the result.

```js
db.collection.count({ field: value })
```

### `distinct`

The `distinct` command returns unique values for a specified field in a collection.
Returns an array of distinct values for the specified field.

```js
db.collection.distinct("field")
```
