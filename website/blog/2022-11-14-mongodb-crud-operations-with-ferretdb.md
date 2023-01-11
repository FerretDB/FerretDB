---
slug: mongodb-crud-operations-with-ferretdb
title: "How to Pass Basic MongoDB CRUD Operations With FerretDB"
author: Alexander Fashakin
date: 2022-11-14
---

![Pass Basic CRUD Operations in FerretDB](https://www.ferretdb.io/wp-content/uploads/2022/11/uriel-sc-11KDtiUWRq4-unsplash-1024x680.jpg)

<!--truncate-->

As MongoDB moves away from its open-source roots with SSPL, developers and tech enthusiasts are on the lookout for a truly open-source alternative to help manage and execute NoSQL operations.

[FerretDB](https://www.ferretdb.io/ "") is an open-source proxy that converts MongoDB NoSQL commands and queries to SQL.
You won't have to learn a new syntax or query method.
With FerretDB, you can easily execute and pass MongoDB operations.

This blog post will show you how to pass basic MongoDB CRUD operations using FerretDB.

## Perform MongoDB CRUD operations with FerretDB

CRUD (Create, Read, Update, and Delete) operations sit at the heart of any database management system.
They allow users to easily interact with the database to create, sort, filter, modify, or delete records.

For users looking for an open-source MongoDB alternative, FerretDB offers you a direct replacement where you can still use all your favorite MongoDB CRUD methods and commands, without having to learn entirely new commands.

### How to set up FerretDB database

To set up FerretDB locally, you can install it using [Docker](https://www.docker.com/ "") or the .deb and .rpm packages available for each [release](https://github.com/FerretDB/FerretDB/releases "").
Follow the [quickstart instructions on Github](https://github.com/FerretDB/FerretDB#quickstart "") to quickly get started.

In the same way as MongoDB, you can show the list of databases with FerretDB using the following command:

```js
show dbs
```

Similarly, the `use your_database_name` command allows you to switch from one database to another.
And if the database does not exist, FerretDB creates a new database.

```js
use league
```

If there’s no existing database with this name, a new database (**league**) is created in your FerretDB storage backend on PostgreSQL.
Read on to learn more about all the basic MongoDB commands or operations you can perform with FerretDB.

## Create operation

Much like MongoDB, FerretDB provides the `insertOne()` and `insertMany()` method for you to add new records to a collection.

### insertOne()

Using the database we created earlier, let’s insert a single document record with fields and values into the collection by calling the `insertOne()` method.
The syntax for this operation looks like this:

```js
db.collection_name.insertOne({field1: value1, field2: value2, …fieldN: valueN})
```

The process is identical to how you’d insert a single data record in MongoDB.
To begin with, let’s insert a single document into the collection:

```js
db.league.insertOne({club:'PSG', points: 30, average_age: 30, discipline:{red:5, yellow:30},qualified: false})
```

This line of code creates a new document in your collection.
If the operation is successful, you’ll get a  response with `acknowledged` set as ‘true’, and the id of the inserted document  (`insertedID`) containing the ObjectId.

```js
{
  acknowledged: true,
  insertedId: ObjectId("63109e9251bcc5e0155db0c2")
}
```

Note that if a collection does not exist, the insert command will automatically create one for you.

### insertMany()

Instead of creating just a single document in the collection, you can create multiple documents with the `insertMany()` method.
Indicate that you are inserting multiple document records with a square bracket and separate each document using commas.

```js
db.collection_name.insertMany([{document1}, {document2},... {documentN}])
```

To see how this works, we are going to insert three new documents into the collection:

```js
db.league.insertMany([{club:'Arsenal', points: 80, average_age: 24, discipline: {red: 2, yellow: 15},qualified: true }, {club:'Barcelona', points: 60, average_age: 31, discipline: {red: 0, yellow: 7},qualified:false}, {club:'Bayern', points: 84, average_age: 29 , discipline: {red: 1, yellow: 20}, qualified:true}])
```

The output from this command:

```js
{
  acknowledged: true,
  insertedIds: {
    '0': ObjectId("63109f4d51bcc5e0155db0c3"),
    '1': ObjectId("63109f4d51bcc5e0155db0c4"),
    '2': ObjectId("63109f4d51bcc5e0155db0c5")
  }
}
```

## Read operation

The read operation in FerretDB is similar to those of MongoDB.
We'll be exploring two basic methods for querying records in a collection: `find()` and `findOne()`.

### find()

In every database operation, you'll need to filter the documents or collection for records based on some specific queries.
The `find()` method filters and selects all the documents in a collection that matches the query parameters.

If there is no query parameter for the method, all the records present in the collection are returned.
First, let’s select all the documents in the **league** collection created earlier.

```js
db.league.find({})
```

This operation retrieves and displays all the documents present in the collection.
Now, let's add a query parameter to the `find()` operation to filter for a specific item.

```js
db.league.find({club:'PSG'})
```

This retrieves all the records that match the query:

```js
[
  {
    _id: ObjectId("63109e9251bcc5e0155db0c2"),
    club: 'PSG',
    points: 30,
    average_age: 30,
    discipline: { red: 5, yellow: 30 },
    qualified: false
  }
]
```

You can also filter a collection in FerretDB using any of the commonly used MongoDB operators:

* `$gt`: selects records that are greater than a specific value
* `$lt`:  selects records that are less than a specific value
* `$gte`: selects records greater or equal to a specific value
* `$lte`: selects records less than or equal to a specific value
* `$in`: selects any record that contains any of the items present in a defined array
* `$nin`: selects any record that does not contain any of the items in a defined array
* `$ne`: selects records that are not equal to a specific value
* `$eq`: select records that are equal to a specific value

#### Find documents using the `$in` operator

Say we want to query the collection for documents that contain any of the values present in an array.
To do this, we'll filter the document using an array of values with the `$in` operator:

```js
db.collection_name.find({ field: { $in: [<value1>, <value2>, ... <valueN> ] } })
```

Let's filter the `league` data for teams with 80 or 60 `points`:

```js
db.league.find({points:{$in:[80,60]}})
```

This displays the documents that match this query:

```js
[
  {
    _id: ObjectId("63109f4d51bcc5e0155db0c3"),
    club: 'Arsenal',
    points: 80,
    average_age: 24,
    discipline: { red: 2, yellow: 15 },
    qualified: true
  },
  {
    _id: ObjectId("63109f4d51bcc5e0155db0c4"),
    club: 'Barcelona',
    points: 60,
    average_age: 31,
    discipline: { red: 0, yellow: 7 },
    qualified: false
  }
]
```

#### Find documents using the `$lt` operator

The `$lt` operator filters a collection for records that are less than a specific value.
For example, let's select the documents with less than 60 *points* :

```js
db.league.find({points:{$lt:60}})
```

The output:

```js
[
  {
    _id: ObjectId("63109e9251bcc5e0155db0c2"),
    club: 'PSG',
    points: 30,
    average_age: 30,
    discipline: { red: 5, yellow: 30 },
    qualified: false
  }
]
```

### findOne()

The `findOne()` method selects the first document that matches a specified set of query parameters.
For instance, let’s filter the collection for documents with the *qualified* set to true.

```js
db.league.findOne({qualified:true})
```

The response displays the first document that matches the query:

```js
{
  _id: ObjectId("63109f4d51bcc5e0155db0c3"),
  club: 'Arsenal',
  points: 80,
  average_age: 24,
  discipline: { red: 2, yellow: 15 },
  qualified: true
}
```

Even though two documents match this query, the result only displays the first document.

## Update operation

Update operations are write commands that accept a query parameter and changes to be applied to a document.

We'll be exploring  three basic MongoDB methods for updating documents using FerretDB: `updateOne()`, `updateMany()`, and `replaceOne()`.

### updateOne()

The `updateOne()` method uses a query parameter to filter and then update a single document in a collection.
The following syntax depicts the update operation where the atomic operator `$set` contains the new record:

```js
db.collection_name.updateOne({<query-params>}, {$set: {<update fields>}})
```

Using our database, let’s update a document where the `club` field is  `PSG` and set the new `points` field as `35`.
This update operation will only affect the first document that’s retrieved in the collection.

```js
db.league.updateOne({club:'PSG'}, {$set: {points:35}})
```

If this operation is successful, the queried document will be updated.

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 1,
  modifiedCount: 1,
  upsertedCount: 0
}
```

### updateMany()

The `updateMany()` method can take a query and make updates to many documents at once.
For example, let’s update all documents with a *points* field that’s less than or equal to 90 and set the `qualified` field to false.

```js
db.league.updateMany({points:{$lte: 90}}, {$set: {qualified:false}})
```

```js
{
  acknowledged: true,
  insertedId: null,
  matchedCount: 4,
  modifiedCount: 2,
  upsertedCount: 0
}

```

### replaceOne()

The replaceOne() method is ideal if you intend to replace an entire document at once.

```js
db.league.replaceOne({club: "Barcelona"}, {club:'Inter', points: 83, average_age: 32, discipline:{red:2, yellow:10},qualified: true})
```

If we run `db.league.find({})`, our database now looks like this:

```js
[
  {
    _id: ObjectId("63109e9251bcc5e0155db0c2"),
    club: 'PSG',
    points: 35,
    average_age: 30,
    discipline: { red: 5, yellow: 30 },
    qualified: false
  },
  {
    _id: ObjectId("63109f4d51bcc5e0155db0c3"),
    club: 'Arsenal',
    points: 80,
    average_age: 24,
    discipline: { red: 2, yellow: 15 },
    qualified: false
  },
  {
    _id: ObjectId("63109f4d51bcc5e0155db0c5"),
    club: 'Bayern',
    points: 84,
    average_age: 29,
    discipline: { red: 1, yellow: 20 },
    qualified: false
  },
  {
    _id: ObjectId("63109f4d51bcc5e0155db0c4"),
    club: 'Inter',
    points: 83,
    average_age: 32,
    discipline: { red: 2, yellow: 10 },
    qualified: true
  }
]
```

## Delete operation

The delete operations are operations that only affect a single collection.
Let’s check out two methods for deleting documents in a collection: deleteOne() and deleteMany().

### deleteOne()

The `deleteOne()` method takes in a query parameter that filters a collection for a particular document and then deletes it from the record.
Note that this operation only deletes the first document that matches the query in the collection.

```js
db.league.deleteOne({club:'Arsenal'})
```

This operation deletes one document from the collection:

```js
{ acknowledged: true, deletedCount: 1 }
```

### deleteMany()

The deleteMany() method is used for deleting multiple documents in a collection.
The operation takes in a query and then filters and deletes all the documents matching the query.

```js
db.league.deleteMany({qualified: false})
```

Run `db.league.find({})` to show the current state of records of the database.

```js
[
  {
    _id: ObjectId("63109f4d51bcc5e0155db0c4"),
    club: 'Inter',
    points: 83,
    average_age: 32,
    discipline: { red: 2, yellow: 10 },
    qualified: true
  }
]
```

## Get started with FerretDB

Voila!
Just the exact MongoDB replacement you’ve been looking for.

Beyond the basic CRUD operations in this post, you can pass even more complex MongoDB commands without having to reinvent the wheel to learn new commands or give up the option of using open-source software.

[FerretDB](https://www.ferretdb.io/) serves as a truly open-source replacement for MongoDB.
That means you don’t have to sacrifice the integrity and benefits of open-source software while still enjoying the benefits of a non-relational NoSQL database.

To know more about the importance of an open-source alternative to MongoDB, read [this article](https://www.ferretdb.io/open-source-is-in-danger/).

Photo by [Uriel SC](https://unsplash.com/@urielsc26?utm_source=unsplash&amp;utm_medium=referral&amp;utm_content=creditCopyText) on [Unsplash](https://unsplash.com/?utm_source=unsplash&amp;utm_medium=referral&amp;utm_content=creditCopyText)
