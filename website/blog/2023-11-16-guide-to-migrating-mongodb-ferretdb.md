---
slug: guide-to-migrating-mongodb-ferretdb
title: A Guide on Migrating from MongoDB to FerretDB
authors: [alex]
description: >
  This blog post is a guide to help you migrate data from MongoDB to FerretDB and run your application successfully.
image: /img/blog/migrating-mongodb-ferretdb.jpg
tags: [community, product, tutorial]
---

![A Guide on Migrating from MongoDB to FerretDB](/img/blog/migrating-mongodb-ferretdb.jpg)

This blog post is a guide to help you migrate data from MongoDB to FerretDB.

<!--truncate-->

More and more users favor open source software, as it provides a way to avoid vendor lock-in and reduce costs.
However, the complexity and cost of migration should also be a factor to consider.
Fortunately, with FerretDB, migration from a MongoDB-compatible database is easy.

According to [The 2023 State of Open Source Report](https://www.openlogic.com/resources/2023-state-open-source-report), over 80% of respondents reported an increase in the usage of open source software at their organization within the past year.
So for those looking to migrate from MongoDB to a truly open-source document database that works for many MongoDB use cases, FerretDB is a great option to consider.

In this guide, you'll learn to prepare for migration to FerretDB, how to migrate your data and successfully run your app with FerretDB.

## Best practices for migrating from MongoDB to FerretDB

### Identify reasons for migration

There are many reasons why you'd want to migrate to FerretDB.
For many, it is an opportunity to contribute to and influence the direction of open source projects; it could also be a way to escape the vendor lock-in associated with proprietary options.
For others, it may simply be a way to reduce costs.
Whatever the reason may be, it's important to identify it and ensure that FerretDB is the right fit for your use case.

### Plan your migration process

Of course, migration is not something you want to do in a hurry.
It requires careful planning and execution.
So it's important to plan your migration process to ensure that it goes smoothly.
You probably want to run FerretDB with your application in a test environment for a period of time before migrating your production data.
This will help you identify any potential migration issues.
You can also use this opportunity to test your application against FerretDB to ensure that it works as expected.
Check out the [pre-migration planning process](https://docs.ferretdb.io/migration/premigration-testing/) to FerretDB on how to achieve this.

### Evaluate your existing MongoDB setup with FerretDB

FerretDB is a good alternative for MongoDB in many use cases, but not all.
As such, it is probably a good idea to evaluate your existing MongoDB setup with FerretDB to ensure that it is a good fit for your application.
This approach would help you identify if there are any potential challenges that could impact migration.
If there are, it would be a nice starting point when communicating with the FerretDB team on finding a solution.

For example, FerretDB does not support all MongoDB features, so it is important to check that the features you need are supported ([see list of supported features here](https://docs.ferretdb.io/reference/supported-commands/)).
There are also some differences in usage that you should be aware of – [take a look at those differences here](https://docs.ferretdb.io/diff/).

### Backup your data

No matter the database you are migrating to, having a backup is always critical just incase something goes wrong.
Backing up your data helps ensure that you can always revert to your previous setup if something goes wrong during the migration process.

### Communicate with the FerretDB team

While this may not be necessary, it'll be great to communicate with the FerretDB team on your needs, especially if you have any concerns about the migration process.
The FerretDB team is always happy to help and can provide you with the necessary support to ensure a smooth migration process.
Besides, FerretDB has a growing community so you can also get help from the experience of other users.

Asides that, FerretDB is an open source software, so you can always improve it by contributing to the code yourself, submitting bug reports, or even requesting new features.
That's the joy of open source!

## A guide to migrating data from MongoDB to FerretDB

### Prerequisites

Before starting any migration processs, you should have the following:

- MongoDB connection URI
- FerretDB connection URI
- MongoDB native tools (`mongodump`/`mongorestore`, `mongoimport`/`mongoexport`)

### Step 1: Pre-migration planning

Before migrating to FerretDB, you need to confirm its suitability for your application's use case.
While FerretDB is the ideal open-source alternative to MongoDB in many use cases, it's important to evaluate your current application set up to be sure that FerretDB supports all the essential features needed.

FerretDB provides different operation modes to run FerretDB: `normal`, `proxy`, `diff-normal`, `diff-proxy`.
These modes describe how FerretDB processes incoming requests.
The `diff` modes forward requests to the FerretDB and the FerretDB proxy, which is another MongoDB-compatible database, and log the difference between them.

To test your application against FerretDB, follow the guide on the [pre-migration planning process](https://docs.ferretdb.io/migration/premigration-testing/).

Once you are satisfied that the pre-migration testing was successful, proceed to the next stage.

### Step 2: Set up FerretDB

There are different ways to set up FerretDB, depending on your use case and environment.
FerretDB provides Postgres as a database backend, so you can run it on any environment that supports Postgres.
That could be running it on a local machine, a Docker container, on the cloud – anywhere you want.

For more information on how to set up FerretDB, check out the [FerretDB official documentation](https://docs.ferretdb.io/quickstart-guide/).

FerretDB is also available as a managed service on [Civo](https://www.civo.com/marketplace/FerretDB), [Scaleway](https://www.scaleway.com/en/managed-document-database/), and [Vultr](https://www.vultr.com/products/managed-databases/ferretDB/), so you can easily get started with FerretDB without having to worry about the setup process.

### Step 3: Backup and restore data using MongoDB native tools

By now, it's assumed that you've set up a FerretDB instance.
Before migration, you should backup your existing data and then restore them on FerretDB.
To do this, use MongoDB native tools like `mongodump/mongorestore` and `mongoexport/mongoimport`.
You'll need your existing MongoDB connection URI for this operation.

#### Migrate your data using `mongodump`/`mongorestore`

To backup your existing data, you can use `mongodump` to pull all data.
Say you have a MongoDB instance with connection URI (`mongodb://127.0.0.1:27017`), you can dump the entire data by running:

```sh
mongodump --uri="mongodb://127.0.0.1:27017"
```

Once successful, this action creates a BSON file dump of all your data.
Note that when migrating your data, it's critical to securely do so by specifying the necessary authentication credentials.
And just in case you intend to only migrate a specific database or collection, specify them along with your connection uri.

For example, to dump data from a collection "testcoll" in "maindb" database, run the following command:

```sh
mongodump --uri="mongodb://127.0.0.1:27017/" --nsInclude=maindb.testcoll
```

To restore all the data from `mongodump` to your FerretDB instance, use `mongorestore`.
Specify the FerretDB connection string (e.g. `mongodb://127.0.0.1:27017/ferretdb?authMechanism=PLAIN`), including the authentication parameters.

```sh
mongorestore --uri="mongodb://127.0.0.1:27017/ferretdb?authMechanism=PLAIN"
```

To restore a specific database '"maindb" and collection "testcoll", run the following command:

```sh
mongorestore --uri="mongodb://username:password@127.0.0.1:27018/?authMechanism=PLAIN" --nsInclude=maindb.testcoll
```

#### Migrate your data using `mongoexport`/`mongoimport`

Like `mongodump`/`mongorestore`, you can use `mongoexport`/`mongoimport` to migrate your data.
However, with `mongoexport`/`mongoimport`, there's no direct way to export all the collections at once.
To export your data, specify the connection string, the database and collection you wish to export, and the directory to export the data.

Suppose you have a collection "testcoll" and "maindb" that you want to export, run the following command:

```sh
mongoexport --uri="mongodb://127.0.0.1:27017/" --db=maindb --collection=testcoll --out=testcoll.json
```

To import the data using the `mongoimport`, run the following command:

```sh
mongoimport --uri="mongodb://username:password@127.0.0.1:27018/?authMechanism=PLAIN" --db=maindb --collection=testcoll --file=testcoll.json
```

## Start your journey with FerretDB

In this guide, you've learned the crucial steps of migrating from MongoDB to FerretDB, including the best practices, and how to ensure a successful migration.
All software migrations come with their own challenges, so it's important to be prepared for them.
But with the right preparation, you can ensure a smooth migration process.
Besides, the FerretDB team is always happy to help you with any issues you discover or requests you may have.
Contact us on any of our [community channels](https://docs.ferretdb.io/#community)!

For more detailed and technical insights on FerretDB, check out the [FerretDB official documentation](https://docs.ferretdb.io/).
