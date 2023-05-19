---
sidebar_position: 1
---

# Aggregation operations

Here we should mention the differences between the commands, stages, and operators, and what they are all used for.

Aggregation operations involve performing various operations on a large number of data records, such as data grouping, sorting, restructuring, or modifying.
These operations pass through one or more stages, which make up a pipeline.

![aggregation stages](../../static/img/docs/aggregation-stages.jpg)

Each stage acts upon the returned documents of the previous stage, starting with the input documents.
As shown above, the documents pass through the pipeline with the result of the previous stage acting as input for the next stage, going from `$count` => `$group` => `$sort` stage.

In the pipeline, a complex query is broken down into separate stages where the record goes through a series of transformations until it finally produces the desired result.

This section of the documentation will focus on [aggregation commands](#aggregation-commands), [aggregation stages](aggregation-stages), and aggregation operators.

## Aggregation commands

Aggregation commands are top-level commands used for aggregating data, and are typically either via the `aggregate`, `count`, or `distinct` command.

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

See aggregation pipelines for more details on the various stages.

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
