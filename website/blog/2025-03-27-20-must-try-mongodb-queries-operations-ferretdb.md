---
slug: 20-must-try-mongodb-queries-operations-ferretdb
title: '20 must-try MongoDB operations and queries on FerretDB'
authors: [alex]
description: >
  FerretDB got better with the release of v2, bringing deeper MongoDB compatibility and enabling more advanced workloads to run complex queries for most use cases. Find out in this post.
image: /img/blog/mongodb-operations-ferretdb.jpg
tags: [document databases, community, product, open source]
---

![20 must-try MongoDB queries on FerretDB](/img/blog/mongodb-operations-ferretdb.jpg)

FerretDB just got better with [the release of v2.0 GA](2025-03-05-ferretdb-v2-ga-open-source-mongodb-alternative-ready-for-production.md), bringing deeper MongoDB compatibility and enabling more advanced workloads to run complex queries for most use cases.

<!--truncate-->

Beyond basic CRUD operations, most developers and businesses work with large datasets with analytical queries, text and vector indexing, geospatial searches, and replication, among other features.
All of these operations are available on FerretDB – in open source and without any danger of vendor lock-in or proprietary limitations.

In this post, we're diving into 20 must-try MongoDB features and operations you can run on FerretDB.
We will cover authentication, indexes, aggregation operations, geospatial queries, and more.

## Prerequisites

- FerretDB v2 instance running on your local machine or server.
  See the [installation guide](https://docs.ferretdb.io/installation/) for more.
- `mongosh`
- `mongoimport`

## Basic queries

If you're new to FerretDB, you might want to start with the basics – [understanding how to set up an instance](https://docs.ferretdb.io/installation/) and how to [run some basic CRUD operations](https://docs.ferretdb.io/usage/).
You can also check out this [MongoDB CRUD queries guide](2022-11-14-mongodb-crud-operations-with-ferretdb.mdx) before diving into the other sections in this post.

### 1. Insert documents with `mongoimport`

Without data, there's nothing to query!
So we will start by adding some documents to our database instance.

We already prepared a sample dataset to use; it's a collection of books with details like title, authors, genres, publication date, and price.
See the [books.fixture.json](https://raw.githubusercontent.com/FerretDB/FerretDB/refs/tags/v2.0.0/website/docs/guides/requests/books.fixture.json) file for the data.

Since we are using the FerretDB `books` collection, we will start by downloading the data and then importing it:

```sh
# Download the JSON data
curl -LJO "https://raw.githubusercontent.com/FerretDB/FerretDB/refs/tags/v2.0.0/website/docs/guides/requests/books.fixture.json"

# Import the data into FerretDB
mongoimport --uri "mongodb://username:password@localhost:27017/" --db library --collection books --file books.fixture.json --jsonArray
```

Now that you have some documents in the collection, start running some queries and operations on them.

### 2. Update documents

Prices change all the time – whether due to promotions, inflation, or publisher adjustments.
Updating an existing document helps keep your data relevant.

Due to a price drop for "Pride and Prejudice" from the initial price of `$19.99` to `$17.99`, let's update the document to reflect the change.
The `$set` operator lets you modify specific fields without touching the rest of the document:

```js
db.books.updateOne(
  { _id: 'pride_prejudice_1813' },
  { $set: { 'price.value': 17.99 } }
)
```

### 3. Sort documents

Sorting is crucial for presenting data in a meaningful way.
For instance, you might want to list books by publication date, price, or rating.

Let's sort books by publication date in descending order:

```js
db.books.find().sort({ 'publication.date': -1 })
```

### 4. Count documents

Are you curious about the number of books you have in the database?

`countDocuments()` gives you a quick total of all documents – quite handy when dealing with large datasets or checking if an import was successful:

```js
db.books.countDocuments()
```

This should return `5` as the count for all the documents in the collection.

### 5. Query nested array with `$elemMatch`

Books often have multiple authors, genres, and availability formats.
What if you only want books written by specific authors, e.g.
British authors?
`$elemMatch` helps filter specific values inside nested arrays.

```js
db.books.find({ authors: { $elemMatch: { nationality: 'British' } } })
```

### 6. Logical queries

Need to find books that match multiple conditions?
Logical operators let you chain queries, which is great for complex filtering.
Let's find books that are either in the "Romance" genre or were published before 1900:

```js
db.books.find({
  $or: [
    { genres: 'Romance' },
    { 'publication.date': { $lt: '1900-01-01T00:00:00Z' } }
  ]
})
```

### 7. Projection

When querying a database, sometimes you don't need everything – just specific fields.
Projection lets you control what's returned, reducing network load and making queries more efficient.
This is great when dealing with heavy fields like logs or analytics that aren't always needed.

Let's say you only need the `title` and `authors` of books – no need to pull in the entire document:

```js
db.books.find({}, { title: 1, authors: 1, _id: 0 })
```

Note that `1` means include, while `0` means exclude.
Also, `_id` is included by default, so we explicitly set it to `0` in the example so as to remove it.
This makes our queries faster and responses smaller – perfect when dealing with large datasets.

## Indexes

### 8. Create indexes and drop indexes

Speed is always a hot topic for databases.
Indexes help speed up queries, and that's crucial for better database performance.

Indexes create optimized data structures (sort of like a table of contents) that store references to document locations.
That way, queries can jump straight to the data via the document locations instead of scanning entire collections.

On FerretDB, you can create an index on a book's title and price:

```js
db.books.createIndex({ title: 1, 'price.value': -1 })
```

If you no longer need an index, you can drop it to free up space and resources.

For instance, to drop the index on the `title` field:

```js
db.books.dropIndex({ title: 1 })
```

Learn more about [indexes in FerretDB](https://docs.ferretdb.io/usage/indexes/).

### 9. TTL indexes

A Time-To-Live (TTL) index auto-deletes expired data after a set time.
That means no manual deletions and no bloated storage.

If a library tracks book reservations for instance, a TTL index can remove records after a year, so outdated holds don't clutter queries or waste space.

It is recommended to create the TTL index before inserting documents.
So if the document was added before the TTL index, it may not be affected until a new one is inserted.

Start by creating a TTL index on book reservations that expire after 60 seconds:

```js
db.books.createIndex({ 'reservation.date': 1 }, { expireAfterSeconds: 60 })
```

Next, let's insert a document with a reservation date:

```js
db.books.insertOne({
  _id: 'temporary_reservation',
  title: 'Reserved Book',
  reservation: {
    name: 'Noah Smith',
    date: new Date()
  }
})
```

After 60 seconds, the document will be automatically removed by the TTL index.

### 10. Partial indexes

Unlike full indexes, partial indexes only index documents that match a specific condition, and skips the rest.
Typically, this leads to smaller index sizes, faster writes, and optimized queries, making it perfect for filtering out irrelevant data without the overhead of a full index.

Suppose we only want to index books that cost more than $10, run the following command:

```js
db.books.createIndex(
  { genres: 1 },
  { partialFilterExpression: { 'price.value': { $gt: 10 } } }
)
```

### 11. Full-text search with `text` index

Regular indexes work great for exact matches or range queries but they struggle with searching inside text.
For instance, if you're looking for books with "romance novel" somewhere in the summary (`db.books.find({ "summary": "romance novel" })`), a standard index on "summary" won't help.
It expects an exact match and that's not good for full-text search.

Text indexes tokenize words and allow efficient full-text searches.
Create a text index for the summary field:

```js
db.books.createIndex({ summary: 'text' })
```

Using the earlier search for books mentioning "romance novel", regardless of position:

```js
db.books.find({ $text: { $search: 'romance novel' } })
```

Note that text indexes are case-insensitive and sorting by relevance can sometimes be unpredictable.

[Learn more about full-text search here.](https://docs.ferretdb.io/guides/full-text-search/)

### 12. Vector indexing for vector search

For more advanced similarity or context-based searches (like finding books with similar themes), vector indexing is useful.
It's a great choice if you're building recommendation systems, generative AI, or searching for related content based on embeddings.

FerretDB supports these vector indexes: Hierarchical Navigable Small World (HNSW) and Inverted File (IVF) indexes.

For example, let's create a vector index for the `summary` field using the HNSW algorithm with cosine similarity:

```js
db.runCommand({
  createIndexes: 'books',
  indexes: [
    {
      name: 'vector_hnsw_index',
      key: {
        vector: 'cosmosSearch'
      },
      cosmosSearchOptions: {
        kind: 'vector-hnsw',
        similarity: 'COS',
        dimensions: 12,
        m: 16,
        efConstruction: 64
      }
    }
  ]
})
```

For vector search, you need to generate embeddings for the field (e.g. `summary`) you want to search using any embedding model you prefer.
[See the FerretDB vector search guide for more.](https://docs.ferretdb.io/guides/vector-search/)

## Aggregation operations

### 13. Aggregation pipeline stages

Aggregation pipelines let you process and transform data in stages where each stage refines the result.
This is essential for analytics, reporting, and summarizing large datasets.

Let's say you need to find how many books belong to the "Classic" genre; `$match` filters books that have "Classic" in their `genres` array and `$count` gives the total number of matching documents.

```js
db.books.aggregate([
  { $match: { genres: 'Classic' } },
  { $count: 'classic_books' }
])
```

Find out more about aggregation pipelines in FerretDB [here](https://docs.ferretdb.io/usage/aggregation/).

### 14. Run analytical operations on FerretDB with `$group` and `$avg`

Say you want to analyze the average book rating per genre, which is a common use case for dashboards, trend analysis, or user recommendations:

```js
db.books.aggregate([
  {
    $group: {
      _id: '$genres',
      average_rating: { $avg: '$analytics.average_rating' }
    }
  }
])
```

### 15. Lookup (JOIN queries)

Sometimes, you need to combine data from multiple collections – just like SQL JOINs in relational databases.
This is useful when you store related data in different collections but want to retrieve them together in a single query.

Suppose there's another collection `publishers` with more details on the publishers, as shown below:

```js
db.publishers.insertMany([
  {
    _id: 1,
    name: 'T. Egerton',
    location: 'London, United Kingdom',
    established: 1780
  },
  {
    _id: 2,
    name: 'Harper & Brothers',
    location: 'New York, United States',
    established: 1817
  },
  {
    _id: 3,
    name: 'Lackington, Hughes, Harding, Mavor & Jones',
    location: 'London, United Kingdom',
    established: 1790
  }
])
```

You can join the data from both collections (`books` and `publishers`) using `$lookup`:

```js
db.books.aggregate([
  {
    $lookup: {
      from: 'publishers',
      localField: 'publication.publisher.name',
      foreignField: 'name',
      as: 'publisher_details'
    }
  }
])
```

This should fetch the books and their corresponding publishers' details in a single query.

## Geospatial queries

### 16. Find books published in a specific location

Some books aren't just about different places – they're from them.
Geospatial queries help you query location-based data, like books published in a specific city or country.
If a publisher's location is stored as GeoJSON points, FerretDB lets you query by geography instead of just text.

What if you want to search for books published in a specific city?
Instead of manually stating their locations, you can store and query precise geographic coordinates using `$geoWithin`.

Using London's longitude and latitude (`[-0.1276, 51.5072]`), let's run some queries to find books published within a `1km` radius.
Note that the distance is in radians, so we need to convert it to the Earth's radius.
Since the Earth's radius is about 6378.1km, we divide `1km` by the Earth's radius to get the distance in radians.

```js
db.books.find({
  'publication.publisher.location.geolocation': {
    $geoWithin: {
      $centerSphere: [[-0.1276, 51.5072], 1 / 6378.1]
    }
  }
})
```

## Diagnostic operations

### 17. Monitor operations with `currentOp`

Sometimes, your database operations become slow, and you wonder what's causing the delay.
`currentOp()` lets you inspect active operations, and find long-running queries that may overwhelm your instance.

To actually see something in `currentOp()`, we need some active operations that we can inspect.
Let's simulate a background operation by inserting 10,000 documents and running a `find` query that runs 1,000 times:

```js
db.books.insertMany([...Array(10000)].map((_, i) => ({ num: i }))) &&
  [...Array(1000)].forEach(() =>
    db.books.find({ num: { $gt: 5000 } }).toArray()
  )
```

To see all active operations, run the following query in another `mongosh` session:

```js
db.currentOp({ active: true })
```

This will show you all active operations, including the ones we just inserted and queried.

```js
{
  inprog: [
    {
      shard: 'defaultShard',
      active: true,
      type: 'op',
      opid: '10000000054:1742936581040255',
      op_prefix: Long('10000000054'),
      currentOpTime: ISODate('2025-03-25T21:03:01.000Z'),
      secs_running: Long('0'),
      command: { aggregate: '' },
      op: 'command',
      waitingForLock: false
    },
    {
      shard: 'defaultShard',
      active: true,
      type: 'op',
      opid: '10000000058:1742936581067302',
      op_prefix: Long('10000000058'),
      ns: 'library.books',
      currentOpTime: ISODate('2025-03-25T21:03:01.000Z'),
      secs_running: Long('0'),
      command: { getMor: Long('0') },
      op: 'getmore',
      waitingForLock: false
    }
  ],
  ok: 1
}
```

## Authentication

### 18. Create a user

Authentication is important for securing your database and ensuring only authenticated users can access it, especially for production-based environments.
FerretDB relies entirely on PostgreSQL authentication mechanisms.
When you create a user in FerretDB, you can manage it using the same commands you would on MongoDB.

You can create a user in FerretDB as you would on MongoDB by just running the `createUser` command.

For example, the following command creates the user `newuser` with the password `newpassword` with all the credentials stored on PostgreSQL:

```js
db.createUser({
  user: 'newuser',
  pwd: 'newpassword',
  roles: []
})
```

You can learn more about authentication in FerretDB [here](https://docs.ferretdb.io/security/authentication/).

### 19. Delete user

What if you no longer need a particular user account?
For example, you can remove the newly created `newuser` account on FerretDB simply by running the command.

```js
db.dropUser('newuser')
```

## Cleanup operations

### 20. Drop database

Once you are done with everything, you can proceed to drop the database, completely deleting it from the instance:

```js
db.dropDatabase()
```

## Get started with FerretDB

As you can see so far, FerretDB lets you run your MongoDB workloads in open-source using familiar syntax, commands, and features.
You don't have to change anything in your application.
Besides, there's no need to worry about vendor lock-in or any limitations that come with proprietary solutions.

[Get started with FerretDB today and see how it can power your MongoDB workloads](https://docs.ferretdb.io/).
