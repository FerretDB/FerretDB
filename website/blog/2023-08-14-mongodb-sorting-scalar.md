---
slug: mongodb-sorting-scalar
title: How MongoDB Sorting Works for Scalar Values
authors: [chi]
description: >
  In this blog post, we explore how MongoDB sorting works for scalar values.
keywords:
  - MongoDB sorting
  - BSON comparison order
image: /img/blog/post-cover-image.jpg
tags: [community, product, tutorial]
---

![MongoDB sorting works for scalar values](/img/blog/mongodb-sorting-scalar.jpg)

In this blog post, we explore how MongoDB sorting works for scalar values.

<!--truncate-->

Sorting compares BSON values to determine which value is equal, greater or less than the other to order them in the ascending or descending order.
For comparing different BSON types, [BSON comparison order](#bson-comparison-order) is used.

## BSON comparison order

If two BSON values share the same BSON type, their values are compared to determine which value is greater or less.
However, if the BSON types are different, a predefined BSON comparison order is used to determine which BSON type is greater or less.

Below table shows the predefined BSON comparison order for each BSON type.

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

For comparing different BSON types, each BSON type has a predefined order of comparison assigned.
For example, Null BSON type has the lowest order of comparison 1, so Null is less than any other BSON values.
Boolean BSON type on the other hand has the order of comparison 8, a BSON type such as ObjectId with a lower order of comparison is less than a Boolean value
and a BSON type such as Timestamp with a higher order of comparison is greater than a Boolean value.

Comparing different BSON types is merely looking up the order of comparison table for each BSON type and comparing the predefined order.
There is an exception for Array values, we will be discussing them in another blog post.

### Number comparison

Although numbers have different BSON types, namely Integers, Longs, Doubles and Decimals, they are considered the equivalent BSON type for the purpose of comparison.
That means comparing numbers consider their values but whether they are Integers, Longs, Doubles or Decimals are not relevant.
For instance, an Integer value 0 is equivalent to Double 0.0 as far as comparison is concerned.

### Null and non-existent field comparison

For the comparison purpose, non-existent field is equivalent to Null.
This means that Null and non-existent field are equal as far as comparison is concerned.

## Examples showcasing Sorting for scalar values

Suppose there is an `outfits` collection, following query inserts documents.

```js
db.outfits.insertMany([
  { _id: 1, name: 'flip flops', size: 'M', color: 'blue' },
  { _id: 2, name: 'sandals', size: 9, color: null },
  { _id: 3, name: 'boots', size: 8, color: 'black' },
  { _id: 4, name: 'sneakers', size: 8.5, color: 'blue' },
  { _id: 5, name: 'slippers' }
])
```

The `outfits` collection has a `size` field, and it contains different BSON types.
The document `flip flops` has String field value, `sandals` and `boots` have Integer field values,
`sneakers` has Double field value and `slippers` is missing the field `size`.
To sort the collection in ascending order by `size` field, sorting order 1 is used and the following query is run.

```js
db.outfits.find().sort({ size: 1 })
```

```json5
[
  { _id: 5, name: 'slippers' },
  { _id: 3, name: 'boots', size: 8, color: 'black' },
  { _id: 4, name: 'sneakers', size: 8.5, color: 'blue' },
  { _id: 2, name: 'sandals', size: 9, color: null },
  { _id: 1, name: 'flip flops', size: 'M', color: 'blue' }
]
```

The output is sorted and the first document is `slippers` which is missing the `size` field.
A [non-existent field is equivalent to Null](#null-and-non-existent-field-comparison), so it has the lowest BSON type.

Then the next documents are Numbers.
The numbers have higher BSON order of comparison than Null BSON type, so they come after `slippers` of the missing field.
The documents with numbers are `boots` with Integer BSON type, `sneaker` with Double BSON type and `sandals` with Integer BSON type.
Notice that Integer BSON type is followed by Double BSON type then by another Integer BSON type?
The numbers are considered [equivalent BSON types](#number-comparison) so only the values of each number are compared regardless of its specific BSON number type.
The document `boots` has `size` field value of 8 which is less than 8.5 of `sneakers` or 9 of `sneakers`, so it comes next.
Then the document `sneakers` comes next.

Finally, `flip flops` with String BSON type which has a higher BSON order of comparison than Numbers comes last.

Similarly, to sort the collection in descending order by `size` field, sorting order -1 is used and the following query is run.

```js
db.outfits.find().sort({ size: -1 })
```

```json5
[
  { _id: 1, name: 'flip flops', size: 'M', color: 'blue' },
  { _id: 2, name: 'sandals', size: 9, color: null },
  { _id: 4, name: 'sneakers', size: 8.5, color: 'blue' },
  { _id: 3, name: 'boots', size: 8, color: 'black' },
  { _id: 5, name: 'slippers' }
]
```

This time, the output is sorted by the higher BSON type String document `flip flops`, then by documents with Numbers `sandals`, `sneakers` and `boots` and finally `slippers` with a non-existent `size` field.

Suppose you want to sort by `color`. There are more than one document with color with `blue`, and also there is a document with Null and missing `color` field.

For example, `flip flops` has a Null value for `color` field and `slippers` is missing the field.
Null and non-existent field is considered equivalent so either of them can be the first.
In such a scenario, the default order that results were found from the database is used.

To consistently preserve the same sorting order, it is recommended to use `_id` as the second field for sorting.
Such case, if `color` has an equivalent value, it uses `_id` to sort them, allowing consistent output.
The uniqueness property of `_id` makes the sorting output consistent.

```js
db.outfits.find().sort({ color: 1, _id: 1 })
```

```json5
[
  { _id: 2, name: 'sandals', size: 9, color: null },
  { _id: 5, name: 'slippers' },
  { _id: 3, name: 'boots', size: 8, color: 'black' },
  { _id: 1, name: 'flip flops', size: 'M', color: 'blue' },
  { _id: 4, name: 'sneakers', size: 8.5, color: 'blue' }
]
```

The output shows that `sandals` is sorted before `slippers` even though they have equivalent `color` field value, because it uses the second sort field `_id` and `sandals` has a lower `_id`.
Similarly, `flip flops` and `sneakers` have the same `color` field value. But `flip flops` is sorted before because it has a lower `_id`.

## Roundup

In this blog post, it shows how BSON comparison order is used in sorting and how scalar values are compared.
We will explain about how Array and Object sorting works in an upcoming blog post.
Stay tuned!
