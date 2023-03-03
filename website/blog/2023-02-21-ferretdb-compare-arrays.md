---
title: "How MongoDB Compares Arrays in Filtering"
slug: ferretdb-compare-arrays
authors: [chi]
description: "In this article, we explore in detail how MongoDB BSON arrays are compared
in filter operators."
keywords:
  - MongoDB $eq operator
  - MongoDB implicit match
  - MongoDB $lt operator
  - MongoDB $gt operator
  - MongoDB array element match
  - MongoDB array value match
image: /img/blog/ferretdb-how-mongodb-compares-arrays.jpg
unlisted: true
---

In this article, we will explore how MongoDB BSON arrays are compared in filter operators.

![How MongoDB compares arrays](/img/blog/ferretdb-how-mongodb-compares-arrays.jpg)

<!--truncate-->

Comparison is a big part of various operations, such as sorting, filter operators, as well as update operators.
The major purpose of comparing BSON values is to determine which value is equal, greater or less than the other.

In this blog post, we aim to discuss how BSON values are compared in [filter operators](https://docs.ferretdb.io/basic_operations/read/#retrieve-documents-using-operator-queries), with a focus on arrays.

## BSON comparison order

The order of comparison for BSON types is predefined and used to compare different BSON values.
If two BSON values share the same comparison order, their values are compared to determine which one is greater.
However, if the comparison orders are different, the operator has a predefined logic to handle such cases.
Additionally, an element from a BSON array may be used to compare with other BSON types.

Below is the predefined specific order of comparison, listed from the lowest to the highest in the table below.

![BSON comparison order](/img/blog/bson-comparison-order.jpg)

:::note
All number types go through conversion, so they have the same BSON comparison order, regardless of `double`, `int32`, `int64` or `decimal`.
:::

### Understanding filter comparison rules

The `$eq` operator will not match a document value with a filter argument that is not the same type.
This is because the filter argument and the value are not identical types, so the operator is unable to find a match.

Similarly, the `$lt` operator behaves in the same way.
It will not select values that are not of the same type as the filter argument.

Suppose you want to use the `$lt` operator to compare similar data types with the filter argument as a string.
Then only the string values will be selected for comparison.

Meanwhile, to compare different types, let’s use a string filter argument `'a'` and a null value in the document.
It's important to note that **null is the lowest in the BSON comparison order**.
So what happens in this scenario?

When the BSON types being compared are not the same, the operator cannot make a comparison, and the filter will not match the value.
If the `$lt` operator encounters such cases, it will return the response that `null` is not less than `'a'`.

This shows that the `$lt` filter does not match if the BSON type is different.
In this case, BSON comparison order is used to filter out different
BSON type, but it does not use the BSON comparison order for `$lt` filter result.

This is also the case for the `$gt` operator which only selects the same BSON type.

**Example:** Array value comparison

To compare an array filter argument with array values, the elements of the array are iterated.
At each index, the BSON comparison order is used to compare the values.

:::note
If the values at a particular index are not equal, that result is used as the final comparison result.
If there are no more elements in the array during the iteration, then that array is considered smaller.
:::

To compare an array filter argument and an array field in a collection document, we compare each element in the arrays until we discover a difference.

For example, the following steps show how to use the `$lt` filter to compare the filter argument `[3, 1, 'b']` with the array field `[3, 'b', 'a']` in a document.

![Array comparison](/img/blog/array-comparison.jpg)

1. Compare `[3, 1, 'b']` and `[3, 'b', 'a']` using comparison order.
   They have the same BSON type; both are arrays.
2. Compare the first index `3` and `3` using comparison order, both are numbers, and they are the same.
3. Compare the next index, `1` and `'b'` using the BSON comparison order since they have different types, so `1` is a number and less than string `'b'`.

So `[3, 1, 'b']` is less than `[3, 'b', 'a']`.

:::note
If an array has no more element to compare, the one with no more element is less.
For the same reason, an empty array is less than arrays with elements.
:::

**Example:** Array element comparison

When using a filter argument to compare with an array, an element from the array is selected for comparison.
This is true for both implicit operators and the `$eq` operator.

When using the `$lt` and `$lte` operators, the smallest element from the array in the collection is used to compare against the filter argument.

When comparing a filter argument with an array in the collection using the `$lt` filter, the smallest element from the array is used for the comparison.
The following steps show how the `$lt` filter is used to compare the filter argument `2` with the document `{v: [3, null, 'a', 1]}` in the collection:

![Array element comparison](/img/blog/number-array.jpg)

1. Identify the BSON comparison order element of the filter argument, which is a number.
2. Choose the array elements that are numbers from the array field, which are `[3, 1]`.
3. Get the smallest element from the array field, which is `1`.
4. Compare `1` with filter argument `2`.
Since `1` is less than `2`, the operation result is true.
5. Therefore, when using `{v: {$lt: 2}}`, the document `{v: [3, null, 'a', 1]}` is selected.

For the `$gt` and `$gte` operators, the largest element is used for comparison instead.
The following steps show the `$gt` operator being used to compare the filter argument `2` with the array field `[3, 1]` in the collection:

1. Get the largest element of the array field `[3, 1]`, which is `3`.
2. Compare `3` with filter argument `2`.
Since `3` is greater than `2`, the operation result is true.
3. Therefore, when using the `$gt` filter, `[3, 1]` is greater than `2`.

It may seem confusing at first, but note that `$lt` uses the smallest element from the array, while `$gt` uses the largest element.
Additionally, `$lt`, `$gt`, `$lte`, and `$gte` only select elements of the same BSON comparison order as the filter argument by using BSON comparison order.

## Filter operator comparison

In this section, we use `mongosh` to show examples of the comparison in filtering.

### `$lt` operator comparison

[$lt](https://docs.ferretdb.io/basic_operations/read/#retrieve-documents-using-operator-queries)
filter operators select records that are less than the query.
We use the following collection to demonstrate the `$lt` filter argument.

```js
db.ltexample.insertMany([
  { _id: 'scalar-null', v: null },
  { _id: 'scalar-number', v: 1 },
  { _id: 'scalar-string', v: 'a' },
  { _id: 'array-empty', v: [] },
  { _id: 'array-null', v: [ null ] },
  { _id: 'array-mixed', v: [ 1, 'z', 'b' ] },
  { _id: 'array-string', v: [ 'd' ] },
  { _id: 'array-bool', v: [ false ] },
  { _id: 'nested-empty', v: [ [] ] },
  { _id: 'nested-mixed', v: [ [ 'c' ], [ 1, 'z', 'b' ] ] },
  { _id: 'nested-string', v: [ [ 'g' ] ] },
  { _id: 'nested-bool', v: [ [ true ] ] }
])
```

**Example:** `$lt` operator with scalar argument

When the filter argument is a string, documents with string records
that has less value would be selected.

```js
db.ltexample.find({v: {$lt: 'c'}})
```

The output:

```js
[
  { _id: 'scalar-string', v: 'a' },
  { _id: 'array-mixed', v: [ 1, 'z', 'b' ] }
]
```

The documents selected are `scalar-string` and `array-mixed`.

The `scalar-string` document has a smaller value compared to the filter argument.

The array record `array-mixed` was selected because the smallest element `b` is
less than the filter argument `c`.

If a document does not contain the same BSON type as the filter argument such as `array-null`, no selection is made.

**Example:** `$lt` operator with array argument

When an array is used as the filter argument, documents with array fields containing values that are less than the values in the filter array are selected.
It also selects nested arrays if the inner array is less than the filter argument.

The following example applies the `$lt` operator with an array argument `['c']`.

```js
db.ltexample.find({v: {$lt: ['c']}})
```

The output:

```js
[
  { _id: 'array-empty', v: [] },
  { _id: 'array-null', v: [ null ] },
  { _id: 'array-mixed', v: [ 1, 'z', 'b' ] },
  { _id: 'nested-empty', v: [ [] ] },
  { _id: 'nested-mixed', v: [ [ 'c' ], [ 1, 'z', 'b' ] ] }
]
```

No scalar values were selected, because array argument only selects arrays.
The filter argument selected arrays that are less
and nested arrays which contain an array that is less.

So _contains_ condition is applied on documents containing less than the filter argument, but not the other way around.

For selected `array-null` and `array-mixed`, each index was compared against the filter argument.
Using BSON comparison order, the first index element of both arrays is less than the filter argument.

An empty array is smaller than an array with items, so `array-empty` was selected.
The nested array `nested-empty` was also selected, this shows that `$lt` fetches the smallest
element from nested array and compares it to its array filter.

For the selected `nested-mixed`, the smallest inner array `[ 1, 'z', 'b' ]` is used for comparison.
Using the same logic, `nested-string` was not selected because
it does not contain an inner array smaller than the argument filter.

**Example:** `$lt` operator with nested array argument

The following example shows how to apply `$lt` operator with a nested array argument `[['c']]`.

```js
db.ltexample.find({v: {$lt: [['c']]}})
```

The output:

```js
[
  { _id: 'array-empty', v: [] },
  { _id: 'array-null', v: [ null ] },
  { _id: 'array-mixed', v: [ 1, 'z', 'b' ] },
  { _id: 'array-string', v: [ 'd' ] },
  { _id: 'nested-empty', v: [ [] ] },
  { _id: 'nested-mixed', v: [ [ 'c' ], [ 1, 'z', 'b' ] ] },
  { _id: 'nested-string', v: [ [ 'g' ] ] }
]
```

No scalar values from the collection was selected as scalar values are not
array type.

For arrays that were selected, upon comparing
each index, they are less than the filter argument.

For nested arrays, `nested-string` was selected even though the first element
of the inner array is not less than the argument filter.
This shows that nested array arguments and nested arrays in collections are not compared index to index.

Instead, it takes the smallest array from the nested array and uses it to compare against the filter argument.
The smallest inner array `['g']` is used to compare with the filter argument `[['c']]`.
Using BSON comparison order to compare, the first element `g` of string BSON type
which is smaller than the array BSON type `['c']` of the filter argument.

An array that was not selected was `nested-bool` due to boolean BSON type being greater than
array BSON type in the BSON comparison order.
For this, argument `[['c']]` was compared with the smallest
element of the nested array of `nested-bool` which is `[true]`.

### `$gt` operator comparison

[$gt](https://docs.ferretdb.io/basic_operations/read/#retrieve-documents-using-operator-queries)
filter operators select records that are greater than the query.

We use the `gtexample` collection.

```js
db.gtexample.insertMany([
  { _id: 'scalar-null', v: null },
  { _id: 'scalar-number', v: 1 },
  { _id: 'scalar-string', v: 'a' },
  { _id: 'array-empty', v: [] },
  { _id: 'array-null', v: [ null ] },
  { _id: 'array-mixed', v: [ 1, 'z', 'b' ] },
  { _id: 'array-string', v: [ 'd' ] },
  { _id: 'array-bool', v: [ false ] },
  { _id: 'nested-empty', v: [ [] ] },
  { _id: 'nested-mixed', v: [ [ 'c' ], [ 1, 'z', 'b' ] ] },
  { _id: 'nested-string', v: [ [ 'a' ] ] },
  { _id: 'nested-bool', v: [ [ true ] ] }
])
```

**Example:** `$gt` Operator with scalar argument

The following uses `$gt` to select records that are greater than `'c'`.

```js
db.gtexample.find({v: {$gt: 'c'}})
[
  { _id: 'array-mixed', v: [ 1, 'z', 'b' ] },
  { _id: 'array-string', v: [ 'd' ] }
]
```

The records selected are `array-mixed` and `array-string`.
They are both arrays contain string BSON value that is greater than the filter argument of `c`.

It did not select arrays that does not contain string BSON type such as `array-null`.

No nested arrays were selected because array is not the same BSON type as string.

**Example:** `$gt` Operator with Array Argument

The following demonstrates the array filter argument `['c']` using the `$gt` operator.

In the previous section, `$lt` array argument operator selected arrays that are less
and nested arrays that contain less.

:::caution
This example shows an uncertain result of `$gt` operator.
:::

```js
db.gtexample.find({v: {$gt: ['c']}})
```

The output:

```js
[
  { _id: 'array-string', v: [ 'd' ] },
  { _id: 'array-bool', v: [ false ] },
  { _id: 'nested-empty', v: [ [] ] },
  { _id: 'nested-mixed', v: [ [ 'c' ], [ 1, 'z', 'b' ] ] },
  { _id: 'nested-string', v: [ [ 'g', 'z' ] ] },
  { _id: 'nested-bool', v: [ [ true ] ] }
]
```

No scalar documents were selected.

Here, we witness the same behaviour of `$lt` on array argument because it only selects array BSON type.

In the query, `array-bool` and `array-string` were selected since they contain values that is
greater than the filter argument of `c` at its first index.

The arrays are compared index to index using BSON comparison order then by value comparison.

When dealing with nested arrays, using the `$gt` operator on an array filter argument does not select the largest element from a nested array, unlike how `$lt` operates.

Instead, the `$gt` operator compares the array filter argument and the nested array in the collection index to index, using the BSON comparison order.

Because the BSON type of the array in the collection is greater than the BSON type of the filter argument (which is a string), all nested arrays are selected.
This is a significant difference from how the `$lt` filter works with an array argument.
`$lt` selects the smallest element from a nested array, while `$gt` compares index to index.

We are yet to fully understand how a nested array is expected to work with a filter comparison operator.

**Example:** `$gt` Operator with Nested Array Argument

:::caution
This example also shows an uncertain result of the `$gt` operator.
:::

The following example shows the `$gt` operator with a nested array argument `[['c']]`.

```js
db.gtexample.find({v: {$gt: [['c']]}})
```

The output:

```js
[
  { _id: 'array-bool', v: [ false ] },
  { _id: 'nested-mixed', v: [ [ 'c' ], [ 1, 'z', 'b' ] ] },
  { _id: 'nested-bool', v: [ [ true ] ] }
]
```

No scalar values from the collection were selected since the scalar values are not array type.

The selected array `array-bool` contains a boolean value, which has a greater BSON comparison order than that of an array type.

Using the BSON comparison order, the boolean value is considered greater than an array, hence the entire `array-bool` array was selected.

When dealing with nested arrays, the arrays `nested-mixed` and `nested-bool` were selected when using the `$gt` operator on an array filter argument.

In `$gt`, index-to-index comparison is performed between the filter argument and the arrays in the collection.

For `nested-mixed`, both the array in the collection and the filter argument have the same element `['c']` at the first index, but the array in the collection has an additional item at the second index, making it greater than the filter argument.

For `nested-bool`, the array in the collection has a boolean value at the first index of the inner array, which has a greater BSON comparison order than the string in the filter argument.
This makes `nested-bool` greater than the filter argument.

Unlike `$lt` which takes the smallest element from the array, nested array is compared index to index in `$gt`.

## Nested array at FerretDB

At FerretDB, we do not support nested arrays at the moment.
This is due to the complexities caused by
differences in the comparison of array and other BSON types.

In particular, differentiating between the cases where the BSON comparison order of an array is used versus an element from the comparison array was not entirely straightforward.

Since then, we’ve taken time to understand the inner workings of comparison for various operations.

After investigating the behaviour of the `$lt` and `$gt` operators with array and nested array filter arguments, we discovered that the comparison process is different between the two cases.

However, given an array argument filter and array collection, `$lt` uses the smallest element from the array and `$gt` on the other hand compares index to index of array collection.

Perhaps using comparison operators like `$lt` and `$gt` for nested array is
not a common usage and the result we found was a corner case.

If you have any questions or feedback, please [let us know](https://docs.ferretdb.io/#community)!
We're always here to help you get the most out of FerretDB.
