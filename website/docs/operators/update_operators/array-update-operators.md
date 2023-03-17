---
sidebar_position: 1
---

# Array update operators

Array update operators  allow you to modify the elements of an array field in a document.

| Operator                 | Description                                                                                  |
| ------------------------ | -------------------------------------------------------------------------------------------- |
| [`$push`](#push)         | Adds an element to an array                                                                 |
| [`$addToSet`](#addtoset) | Adds elements to a specific array as long as the element does not already exist in the array |
| [`$pop`](#pop)           | Removes either the first or the last element of an array                                     |
| [`$pullAll`](#pullall)   | Removes all matching values in a specified query from an array                               |

## $push

The `$push` operator updates a document by adding an element to a specified array.
If the array does not exist, a new array is created with the element added to it.

Insert the following document into a `store` collection:

```js
db.store.insertMany([
  { _id: 1, items: ["pens", "pencils", "paper", "erasers", "rulers"] },
]);
```

**Example:** Use the `$push` operator to add an element to an existing array.

```js
db.store.updateOne(
  { _id: 1 },
  { $push: { items: "markers" } }
);
```

After the operation, the updated document looks like this:

```js
[ {
    _id: 1,
    items: [ 'pens', 'pencils', 'paper', 'erasers', 'rulers', 'markers' ]
    }]
```

## $addToSet

The `$addToSet` operator updates an array by adding a specified element to an array if the elenent does not already exist in the array.
If the specified element exists in the array, the `$addToSet` operator will not modify the array.

Insert the following documents into a `store` collection:

```js
db.store.insertMany([
  { _id: 1, items: ["pens", "pencils"] },
]);
```

**Example:** Use the `$addToSet` operator to update the array with non-existing elements.

```js
db.store.updateOne(
  { _id: 1 },
  { $addToSet: { items: "paper" } }
);
```

The document is subsequently updated with the new element, as depicted below:

```js
[ { _id : 1, items : [ 'pens', 'pencils', 'paper' ] } ]
```

**Example:** Use the `$addToSet` operator to update the array with already existing elements.

```js
db.store.updateOne(
  { _id: 1 },
  { $addToSet: { items: "pens" } }
);
```

Since the array already contains the element, there won't be any changes.

```js
[ {_id: 1, items: [ 'pens', 'pencils', 'paper' ] } ]
```

:::note
The `$addToSet` is different from the `$push` operator which adds the element to the array either it exists or not.
:::

**Example:** Use the `$addToSet` operator for non-existing array fields.

If the array field does not exist in the document, the `$addToSet` operator will create the field and add the element to the array.

```js
db.store.updateOne(
  { _id: 1 },
  { $addToSet: { colors: "red" } }
);
```

The updated document looks like this:

```js
[
  { _id: 1, items: [ 'pens', 'pencils', 'paper' ], colors: [ 'red' ] }
]
```

## $pop

With the `$pop` operator, you can update a document by removing the first or last element of an array.
Assign a value of `-1` to remove the first element of an array, or `1` to remove the last element.

Insert this document into a `products` collection:

```js
db.products.insertMany([
    { _id: 1, items: [ "pens", "pencils", "paper", "erasers", "rulers" ] }
]);
```

**Example:** Use the `$pop` operator to remove the first element of an array.

```js
db.products.updateOne(
  { _id: 1 },
  { $pop: { items: -1 } }
);
```

The document is subsequently updated with the first element `pens` removed, as depicted below:

```js
[ {
    _id: 1,
    items: [ 'pencils', 'paper', 'erasers', 'rulers' ]
    }]
```

To remove the last element of the array, assign `1` as the value for the `$pop` operator.

```js
db.products.updateOne(
  { _id: 1 },
  { $pop: { items: 1 } }
);
```

The updated now looks like this:

```js
[ {
    _id: 1,
    items: [ 'pencils', 'paper', 'erasers' ]
    }]
```

## $pullAll

The `$pullAll` operator removes all the matching elements in a specified query from an array.

Insert the following document into a `store` collection:

```js
db.store.insertMany([
  { _id: 1, items: ["pens", "pencils", "paper", "erasers", "rulers"] },
]);
```

**Example:** Use the `$pullAll` operator to remove multiple elements from an array.

```js
db.store.updateOne(
  { _id: 1 },
  { $pullAll: { items: ["pens", "pencils", "paper"] } }
);
```

After removing all instances of the specified array elements, the document is updated as follows:

```js
[ {
    _id: 1,
    items: [ 'erasers', 'rulers' ]
}]
```

**Example:** Use the `$pullAll` operator to remove array of objects from an array.

Insert the following document into a `fruits` collection:

```js
db.fruits.insertMany([
    { _id: 1, fruits: [
        { type: "apple", color: "red" },
        { type: "banana", color: "yellow" },
        { type: "orange", color: "orange" }
        ] }
]);
```

The following query uses the `$pullAll` to remove all matching array objects from the specified document.

```js
db.fruits.update(
  { _id: 1 },
  {
    $pullAll: {
      fruits: [
        { type: "apple", color: "red" },
        { type: "banana", color: "yellow" }
      ]
    }
  }
);
```

The updated document now looks like this:

```js
[
  { _id: 1, fruits: [ { type: 'orange', color: 'orange' } ] }
]
```
