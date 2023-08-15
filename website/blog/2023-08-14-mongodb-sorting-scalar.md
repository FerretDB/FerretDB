---
slug: mongodb-sorting-scalar
title: How MongoDB sorting works for scalar values
authors: [chi]
description: >
  In this blog post, we explore how MongoDB sorting works for scalar values.
keywords:
  - MongoDB sorting
  - BSON comparison order
image: /img/blog/post-cover-image.jpg
tags: [community, product, tutorial]
draft: true
---

![MongoDB sorting works for scalar values](/img/blog/mongodb-sorting-scalar.jpg)

In this blog post, we explore how MongoDB sorting works for scalar values.

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

## Comparison of different BSON types

For comparing different BSON types, each BSON type has predefined order of comparison assigned.
Null BSON type has the lowest order of comparison, so Null is the lowest BSON value.
When you compare Null with String, Null is less than any String.

Comparing different BSON type is merely looking up the order of comparison table for each BSON type and using predefined order.
There are exceptions for Object and Array values, we will be discussing them in another blog post.

## Number comparison

Although numbers have different BSON types Integers, Longs, Doubles and Decimals, they are considered equivalent BSON type for the purpose of comparison.
That means comparing numbers consider their values but whether they are Integers, Longs, Doubles or Decimals are not relevant.
For instance, an Integer value 0 is equivalent to Double 0.0 as far as comparison is concerned.

## Null and missing field comparison

For the comparison purpose, missing field is equivalent to Null.
This means that Null and missing field are equal as far as comparison is concerned.

## Sorting examples

Suppose you have `outfits` collection, it has `size` field specifying its size.
Of the collection, `size` field of `flip flops` has String field value, `sandals` and `boots` have Integer field value,
`sneakers` has Double field value and `slippers` is missing field `size`.

```js
db.outfits.insertMany([
  { _id: 1, name: 'flip flops', size: 'M' },
  { _id: 2, name: 'sandals', size: 9 },
  { _id: 3, name: 'boots', size: 8 },
  { _id: 4, name: 'sneakers', size: 8.5 },
  { _id: 5, name: 'slippers' }
])
```

To sort the collection in ascending order by `size` field, following query is run.

```js
db.outfits.find().sort({ size: 1 })
```

The output is sorted first by `slippers` which is missing the `size` field.
A [missing field is equivalent to Null](#null-and-missing-field-comparison), so it has the lowest BSON type Null.

Then the next ones are Numbers.
The numbers have higher BSON order of comparison than Null BSON type, so they come after `slippers`.
The ones with number are `boots` with Integer BSON type, `sneaker` with Double BSON type and `sandals` with Integer BSON type.
The numbers are considered [equivalent BSON types](#number-comparison) so only the values of each number is compared regardless of specific BSON number type.
`boots` has 8 which is less than 8.5 of `sneakers` or 9 of `sneakers`, so it comes next.

Finally `flip flops` with String BSON type which has higher BSON order of comparison than Number is sorted the last.

```json5
[
  { _id: 5, name: 'slippers' },
  { _id: 3, name: 'boots', size: 8 },
  { _id: 4, name: 'sneakers', size: 8.5 },
  { _id: 2, name: 'sandals', size: 9 },
  { _id: 1, name: 'flip flops', size: 'M' }
]
```

Similarly, the collection is sorted in descending order by `size` field by running following query.

```js
db.outfits.find().sort({ size: -1 })
```

This time, the output is sorted by the higher BSON type String, then by Numbers and finally one with missing field.

```json5
[
  { _id: 1, name: 'flip flops', size: 'M' },
  { _id: 2, name: 'sandals', size: 9 },
  { _id: 4, name: 'sneakers', size: 8.5 },
  { _id: 3, name: 'boots', size: 8 },
  { _id: 5, name: 'slippers' }
]
```

## Roundup

In this blog post, it shows how BSON comparison order is used in sorting and how scalar values are compared.
We will explain about how Array and Object sorting works in upcoming blog post.
Stay tuned!
