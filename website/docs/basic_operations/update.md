---
sidebar_position: 4
---

# Update operation

The update operation modifies a document record in a collection, based on a given query parameter and update.
FerretDB supports update operators, such as `$set` and `$setOnInsert` to update documents in a record.

At present, FerretDB currently supports the following update operators:

| Operator name  | Description                                                                                                                                                        |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `$set`         | Assigns the value for an updated field to the document.                                                                                                            |
| `$setOnInsert` | Specifies the value of a field when an update operation results in the addition of a document.  However, there is no effect when it modifies an existing document. |
| `$unset`       | Removes a specific field from a document.                                                                                                                          |
| `$pop`         | In an array, this operator removes the first or last item.                                                                                                         |

## Update a single document

Use the `updateOne()` method to update a single document in a collection.
This operation filters a collection using a query parameter, and updates given fields within that document.

```js
db.collection.updateOne({<query-params>}, {$set: {<update fields>}})
```

First, populate the database with a collection containing a list of documents.

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

Using the document record in the collection, update the document where `firstname` is "Graham", and set it as "Alexander Graham".
The `updateOne()` operation will only affect the first document that’s retrieved in the collection.

```js
db.scientists.updateOne({
   firstname: "Graham"
},
{
   $set:{
      firstname: "Alexander Graham"
   }
})
```

## Replace a document

Besides updating a document, you can replace it completely using the `replaceOne()` method.

```js
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

## Update many documents

Using the `updateMany()` command, you can modify many documents at once.
In the example below, where `nobel` is set as false, update and set to true.

```js
db.scientists.updateMany({nobel:false}, {$set: {nobel:true}})
```

This operation updates all the documents where the field `nobel` was previously false.

## Update an array element

The following update example uses the `employees` collection.
To populate the collection, run the following in your terminal:

```js
db.employees.insertMany([
  {
     name: {
        first: "Earl",
        last: "Thomas"
     },
     employeeID: 1234,
     age: 23,
     role: "salesperson",
     catalog: [
        "printer",
        "cardboard",
        "crayons",
        "books"
     ]
  },
  {
     name: {
        first: "Sam",
        last: "Johnson"
     },
     employeeID: 2234,
     age: 35,
     role: "salesperson",
     catalog: [
        "cabinet",
        "fridge",
        "blender",
        "utensils"
     ]
  },
  {
     name: {
        first: "Clarke",
        last: "Dane"
     },
     employeeID: 3234,
     age: 21,
     role: "salesperson",
     catalog: [
        "printer",
        "pencils",
        "crayons",
        "toys"
     ]
  }
])
```

The following command will query and update the `catalog` array in the `employee` collection using dot notation.
The command will query the second field of the array in every document for `"pencil"`, and when there is a match, updates the first element of the array.

```js
db.employees.updateMany({
    "catalog.1": "pencils"
},
{
    $set: {
        "catalog.0": "ruler"
    }
})
```

The response from the command:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 1,
  modifiedCount: 1,
  upsertedCount: 0
}
```

## Update an embedded document

To update an embedded document, use dot notation to specify the fields to modify.
The following operation updates any embedded document that matches the specified query in the `employees` collection.

```js
db.employees.updateMany({
    "name.first": "Clarke"
},
{
    $set: {
        "name.last": "Elliot"
    }
})
```

The following response from the command shows that a single document matching the query was updated:

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 1,
  modifiedCount: 1,
  upsertedCount: 0
}
```
