---
slug: mongodb-sorting-scalar
title: How MongoDB Sorting Works for Scalar Values
authors: [chi]
description: >
  In this blog post, we explore how MongoDB sorting works for scalar values.
image: /img/blog/mongodb-sorting-scalar.jpg
tags: [community, product, tutorial]
---

![How MongoDB Sorting Works for Scalar Values](/img/blog/mongodb-sorting-scalar.jpg)

In this blog post, we delve into the process of sorting scalar values in MongoDB.

<!--truncate-->

Sorting in MongoDB involves comparing BSON values to ascertain their relative order – whether one value is equal to, greater than, or less than another.
The resultant sorted array can be in either ascending or descending order.

When comparing different BSON types, the [BSON comparison order](#bson-comparison-order) is used.

## BSON comparison order

If two BSON values share the same type, their values are compared to determine which is greater or less.

However, if the BSON types are different, a predefined BSON comparison order is used.

The table below shows the predefined BSON comparison order for each BSON type.

<!-- use newline in column header for appropriate spacing of columns -->
<!-- markdownlint-disable MD033 -->

| Order of Comparison<br/>(lowest to highest) | BSON Types                               |
| ------------------------------------------- | ---------------------------------------- |
| 1                                           | Null                                     |
| 2                                           | Numbers (Integer, Long, Double, Decimal) |
| 3                                           | String                                   |
| 4                                           | Object                                   |
| 5                                           | Array                                    |
| 6                                           | BinData                                  |
| 7                                           | ObjectId                                 |
| 8                                           | Boolean                                  |
| 9                                           | Date                                     |
| 10                                          | Timestamp                                |
| 11                                          | Regular Expression                       |

## Comparison of values with different BSON types

To compare values of different BSON types, look at the predefined comparison order given to each type.
For example, the Null BSON type has the lowest order, which is 1.
This means Null is less than any other BSON values.
The Boolean BSON type has an order of 8.
So, if you're comparing it with an ObjectId type that has a lower order, the Boolean is greater.
But if you're comparing it with a Timestamp type that has a higher order, the Boolean is less.

The key to comparing different BSON types is simply to check their predefined orders.
Arrays are an exception, and we'll talk about them in another blog post.

### Number comparison

Even though Numbers come in various BSON types – Integer, Long, Double, and Decimal – they're treated as the same type when comparing.
This means the focus is on the actual numerical values, not on whether they're Integer, Long, Double, or Decimal.
For example, an Integer value of 0 is seen as the same as a Double value of 0.0 when comparing them.

### Null and non-existent field comparison

For comparison, a non-existent field is equivalent to Null.
This means that a field `v` with Null value `{:v null}` is considered the same as a non-existent `v` field in `{}`.

## Examples showcasing sorting for scalar values

Let's create an `outfits` collection using the following query to insert documents.

```js
db.outfits.insertMany([
  { _id: 1, name: 'flip flops', size: 'M', color: 'blue' },
  { _id: 2, name: 'sandals', size: 9, color: null },
  { _id: 3, name: 'boots', size: 8, color: 'black' },
  { _id: 4, name: 'sneakers', size: 8.5, color: 'blue' },
  { _id: 5, name: 'slippers' }
])
```

The `outfits` collection includes a `size` field that represents various BSON types.
For instance, the document for `flip flops` contains a String value in this field, while `sandals` and `boots` have Integer values.
The `sneakers` document has the `size` field as a Double value, and the `slippers` document lacks the `size` field altogether.
To sort these documents in ascending order based on the `size` field, you would use a sorting order of 1 and execute the following query.

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

The sorted output starts with the `slippers` document, which lacks a `size` field.
According to our earlier discussion on [how Null and non-existent fields are equivalent](#null-and-non-existent-field-comparison), it has the lowest BSON type and appears first.

Next in line are documents with Number values in the `size` field.
Numbers hold a higher BSON comparison order than Null, so they appear after `slippers` document with the missing `size` field.
Specifically, we see `boots` with an Integer BSON type, followed by `sneakers` with a Double BSON type, and then `sandals` also with an Integer BSON type.
Why this particular order?
Because all Numbers, regardless of their BSON type, are [considered equivalent for comparison](#number-comparison).
Only the actual numerical values matter.
In this case, `boots` with a size of 8 comes before `sneakers` with a size of 8.5, which in turn precedes `sandals` with a size of 9.

Lastly, we have `flip flops` with a String BSON type.
Strings have a higher BSON comparison order than Numbers, so this document comes at the end of our sorted list.

To sort the documents in descending order by the `size` field, you would use a sorting order of -1 and execute the following query.

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

This time, the output is sorted first by `flip flops` with String `size` field, then by `size` field with Numbers `sandals`, `sneakers` and `boots` and finally `slippers` with a non-existent `size` field.

### Using `_id` as the second sort field

Suppose you want to sort the documents by the `color` field.
You encounter multiple documents with the color `blue`, and one document has a Null value for this field while another is missing it altogether.

For instance, `flip flops` has a Null value in the `color` field, and the `slippers` document lacks this field.
Since Null and non-existent fields are considered equivalent in sorting, either could appear first.
In situations like this, the default order in which the records were retrieved from the database is applied.

To maintain a consistent sort order, it's advised to use `_id` as a secondary sorting option.
In this setup, if multiple documents have the same or equivalent `color` values, it will rely on the `_id` field for sorting.
The unique nature of the `_id` field ensures that the sorted output remains consistent.

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

The output shows that the `sandals` document is sorted before `slippers`, despite both having equivalent values in the `color` field.
This is because the sort mechanism uses the secondary `_id` field for ordering, and `sandals` has a lower `_id` value than `slippers`.

Likewise, both `flip flops` and `sneakers` have the same `color` value, but `flip flops` comes first because its `_id` value is lower than that of `sneakers`.

## Roundup

This blog post has shown how BSON comparison order functions in sorting and how scalar values are compared against each other.
In an upcoming blog post, we will delve into how sorting of Arrays and Objects works.

Stay tuned!
