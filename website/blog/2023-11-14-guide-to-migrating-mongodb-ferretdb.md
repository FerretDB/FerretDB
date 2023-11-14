---
slug: guide-to-migrating-mongodb-ferretdb
title: Guide to Migrating from MongoDB to FerretDB
authors: [alex]
description: >
  This blog post is a guide to help you migrate data from MongoDB to FerretDB.
image: /img/blog/ferretdb-v1.13.0.png
tags: [community, product, tutorial]
---

Intro

- Why migration to FerretDB is important (open source, no vendor lock-in, etc)
- _Reasons for migrating from MongoDB to FerretDB_
- _Scope and target audience of the guide_
- Compatibility considerations

This blog post is a comprehensive guide to help you migrate data from MongoDB to FerretDB.

- _-- Why migration to FerretDB is important (open source, no vendor lock-in, etc)_

In this guide, we are going to show you:

- How to prepare for migration to FerretDB
- How to migrate your data from MongoDB to FerretDB
- How to run your app successfully with FerretDB

## Prerequisites

- Current MongoDB URI
- `mongodump`/`mongorestore`, `mongoimport`/`mongoexport`

## Pre-Migration Planning

- Evaluating your current MongoDB setup
- Identifying potential challenges and solutions
- Backup strategies for MongoDB data
- Preparing your team for migration

Before migrating to FerretDB, you need to confirm its suitability for your application's use case.
While FerretDB is the ideal open-source alternative to MongoDB in many use cases, it's important to evaluate your current application set up to be sure that FerretDB supports all the essential features needed.

FerretDB provides different operation modes to run FerretDB: `normal`, `proxy`, `diff-normal`, `diff-proxy`.
These modes describe how FerretDB processes incoming requests.
The `diff` modes forward requests to the FerretDB and the FerretDB proxy, which is another MongoDB-compatible database, and log the difference between them.

To test your application against FerretDB, follow the guide on the [pre-migration planning process].

Once you are satisfied that the pre-migrating testing was successful, proceed to the next stage.

## Step-by-step guide to migrating data from MongoDB to FerretDB

- See docs for guidance: [https://docs.ferretdb.io/category/migrating-to-ferretdb/](https://docs.ferretdb.io/category/migrating-to-ferretdb/)
- Handling large datasets and downtime
- Guidance from Team

By now, it's assumed that you've set up a FerretDB instance, if you are yet to, please check our quickstart guide.
Before migration, you have to backup your existing data and then restore them on FerretDB.
To do this, use MongoDB native tools like `mongodump/mongorestore` and `mongoexport/mongoimport`.
Also, you'll need your connection string to perform this operation.

### Migrate your data using `mongodump`/`mongorestore`

To backup your existing data, you can use `mongodump` to pull all data.
Say you have a MongoDB instance with connection URI (`mongodb://127.0.0.1:27017`), you can dump the entire data by running:

```sh

mongodump --uri="mongodb://127.0.0.1:27017"

```

Once successful, this action creates a BSON file dump of all your data.
Note that when migrating your data, it's critical to securely do so by specifying the necessary authentication credentials.
And just in case you intend to only migrate a specific database or collection, specify them along with your connection uri.

For example, to dump data from a collection "testcoll" in "bigdb" database, run the following command:

```sh

mongodump --uri="mongodb://127.0.0.1:27017/" --nsInclude=bigdb.testcoll

```

To restore all the data from `mongodump` to your FerretDB instance, use `mongorestore`.
Specify the FerretDB connection string (e.g. `mongodb://127.0.0.1:27017/ferretdb?authMechanism=PLAIN`), including the authentication parameters.

```sh

mongorestore --uri="mongodb://127.0.0.1:27017/ferretdb?authMechanism=PLAIN"

```

To restore a specific database '"bigdb" and collection "testcoll", run the following command:

```sh

mongorestore --uri="mongodb://username:password@127.0.0.1:27018/?authMechanism=PLAIN" --nsInclude=bigdb.testcoll

```

### Migrate your data using `mongoexport`/`mongoimport`

Like `mongodump`/`mongorestore`, you can use `mongoexport`/`mongoimport` to migrate your data.
However, with `mongoexport`/`mongoimport`, there's no direct way to export all the collections at once.
To export your data, specify the connection string, the database and collection you wish to export, and the directory to export the data.

Suppose you have a collection "testcoll" and "bigdb" that you want to export, run the following command:

```sh

mongoexport --uri="mongodb://127.0.0.1:27017/" --db=bigdb --collection=testcoll --out=testcoll.json

```

To import the data using the `mongoimport`, run the following command:

```sh

mongoimport --uri="mongodb://username:password@127.0.0.1:27018/?authMechanism=PLAIN" --db=bigdb --collection=testcoll --file=testcoll.json

```

## Conclusion

In this guide, you've learned the crucial steps of migrating from MongoDB to FerretDB.
This process included ways to prepare for migration and how to ensure a successful migration is possible.
This included techniques for recognizing the main differences and compatibility aspects between MongoDB and FerretDB for your application.

We then delved into the technicalities of setting up FerretDB, and finally the steps to help you migrate your data using tools like `mongodump`/`mongorestore` and `mongoimport`/`mongoexport`.
For more detailed and technical insights on FerretDB, check out the [FerretDB official documentation](https://docs.ferretdb.io/).
