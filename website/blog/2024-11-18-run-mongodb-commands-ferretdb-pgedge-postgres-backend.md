---
slug: run-mongodb-commands-ferretdb-pgedge-postgres-backend
title: 'Run MongoDB workloads on FerretDB with pgEdge Postgres Platform as backend'
authors: [alex]
description: >
  Learn how to run MongoDB workloads in FerretDB with a fully distributed pgEdge PostgreSQL as the backend.
image: /img/blog/ferretdb-pgedge.jpg
tags: [compatible applications, tutorial, cloud, postgresql tools, open source]
---

![Run MongoDB commands on FerretDB with pgEdge Postgres as backend](/img/blog/ferretdb-pgedge.jpg)

FerretDB lets you run MongoDB workloads in open source using PostgreSQL as the backend.
It does this by converting BSON documents to JSONB in PostgreSQL.

<!--truncate-->

As such, you can leverage a fully managed, distributed PostgreSQL service like [pgEdge](https://www.pgedge.com/) as your database storage.
That will help you take advantage of its fast response time, high-availability, multi-cloud deployment options, and low data latency.

In this blog post we'll dive into a complete setup on how to run FerretDB with a fully distributed pgEdge PostgreSQL as the backend.

## Prerequisite

Before you start, ensure you have the following:

- [pgEdge account](https://www.pgedge.com/)
- [`mongosh`](https://www.mongodb.com/docs/mongodb-shell/)
- [Docker](https://www.docker.com/)

## Setting up Postgres with pgEdge

Before setting up a FerretDB instance, you need a Postgres database configured and ready.
Then you'll pass its connection string as the `FERRETDB_POSTGRESQL_URL`/ `--postgresql-url` flag when you run the FerretDB instance.

Start by setting up a Postgres database named `ferretdb` on pgEdge, and wait a few seconds for the database to be provisioned.

![pgedge connection](/img/blog/pgedge-connection.png)

Once it's ready, get the connection URI credentials for the `admin` user and use it as the `FERRETDB_POSTGRESQL_URL`/ `--postgresql-url` flag.

## Connect FerretDB to Postgres using Docker

Ensure you have Docker running.
Run the FerretDB container using the following command to connect via the `FERRETDB_POSTGRESQL_URL`.
Make sure to replace `<password>`, `<host>`, and `<port>` with your pgEdge connection details.

```sh
docker run -e FERRETDB_POSTGRESQL_URL='postgresql://admin:<password>@<host>/ferretdb?sslmode=require' -p 27017:27017 ghcr.io/ferretdb/ferretdb
```

Connect to FerretDB via `mongosh` using the following connection string:

```sh
mongosh "mongodb://admin:<password>@localhost/ferretdb?authMechanism=PLAIN"
```

Output should look like this:

```text
Current Mongosh Log ID: 672ce02bee320eddb90952bf
Connecting to:    mongodb://<credentials>@localhost:27017/ferretdb?authMechanism=PLAIN&directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.3.0
Using MongoDB:    7.0.42
Using Mongosh:    2.3.0
mongosh 2.3.3 is available for download: https://www.mongodb.com/try/download/shell
For mongosh info see: https://www.mongodb.com/docs/mongodb-shell/
------
   The server generated these startup warnings when booting
   2024-11-07T15:43:39.663Z: Powered by FerretDB v1.24.0 and PostgreSQL 16.4 on x86_64-pc-linux-gnu, compiled by gcc.
   2024-11-07T15:43:39.664Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2024-11-07T15:43:39.664Z: The telemetry state is undecided.
   2024-11-07T15:43:39.664Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.com.
------
ferretdb>
```

## Run MongoDB CRUD commands on FerretDB

Run basic MongoDB CRUD commands on your Postgres database using FerretDB:

Let's start by inserting a couple of interesting database books into a `books` collection.

```js
db.books.insertMany([
  {
    title: 'Introduction to Database Systems',
    author: 'C.J. Date',
    genre: 'Technical',
    publication_year: 1995,
    isAvailable: true
  },
  {
    title: 'Learning SQL',
    author: 'Alan Beaulieu',
    genre: 'Technical',
    publication_year: 2005,
    isAvailable: false
  }
])
```

Then let's run some quick `find` commands to check out our data and see the inserted book records.
For example, let find the book "Learning SQL" using its title:

```js
db.books.find({ title: 'Learning SQL' })
```

Output:

```json5
[
  {
    _id: ObjectId('672ce25eee320eddb90952c3'),
    title: 'Learning SQL',
    author: 'Alan Beaulieu',
    genre: 'Technical',
    publication_year: 2005,
    isAvailable: false
  }
]
```

You can update the collection using the `upsert` command which will update a document if it exists, or insert a new document if it does not.

Let's upsert a document where the title is "Database Systems Concepts," setting the isAvailable status to true:

```js
db.books.updateOne(
  { title: 'Database Systems Concepts' },
  {
    $set: {
      author: 'Silberschatz, Korth, and Sudarshan',
      genre: 'Technical',
      publication_year: 2002,
      isAvailable: true
    }
  },
  { upsert: true }
)
```

Next, let's run a `find` command with the `sort` option to list all books sorted by the publication_year in descending order.

```js
db.books.find({}).sort({ publication_year: -1 })
```

The entire collection is returned, sorted by the `publication_year` in descending order:

```json5
[
  {
    _id: ObjectId('672ce25eee320eddb90952c3'),
    title: 'Learning SQL',
    author: 'Alan Beaulieu',
    genre: 'Technical',
    publication_year: 2005,
    isAvailable: false
  },
  {
    _id: ObjectId('672ce807d624079c47cb2180'),
    title: 'Database Systems Concepts',
    author: 'Silberschatz, Korth, and Sudarshan',
    genre: 'Technical',
    isAvailable: true,
    publication_year: 2002
  },
  {
    _id: ObjectId('672ce25eee320eddb90952c2'),
    title: 'Introduction to Database Systems',
    author: 'C.J. Date',
    genre: 'Technical',
    publication_year: 1995,
    isAvailable: true
  }
]
```

As you can see, FerretDB lets you run many commands just as you would with MongoDB.
It translates BSON documents to JSONB under the hood.
Now let's see what that looks like on pgEdge.

![pgEdge dashboard](/img/blog/pgedge-dashboard.png)

Go ahead to try out FerretDB with pgEdge as your backend for many MongoDB use cases.
Here are some materials to help you start.

- Get started with [pgEdge Cloud Enterprise](https://docs.pgedge.com/cloud/getting_started/ee_getting_started)
- [Migrate your MongoDB workloads to FerretDB](https://docs.ferretdb.io/migration/migrating-from-mongodb/)
- [Quickstart guide for FerretDB](https://docs.ferretdb.io/quickstart-guide/docker/)
