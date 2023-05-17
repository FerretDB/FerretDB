---
sidebar_position: 3
---

# Read operation

The read operation retrieves documents in a collection.
You can either retrieve all the documents in a collection, or only the documents that match a given query parameter.

## Retrieve a single document

The `findOne()` command retrieves a single document from a collection.

First, populate the database with a new collection containing a list of documents.

```js
db.scientists.insertMany([
  {
    name: {
      firstname: "Alan",
      lastname: "Turing"
    },
    born: 1912,
    invention: "Turing Machine"
  },
  {
    name: {
      firstname: "Graham",
      lastname: "Bell"
    },
    born: 1847,
    invention: "telephone"
  },
  {
    name: {
      firstname: "Ada",
      lastname: "Lovelace"
    },
    born: 1815,
    invention: "computer programming"
  }
])
```

Run the following `findOne()` operation to retrieve a single document from the collection:

```js
db.scientists.findOne({invention: "Turing Machine"})
```

## Retrieve all documents in a collection

The `find()` command is used for retrieveing all the documents in a collection.

```js
db.collection.find()
```

Run `db.scientists.find()` to see the complete list of documents in the collection.

### Retrieve documents based on a specific query

Using the `find()` command, you can also filter a collection for only the documents that match the provided query.
For example, find the document with the field `born` set as 1857.

```js
db.scientists.find({born: 1857})
```

### Retrieve documents using operator queries

The operator syntax allows users to query and retrieve a document.
There are several operator methods that you can use, such as `$gt` or `$lt`.
For example, to find the list of scientists born after the 1900s, we'll need the `$gt` operator:

```js
db.scientists.find({born:{$gt:1900}})
```

Here is a list of the most commonly used operators.

`$gt`: selects records that are greater than a specific value

`$lt`: selects records that are less than a specific value

`$gte`: selects records greater or equal to a specific value

`$lte`: selects records less than or equal to a specific value

`$in`: selects any record that contains any of the items present in a defined array

`$nin`: selects any record that does not contain any of the items in a defined array

`$ne`: selects records that are not equal to a specific value

`$eq`: select records that are equal to a specific value

### Retrieve documents containing a specific value in an array

Insert the following documents into an `employees` collection using this command:

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

To retrieve all documents with a specific array field and value (`catalog: "printer"`), run the following command:

```js
db.employees.find({catalog: "printer"})
```

The response displays all the retrieved documents:

```js
[
  {
    _id: ObjectId("636b39f80466c61a229bbf9b"),
    name: { first: 'Earl', last: 'Thomas' },
    employeeID: 1234,
    age: 23,
    role: 'salesperson',
    catalog: [ 'printer', 'cardboard', 'crayons', 'books' ]
  },
  {
    _id: ObjectId("636b3b0e0466c61a229bbf9d"),
    name: { first: 'Clarke', last: 'Dane' },
    employeeID: 3234,
    age: 21,
    role: 'salesperson',
    catalog: [ 'printer', 'pencils', 'crayons', 'toys' ]
  }
]
```

### Retrieve documents in an array using dot notation

To retrieve all documents containing a specific value in an array, use dot notation to reference its position in the `employees` collection.
The following command retrieves all documents containing `"blender"` in the third field of an array:

```js
db.employees.find({"catalog.2": "blender"})
```

The document that matches the array query is displayed in the response:

```js
[
  {
    _id: ObjectId("636b3b0e0466c61a229bbf9c"),
    name: { first: 'Sam', last: 'Johnson' },
    employeeID: 2234,
    age: 35,
    role: 'salesperson',
    catalog: [ 'cabinet', 'fridge', 'blender', 'utensils' ]
  }
]
```

### Query on an embedded or nested document

To query on an embedded document, use dot notation to specify the fields.
The following command queries on the embedded document in the`employees` collection:

```js
db.employees.find({"name.first": "Clarke"})
```
