---
slug: mongodb-sorting-scalar
title: How MongoDB sorting works for scalar values
authors: [chi]
description: >
  In this blog post, we explore in detail how MongoDB sorting works.
keywords:
  - MongoDB sorting
  - BSON comparison order
image: /img/blog/post-cover-image.jpg
tags: [community, product, tutorial]
draft: true
---

![MongoDB sorting works for scalar values](/img/blog/mongodb-sorting-scalar.jpg)

In this blog post, we explore in detail how MongoDB sorting works.

<!--truncate-->

Sorting compares BSON values to determine which value is equal, greater or less than the other to order them in the ascending or descending order.

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
For Null BSON type, order of comparison is 1.
For a String BSON type, order of comparison is 3.
When you compare Null with String, Null is less than String.

So comparing different BSON type is merely looking up the order of comparison table for each BSON type and using predefined order.

## How does comparison work for numbers?

Although numbers have different BSON types Integers, Longs, Doubles and Decimals, they are considered equivalent for the purpose of comparison.
That means zero value of any BSON number type is equivalent to other BSON number type.

## Sorting examples

## Roundup
