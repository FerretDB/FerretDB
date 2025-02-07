---
slug: 20-must-try-mongodb-queries-operations-ferretdb
title: '20 Must-Try Advanced MongoDB Queries on FerretDB'
authors: [alex]
description: >
  FerretDB just got better with the release of v2, bringing deeper MongoDB compatibility, and enabling more advanced workloads to run complex queries for most use cases. Find out in this post.
tags: [observability, product, open source]
---

![20 Must ](/img/blog/ferretdb-otel/opentelemetry.jpg)

FerretDB just got better with the release of v2, bringing deeper MongoDB compatibility, and enabling more advanced workloads to run complex queries for most use cases.

<!--truncate-->

Beyond just basic CRUD operations, most developers and businesses work with large datasets with real-time analytical queries, text and vector indexing, geospatial searches, replication, among other features.
And these operations are all available on FerretDB – all in open source, without any danger of vendor lock-in or proprietary limitations.

In this post, we're diving into 20 advanced MongoDB queries, each showcasing a real use case.
By the end, if you're working with FerretDB or considering making the switch, these queries will show you exactly what's possible.

you'll know how to leverage FerretDB for complex searches, text indexing, and aggregations that would normally require specialized tools.
Basic Queries
If you're new to FerretDB, you might want to start with the basics – understanding how to set up an instance and how it handles standard MongoDB operations.
You can check out this MongoDB CRUD queries guide before diving into the other sections in this post.

## Basic Queries

If you're new to FerretDB, you might want to start with the basics – understanding how to set up an instance and how it handles standard MongoDB operations.
You can check out this MongoDB CRUD queries guide before diving into the other sections in this post.

### 1. Insert document with `mongoimport`

We will start with adding documents to our database instance – without data, there's nothing to query!

Since we are using the FerretDB `Books` collection, let's start by downloading the data and then importing it.

The following command inserts documents into books collection in database `library`.

```sh
# Download the JSON data
curl -L -o books_fixture.json "https://raw.githubusercontent.com/FerretDB/FerretDB/main/website/docs/guides/requests/books.fixture.json"

# Import the data into FerretDB
mongoimport --host localhost --port 27017 --db library --collection books --file books_fixture.json --jsonArray
```

Now that you have some documents in the collection, start running some queries and operations on them.

### 2. Update Documents

Prices change all the time – whether due to promotions, inflation, or publisher adjustments.
Updating an existing document helps keep your data relevant.

Due to a price drop for "Pride and Prejudice" from the initial price of `$19.99` to `$17.99`, you need to update the document to reflect the change.
The `$set` operator is your best friend here – it lets you modify specific fields without touching the rest of the document.

```js
db.books.updateOne(
  { _id: 'pride_prejudice_1813' },
  { $set: { 'price.value': 17.99 } }
)
```

### 3. Sort Documents

Sorting is crucial for presenting data in a meaningful way.
For instance, you might want to list books by publication date, price, or rating.

Let's sort books by publication date in descending order:

```js
db.books.find().sort({ 'publication.date': -1 })
```

### 4. Count Documents

Are you curious about the number of books you have in the database?

`countDocuments()` gives you a quick total of all documents – quite handy when dealing with large datasets or checking if an import was successful.

```js
db.books.countDocuments()
```

This should return `5` as the count for all the documents in the collection.

### 5. $elemMatch: Nested Array Queries

Books often have multiple authors, genres, and availability formats.
What if you only want books written by British authors?
`$elemMatch` helps filter specific values inside nested arrays.

```js
db.books.find({ authors: { $elemMatch: { nationality: 'British' } } })
```

### 6. Logical Queries: $and, $or

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

### 7. Lookup (Join Queries)

**TBD**.

```js
db.books.aggregate([
  {
    $lookup: {
      from: 'reviews',
      localField: '_id',
      foreignField: 'book_id',
      as: 'book_reviews'
    }
  }
])
```

## Authentication

### 8. Create, List, and Login Users

Databases often need authentication, especially in production.
FerretDB relies on PostgreSQL authentication mechanisms and does not store any user credentials.

You can create a user in FerretDB as you would on MongoDB by just running the `createUser` command.

For example, the following command creates a `newuser` with password `newpassword` – the credentials are all stored on PostgreSQL.

```js
db.createUser({
  user: 'newuser',
  pwd: 'newpassword',
  roles: []
})
```

In the same way, you could create users directly within PostgreSQL itself by setting a new user with SQL:

```sql
CREATE USER newuser WITH PASSWORD 'newpassword';
```

Now, if you want to see existing users:

```js
db.getUsers()
{
  users: [
    {
      _id: 'admin.newuser',
      userId: 'admin.newuser',
      user: 'newuser',
      db: 'admin',
      roles: [
        { role: 'readWriteAnyDatabase', db: 'admin' },
        { role: 'clusterAdmin', db: 'admin' }
      ]
    }
  ],
  ok: 1
}
```

You can login to the account the same way as you would with the initial user credential set on Postgres.
Authentication is especially useful for production environments where you need to manage multiple users with different permission levels.

### 9. Delete User

What if you no longer need a particular user account?
(for example, the newly created `newuser` account), you can remove it using:

```js
db.dropUser('newuser')
```

Be careful with this command – once deleted, the user is gone, and you'll need to recreate it if needed.

## Indexes

### 10. Create Indexes

Speed is always a hot topic for databases – indexes speed up queries, and that's crucial for better database performance.

Indexes create optimized data structures (sort of like a table of contents) that store references to document locations.
That way, queries can jump straight to the data via the document locations instead of scanning entire collections.

A basic example is indexing a book's title and price for faster retrievals:

```js
db.books.createIndex({ title: 1, 'price.value': -1 })
```

### 11. 12.Drop indexes

If you no longer need an index, you can drop it to free up space and resources.

For instance, to drop the index on the `title` field:

```js
db.books.dropIndex({ title: 1 })
```

### Partial Indexes

Unlike full indexes, partial indexes only index documents that match a specific condition, skipping the rest.
This means smaller index sizes, faster writes, and optimized queries – perfect for filtering out irrelevant data without the overhead of a full index.

Suppose we only want to index books that cost more than $10:

```js
db.books.createIndex(
  { genres: 1 },
  { partialFilterExpression: { 'price.value': { $gt: 10 } } }
)
```

### 13. Full-Text Search with Text Indexing

Regular indexes work great for exact matches or range queries but they struggle with searching inside text.
For instance, if you're looking for books with "classic" somewhere in the summary (`db.books.find({ "summary": "romance novel" })`), a standard index on "summary" won't help.
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

Text indexes are case-insensitive and sorting by relevance can sometimes be unpredictable.

### 14. Vector Indexing for Vector Search

For more advanced similarity or context-based searches (like finding books with similar themes), vector indexing is useful.
It's a great choice if you're building recommendation systems, generative AI, or searching for related content based on embeddings.

FerretDB supports these vector indexes: Hierarchical Navigable Small World (HNSW) and Inverted File (IVF) indexes.

For example:

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

You'll need to generate vector embeddings for a particular field (`summary`) using any embedding model you prefer.
See the FerretDB vector search guide for more.

### 15. Projection

When querying a database, sometimes you don't need everything – just specific fields.
Projection lets you control what's returned, reducing network load and making queries more efficient.
This is great when dealing with heavy fields like logs or analytics that aren't always needed.

Let's say you only need the `title` and `authors` of books – no need to pull in the entire document:

```js
db.books.find({}, { title: 1, authors: 1, _id: 0 })
```

Note that 1 means include, while 0 means exclude.
Also, `_id` is included by default, so we explicitly set it to 0 to remove it.
This makes queries faster and responses smaller – perfect when dealing with large datasets.

You can also use projections in FerretDB to exclude fields.
Need all book details except reviews?
Use projection to exclude just that field:

```js
db.books.find({}, { reviews: 0 })
```

## Aggregation Operations

### 16. Aggregation Pipeline Stages (Match, Count)

Aggregation pipelines let you process and transform data in stage where each stage refines the result.
This is essential for analytics, reporting, and summarizing large datasets.

Let's say you need to find how many books belong to the "Classic" genre; `$match` filters books that have "Classic" in their `genres` array and `$count` gives the total number of matching documents.

```js
db.books.aggregate([
  { $match: { genres: 'Classic' } },
  { $count: 'classic_books' }
])
```

### 17. Analytics Aggregation

Say you want to analyze the average book rating per genre, which is a common use case for dashboards, trend analysis, or user recommendations.

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

## Geospatial Queries

### 18. Find Books Published in a Specific Location

Some books aren't just about different places – they're from them.
If a publisher's location is stored as GeoJSON points, FerretDB lets you query by geography instead of just text.

What if you want to search for books published in a specific city.
Instead of manually tagging locations, you can store and query precise geographic coordinates using `$geoWithin`:

Using London's longitude and latitude (`[-0.1276, 51.5072]`), let's run some queries to find books published there.

```js
db.books.find({
  'publication.publisher.location.geolocation': {
    $geoWithin: {
      $centerSphere: [[-0.1276, 51.5072], 1 / 6378.1]
    }
  }
})
```

## Time Series collections

### 19. Create Time Series Queries

Time-series data is everywhere – think stock prices, weather data, sensor readings, or even tracking book sales over time.
FerretDB supports time-series collections, making it easier to store and analyze time-based data efficiently.
Let's create a time-series collection to track book sales over time:

```js
db.createCollection('book_sales', {
  timeseries: {
    timeField: 'sale_date',
    metaField: 'book_id',
    granularity: 'hours'
  }
})
```

Now, let's insert some sales data:

```js
db.book_sales.insertMany([
  {
    book_id: 'pride_prejudice_1813',
    sale_date: new Date('2024-02-01T10:00:00Z'),
    copies_sold: 5
  },
  {
    book_id: 'moby_dick_1851',
    sale_date: new Date('2024-02-01T11:00:00Z'),
    copies_sold: 3
  }
])
```

Now let's find the book sales in a specific time range:

```js
db.book_sales.find({
  sale_date: {
    $gte: new Date('2024-02-01T00:00:00Z'),
    $lt: new Date('2024-02-02T00:00:00Z')
  }
})
```

This fetches all sales within February 1st, 2024 – perfect for daily reports.

If you're working with historical trends, forecasting, or real-time monitoring, time-series collections are quite handy!

### 20. Drop Database

Once you are done with everything, you can proceed to drop the database, completely deleting it from the instance.

```js
db.dropDatabase()
```

## Get Started with FerretDB

As you can see so far, FerretDB lets you run your MongoDB workloads in open-source using familiar syntaxes and commands.
You no longer need to worry about vendor lock-in or any limitations that come with proprietary solutions.

Get started with FerretDB here and see how it can power your MongoDB workloads.
