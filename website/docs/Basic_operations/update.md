---
sidebar_position: 4
---

# Update operation

The update operation modifies a document record in a collection, based on a given query parameter and update.
FerretDB supports update operators, such as `$set` and `$setOnInsert` to update documents in a record.

At present, FerretDB currently supports the following update operators:

| Operator name | Description |
| --- | --- |
| `$set` | Assigns the value for an updated field to the document. |
| `$setOnInsert` | Specifies the value of a field when an update operation results in the addition of a document.  However, there is no effect when it modifies an existing document. |
| `$unset` | Removes a specific field from a document. |
| `$pop` | In an array, this operator removes the first or last item. |

## Modify a single document

Use the `updateOne()` method to update a single document in a collection.
This operation filters a collection using a query parameter, and updates given fields within that document.

```sh
db.collection.updateOne({<query-params>}, {$set: {<update fields>}})
```

First, populate the database with a collection containing a list of documents.

```sh
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

Using the document record in the collection, update the document where `firstname` is "Graham", and set it as "Alexander Graham".
The `updateOne()` operation will only affect the first document that’s retrieved in the collection.

```sh
db.scientists.updateOne({
   firstname:"Graham"
},
{
   $set:{
      firstname:"Alexander Graham"
   }
})
```

## Modify many documents

Using the `updateMany()` command, you can modify many documents at once.
In the example below, where `nobel` is set as false, update and set to true.

```sh
db.scientists.updateMany({nobel:false}, {$set: {nobel:true}})
```

This operation updates all the documents where the field `nobel` was previously false.

## Replace a document

Besides updating a document, you can replace it completely using the `replaceOne()` method.

```sh
db.scientists.replaceOne({
    lastname: "Bell",
},
{
    lastname: "Einstein",
    firstname: "Albert",
    born: 1879,
    invention: "Photoelectric effect",
    nobel: true
})
```
