---
sidebar_position: 3
---

# Read operation

The read operation retrieves documents in a collection.
You can either retrieve all the documents in a collection, or only the documents that match a given query parameter.

## Retrieve all documents in a collection

The `find()` command is used for retrieveing all the documents in a collection.

```sh
db.collection.find()
```

First, populate the database with a new collection containing a list of documents.

```sh
db.scientists.insertMany([{name:{firstname: "Alan", lastname: "Turing"}, born: 1912, invention: "Turing Machine"},{name:{firstname: "Graham", lastname: "Bell"}, born: 1847, invention: "telephone"},{name:{firstname: "Ada", lastname: "Lovelace"}, born: 1815, invention: "computer programming"}])
```

Run `db.scientists.find()` to see the complete list of documents in the collection.

### Retrieve documents based on a specific query

Using the `find()` command, you can also filter a collection for only the documents that match the provided query.
For example, find the document with the field `born` set as 1857.

```sh
db.scientists.find({born: 1857})
```

### Retrieve documents using operator queries

The operator syntax allows users to query and retrieve a document.
There are several operator methods that you can use, such as `$gt` or `$lt`.
For example, to find the list of scientists born after the 1900s, we'll need the `$gt` operator:

```sh
db.scientists.find({born:{$gt:1900}})
```

Here is a list of the most commonly used operators.

`$gt`: selects records that are greater than a specific value

`$lt`:  selects records that are less than a specific value

`$gte`: selects records greater or equal to a specific value

`$lte`: selects records less than or equal to a specific value

`$in`: selects any record that contains any of the items present in a defined array

`$nin`: selects any record that does not contain any of the items in a defined array

`$ne`: selects records that are not equal to a specific value

`$eq`: select records that are equal to a specific value

## Retrieve a single document

The `findOne()` command retrieves a single document from a collection.

```sh
db.scientists.findOne({invention: "Turing Machine"})
```
