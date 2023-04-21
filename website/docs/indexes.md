---
sidebar_position: 7
---

# Indexes

Indexes are essential in improving query performance by enabling fast retrieval of relevant documents when querying large collections.

Indexes in FerretDB are created based on a specific field(s) within documents.
Creating an index involves having a data structure that maps the values in the indexed fields to the locations of the related documents, making it possible to retrieve documents more quickly based on those fields.

## How do indexes work?

Say a `products` collection contains the following documents:

```js
{ _id: 1, name: "iPhone 12", category: "smartphone", price: 799 }
{ _id: 2, name: "iPad Pro", category: "tablet", price: 999 }
{ _id: 3, name: "Galaxy S21", category: "smartphone", price: 699 }
{ _id: 4, name: "MacBook Pro", category: "laptop", price: 1299 }
```

Suppose you want to find all products with a price greater than `$900`.
Without an index, each query will need to scan each document in the collection to find the ones that match, which can be resource intensive.

However, if there's an index on the `price` field, the query only scans two documents where their prices are greater than `$900`, instead of all four.

## How to create indexes

Use the `createIndexes()` command to create an index on a collection.
This command takes two arguments: a document containing the index key (fields to index), and an optional document specifying the index options.

You can create single field indexes or compound indexes.

### Single Field Indexes

Here's an example of creating an index on the `price` field of a products collection:

```js
db.products.createIndex({ price: 1 })
```

This creates an ascending index on the `price` field.
`1` specifies the index direction for ascending order.
If it's `-1`, it specifies a descending order direction for the index.

### Compound Indexes

For compound indexes, you can create an index key combining multiple fields together as a key.
Below is an example of a compound index that uses `price` and `category` fields from the `products` collection as the index key.

```js
db.products.createIndex({ price: 1, category: 1 })
```

:::note

* If the `createIndex()` command is called for a non-existent collection, it will create the collection and its given indexes.
* If the `createIndex()` command is called for a non-existent field, an index for the field is created without creating or adding the field to an existing collection.
* If you attempt to create an index with the same name and key as an existing index, the system will not create a duplicate index.
Instead, it will simply return the name and key of the existing index, since duplicate indexes would be redundant and inefficient.
* Meanwhile, any attempt to call `createIndex()` command for an existing index using the same name and different key, _or_ different name but the same key will return an error.

:::

## How to list Indexes

To display a collection's index details, use the `listIndexes()` command.

To return the list of indexes in the `products` collection, use the following command:

```js
db.runCommand({listIndexes: "products"})
```

The returned indexes should look like this, showing the default index, single field index, and compound index.

```js
{
  cursor: {
    id: Long("0"),
    ns: 'db.products',
    firstBatch: [
      { v: 2, key: { _id: 1 }, name: '_id_' },
      { v: 2, key: { price: 1 }, name: 'price_1' },
      {
        v: 2,
        key: { price: 1, category: 1 },
        name: 'price_1_category_1'
      }
    ]
  },
  ok: 1
}
```

## How to drop indexes

You can also drop all the indexes or a particular index in a specified collection, except the default index (`_id`).

FerretDB supports the use of `dropIndex()` and `dropIndexes() command.

Use the `dropIndex()` command to drop a particular index from a collection.
Using the returned indexes above, let's drop the index with the name `price_1`.

```js
db.products.dropIndex( "price_1" )
```

Another way to perform this action is to use the same index document as the index you want to drop.
For the same example above, you can rewrite it as:

```js
db.products.dropIndex( { "price" : 1 } )
```

Use the `dropIndexes()` command to remove one or more indexes from the collection.

You can remove all the indexes from the collection besides the default index (`_id`).
The following command removes all indexes from the collection.

```js
db.products.dropIndexes()
```

This will drop all the non-`_id` indexes from the collection.
