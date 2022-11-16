---
sidebar_position: 5
---

# Delete operation

The delete operation removes a document from the database when a given query is met.
Two methods for deleting documents in a collection include `deleteOne()` and `deleteMany()`.

## Delete a single document

The `deleteOne()` method removes a single document (the first document that matches the query parameter) completely from the collection.

```js
db.collection.deleteOne({<query_params>})
```

Insert the following list of documents:

```js
db.scientists.insertMany([
  {
    firstname: "Thomas",
    lastname: "Edison",
    born: 1847,
    invention: "LightBulb",
    nobel: true
  },
  {
    firstname: "Graham",
    lastname: "Bell",
    born: 1847,
    invention: "telephone",
    nobel: false
  },
  {
    firstname: "Nikola",
    lastname: "Tesla",
    born: 1856,
    invention: "Tesla coil",
    nobel: false
  },
  {
    firstname: "Ada",
    lastname: "Lovelace",
    born: 1815,
    invention: "Computer programming",
    nobel: false
  }
])
```

This operation returns a response showing `acknowledged` as `true` and the `ObjectId` of the four inserted documents:

```js
{
  acknowledged: true,
  insertedIds: {
    '0': ObjectId("63470121d7a4a1b0b38eb2df"),
    '1': ObjectId("63470121d7a4a1b0b38eb2e0"),
    '2': ObjectId("63470121d7a4a1b0b38eb2e1"),
    '3': ObjectId("63470121d7a4a1b0b38eb2e2")
  }
}
```

Next, delete a document from the collection where the field `nobel` is set to false.

```js
db.scientists.deleteOne({nobel:false})
```

This operation returns a response that shows that a single document was deleted from the collection.

```js
{ acknowledged: true, deletedCount: 1 }
```

## Deletes multiple documents

To delete multiple documents at once, use the  `deleteMany()` method.
Using the same record from earlier, let's delete all the documents with `nobel` set to false.

```js
db.scientists.deleteMany({nobel:false})
```

This command removes all the documents in the collection that matches the query.
