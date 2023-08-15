---
slug: mongodb-sorting-scalar
title: How MongoDB sorting works for scalar values
authors: [chi]
description: >
  In this blog post, we explore in detail how MongoDB sorting works for scalar values.
keywords:
  - MongoDB sorting
  - BSON comparison order
image: /img/blog/post-cover-image.jpg
tags: [community, product, tutorial]
draft: true
---

![MongoDB sorting works for scalar values](/img/blog/mongodb-sorting-scalar.jpg)

In this blog post, we explore in detail how MongoDB sorting works for scalar values.

<!--truncate-->

Sorting compares BSON values to determine which value is equal, greater or less than the other to order them in the ascending or descending order.
For comparing different BSON types, [BSON comparison order](#bson-comparison-order) is used.

## BSON comparison order

The BSON comparison order is predefined and used to compare different BSON values.
If two BSON values share the same BSON type, their values are compared to determine which one is greater.
However, if the BSON type are different, a predefined BSON comparison order is used to handle such cases.
Below table shows the predefined BSON comparison order.

<!-- use newline in column header for appropriate spacing of columns -->
<!-- markdownlint-disable MD033 -->

| Order of Comparison<br/>(lowest to highest) | BSON Types                                   |
| ------------------------------------------- | -------------------------------------------- |
| 1                                           | Null                                         |
| 2                                           | Numbers (Integers, Longs, Doubles, Decimals) |
| 3                                           | String                                       |
| 4                                           | Object                                       |
| 5                                           | Array                                        |
| 6                                           | BinData                                      |
| 7                                           | ObjectId                                     |
| 8                                           | Boolean                                      |
| 9                                           | Date                                         |
| 10                                          | Timestamp                                    |
| 11                                          | Regular Expression                           |

## How does comparison work for different BSON type?

For comparing different BSON types, each BSON type has predefined order of comparison is assigned.
Null BSON type has the lowest order of comparison, so Null is the lowest BSON value.
When you compare Null with String, Null is less than any String.

So comparing different BSON type is merely looking up the order of comparison table for each BSON type and using predefined order.
There are exceptions for Object and Array values, we will be discussing them in another blog post.

## How does comparison work for numbers?

Although numbers have different BSON types Integers, Longs, Doubles and Decimals, they are considered equivalent for the purpose of comparison.
That means same value of any BSON number type is equivalent to same value of other BSON number type.

## Sorting examples

Suppose you have the following collection:

```js
db.outfits.insertMany([
  { _id: 1, name: 'T-shirt', size: 'M' },
  { _id: 2, name: 'jeans', size: 32 },
  { _id: 3, name: 'shorts', size: 34 },
  { _id: 4, name: 'belt', size: null }
])
```

To sort the collection in ascending order by `size` field, following query is run.

```js
db.outfits.find().sort({ size: 1 })
```

The output is sorted by the lowest BSON type Null first, then next lowest BSON type Number and finally by String.

```json5
[
  { _id: 4, name: 'belt', size: null },
  { _id: 2, name: 'jeans', size: 32 },
  { _id: 3, name: 'shorts', size: 34 },
  { _id: 1, name: 'T-shirt', size: 'M' }
]
```

Similarly, the collection is sorted in descending order by `size` field by running following query.

```js
db.outfits.find().sort({ size: -1 })
```

This time, the output is sorted by the higher BSON type String, then by Numbers and finally by Null.

```json5
[
  { _id: 1, name: 'T-shirt', size: 'M' },
  { _id: 3, name: 'shorts', size: 34 },
  { _id: 2, name: 'jeans', size: 32 },
  { _id: 4, name: 'belt', size: null }
]
```

## Roundup
