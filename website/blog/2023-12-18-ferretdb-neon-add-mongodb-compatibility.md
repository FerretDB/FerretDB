---
slug: ferretdb-neon-add-mongodb-compatibility
title: 'FerretDB and Neon: Add MongoDB Compatibility to your Project'
authors: [alex]
description: >
  In this blog post, you’ll learn about FerretDB and how you can add MongoDB compatibility to Neon as its Postgres backend.
image: /img/blog/ferretdb-neon.jpg
tags: [tutorial, postgresql tools, open source]
---

![Start using with Neon](/img/blog/ferretdb-neon.jpg)

[FerretDB](https://www.ferretdb.com/) is an open source document database that adds MongoDB compatibility to other database backends, such as [Postgres](https://www.postgresql.org/) and [SQLite](https://www.sqlite.org/).
By using FerretDB, developers can [access familiar MongoDB features and tools using the same syntax and commands](https://blog.ferretdb.io/mongodb-crud-operations-with-ferretdb/) for many of their use cases.

<!--truncate-->

[Neon](https://neon.tech/) is an open source, fully managed serverless Postgres which is licensed under the Apache 2.0 license.
What makes Neon unique is that it separates storage and compute, which makes it possible to scale up compute based on demand.
It also supports branching, which allows you to create a copy of your database in seconds for testing and development.

Neon maintains 100% compatibility with any application that uses the official Postgres release.
That makes Neon a good option to consider as the database backend for FerretDB, while leveraging Neon's native features to scale and manage your database infrastructure easily.

In this blog post, you'll learn about FerretDB and and how you can add MongoDB compatibility to Neon.

## Advantages of using FerretDB

Let's take a look at some of these benefits of using FerretDB:

### MongoDB compatibility

The best advantage to using FerretDB is the access you get to the syntax, tools, querying language, and commands available in MongoDB, particularly for many common use cases.
MongoDB is known for its simple and intuitive NoSQL query language which is widely loved by many developers today.
By using FerretDB, you can enable Postgres-compatible databases like Neon to run MongoDB workloads.

Read more: [MongoDB Compatibility - What's Really Important?](https://blog.ferretdb.io/mongodb-compatibility-whats-really-important/)

### Open source

As an open source document database, you won't be at a risk of vendor lock-in.
Since MongoDB's license change to Server Side Public License (SSPL), there's [been a lot of disappointment and confusion with regards](https://blog.ferretdb.io/open-source-is-in-danger/) to how it affects many users and what it would mean for their applications.
According to the Open Source Initiative – the steward of open source and the set of rules that define open source software – SSPL is not considered open source.

In that regard, FerretDB, licensed under Apache 2.0, is a good option for users looking for a MongoDB alternative to return MongoDB workloads back to open source.

### Multiple backend options

At the moment, FerretDB supports Postgres and SQLite backends, with many ongoing works to support other backends.
Many databases built on Postgres can actually serve as a backend for FerretDB, including Neon.
That means you can take advantage of all the features available in that backend to scale and manage your database infrastructure without any fear of vendor lock-in.

## Prerequisites

- A Neon project and Postgres URI
- Docker
- `mongosh`

## Create a Neon project

Please refer to the instructions in the [Neon Sign up](https://neon.tech/docs/get-started-with-neon/signing-up) docs if this is your first project.

Before getting started with FerretDB, you need to create a Neon project.
When creating the project, you need to select the version of postgres, and also create a `ferretdb` database.

![Neon project creation](/img/blog/neon-create-project.png)

Once the project is created, Neon will provide a Postgres connection URI for you.
This URI will serve as the Postgres URI for your FerretDB instance.

## Install and connect FerretDB with Neon via Docker

To install the latest version of FerretDB, you'll need to create a `docker-compose` YAML file together with the Neon Postgres URI.

Assign the Neon Postgres URI to the `FERRETDB_POSTGRESQL_URL` environment variable.

```yaml
services:
  ferretdb:
    image: ghcr.io/ferretdb/ferretdb
    restart: on-failure
    environment:
      - FERRETDB_POSTGRESQL_URL=<NEON_URI>
    ports:
      - 27017:27017
```

From the same directory as the `docker-compose.yml` file, run `docker-compose up -d` to start the FerretDB instance
This will also pull the latest FerretDB image and run it in the background.

### Test via `mongosh`

With FerretDB running, let's test and see that it works using `mongosh`.
To connect via `mongosh`, you will need a connection string.
Use your Neon Postgres credentials to connect to the database.

So in this case, the MongoDB URI will be:

```sh
mongosh 'mongodb://<postgres-username>:<postgres-password>@127.0.0.1:27017/ferretdb?authMechanism=PLAIN'
```

This should connect you directly to the FerretDB instance, and you can go ahead to try out different MongoDB commands.

```text
~$ mongosh 'mongodb://<username>:<password>@127.0.0.1:27017/ferretdb?authMechanism=PLAIN'
Current Mongosh Log ID: 657c28296fda6bb93a0c0058
Connecting to:      mongodb://<credentials>@127.0.0.1:27017/?authMechanism=PLAIN&directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.0.2
Using MongoDB:      6.0.42
Using Mongosh:      2.0.2
mongosh 2.1.1 is available for download: https://www.mongodb.com/try/download/shell

For mongosh info see: https://docs.mongodb.com/mongodb-shell/

------
   The server generated these startup warnings when booting
   2023-12-15T10:19:28.991Z: Powered by FerretDB v1.17.0 and PostgreSQL 15.4 on x86_64-pc-linux-gnu, compiled by gcc.
   2023-12-15T10:19:28.991Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2023-12-15T10:19:28.991Z: The telemetry state is undecided.
   2023-12-15T10:19:28.991Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
------

ferretdb>
```

This will connect you directly to the `ferretdb` database.

#### Insert documents into FerretDB

With `mongosh` running, let's try to insert some documents into our FerretDB instance.
You are going to insert two footballer data into a `players` collection.

```json5
db.players.insertMany([
   {
       futbin_id: 3,
       player_name: "Giggs",
       player_extended_name: "Ryan Giggs",
       quality: "Gold - Rare",
       overall: 92,
       nationality: "Wales",
       position: "LM",
       pace: 90,
       dribbling: 91,
       shooting: 80,
       passing: 90,
       defending: 44,
       physicality: 67
   },
   {
       futbin_id: 4,
       player_name: "Scholes",
       player_extended_name: "Paul Scholes",
       quality: "Gold - Rare",
       overall: 91,
       nationality: "England",
       position: "CM",
       pace: 72,
       dribbling: 80,
       shooting: 87,
       passing: 91,
       defending: 64,
       physicality: 82,
       base_id: 246
   }
]);
```

Great!
Now when you run `db.players.find()`, it should return all the documents stored in the collection.

#### Update document record in FerretDB

Next, you need to update "Giggs" record to reflect his current position as a `CM`.
To do this, we can just run an `updateOne` command to target just that particular player:

```json5
db.players.updateOne(
    { player_name: "Giggs" },
    { $set: { position: "CM" } }
);
```

Let's query the collection to see if the changes have been made:

```text
ferretdb> db.players.find({player_name: "Giggs"})
[
  {
    _id: ObjectId("657c2bddc1aa97bc73fd3e7c"),
    futbin_id: 3,
    player_name: 'Giggs',
    player_extended_name: 'Ryan Giggs',
    quality: 'Gold - Rare',
    overall: 92,
    nationality: 'Wales',
    position: 'CM',
    pace: 90,
    dribbling: 91,
    shooting: 80,
    passing: 90,
    defending: 44,
    physicality: 67
  }
]
```

You can run many MongoDB operations on FerretDB.
See the [list of supported commands](https://docs.ferretdb.io/reference/supported-commands/) in the FerretDB documentation for more.

### View database on Neon

Besides having a document database view of the collection in FerretDB, you can also view the data through the Neon dashboard.

To view your current documents, go to the Neon dashboard and navigate to "Tables", then from the "schema" menu, select "ferretdb".
FerretDB stores the documents in Postgres as JSONB.

![Image showing ferretdb schema](/img/blog/neon-display-tables.png)

## Get started with FerretDB

FerretDB gives you a chance to run MongoDB workloads on Postgres and SQLite.
This flexibility means you can easily add MongoDB compatibility to Postgres-based relational databases like Neon, while avoiding vendor lock-in and retaining control of your data architecture.

To get started with FerretDB, check out the [FerretDB quickstart guide](https://docs.ferretdb.io/quickstart-guide/).
