---
slug: migrate-mongodb-data-ferretdb-dsync
title: 'Migrating MongoDB Data to FerretDB with Dsync'
authors: [alex]
description: >

image: /img/blog/ferretdb-migration-dsync.jpg
tags: [open source, sspl, document databases, community]
---

![Migrating MongoDB Data to FerretDB with Dsync](/img/blog/ferretdb-migration-dsync.jpg)

If you're looking to migrate data from MongoDB to FerretDB, ensuring a smooth transition is crucial.

<!--truncate-->

Traditional migration workflows often rely on static dumps and manual restoration steps, which can often lead to issues and data loss.
BSON parsing errors.
Skipped collections.
Metadata mismatches.
Authentication quirks.
And worst of all – no clue where things went wrong.

`dsync` connects directly to both the source and destination of MongoDB-compatible services and streams data in real time.
You get a full initial sync followed by live replication, making it ideal for both one-time migrations and zero-downtime switchover.

In this post, you'll learn to use `dsync` to migrate the data from MongoDB to a running FerretDB instance.

## Prerequisites

Ensure to have the following:

- [Running FerretDB instance](https://docs.ferretdb.io/installation/ferretdb/)
- Running MongoDB instance (local or remote)

## Download `dsync`

Ensure to download the latest release of `dsync` from the [GitHub Releases page](https://github.com/adiom-data/dsync/releases/latest).
For Mac users, you may need to configure a security exception to execute the binary by [following these steps](https://support.apple.com/en-ca/guide/mac-help/mh40616/mac).

Alternatively, you can build dsync from the source code.

```sh
git clone https://github.com/adiom-data/dsync.git
cd dsync
go build
```

## Prepare a local MongoDB instance

For a quick test, we'll use a local MongoDB instance.
If you have one already running, you can skip this step.

```sh
rm -rf ~/temp/data_d
mkdir -p ~/temp/data_d
```

Start `mongod`:

```sh
mongod --dbpath ~/temp/data_d --port 27018 --replSet rs0
```

Leave this terminal open.

You need to initialize the replica set for `mongod` to work properly with `dsync`.
`dsync` requires a replica set to be set on the source for the migration to work.

Open a new terminal:

```sh
mongosh --port 27018
```

Then run:

```js
rs.initiate({
  _id: 'rs0',
  members: [{ _id: 0, host: 'localhost:27018' }]
})
```

Confirm it works:

```js
rs.status()
```

Let's load some sample data into our local MongoDB instance.
These sample data will be used to test the migration process.

Clone and restore the dataset:

```sh
git clone https://github.com/mcampo2/mongodb-sample-databases
cd mongodb-sample-databases
mongorestore --port 27018 --dir=dump/
```

Next, verify the sample data to ensure it was loaded correctly.

```sh
mongosh --port 27018
```

Then:

```js
use sample_mflix
db.movies.findOne()
```

You should see a document like:

```json
{ "_id" : ObjectId(...), "title" : "The Hunger Games", ... }
```

## Run `dsync` to migrate into FerretDB

Assuming FerretDB is running and listening on `localhost:27017`, specify the connection string (including the username and password) of your FerretDB instance if applicable and run `dsync`:

```bash
export MDB_SRC='mongodb://localhost:27018/sample_mflix'
export FERRETDB_DEST='mongodb://<username>:<password>@localhost:27017/sample_mflix'

./dsync --progress --logfile dsync.log "$MDB_SRC" "$FERRETDB_DEST"
```

Basically, `dsync` will open a live-change monitoring session in your terminal to track the progress of the migration.

```text
Dsync Progress Report : ChangeStream
Time Elapsed: 00:12:50        1/1 Namespaces synced
Processing change stream events
```

Next, we need to confirm that the data has been migrated successfully.

Connect to your FerretDB instance using `mongosh` or any MongoDB-compatible client.
If you're using `mongosh`, you can connect to FerretDB like this:

```sh
mongosh mongodb://<username>:<password>@localhost:27017/sample_mflix
```

Once connected, you can check the data:

```js
db.movies.countDocuments()
db.movies.find().limit(5)
```

You should now see the same data in your FerretDB instance.
As long as `dsync` remain running and connected, it will keep the data in sync between your local MongoDB instance and FerretDB.

## Start migrating

Now that you have a working setup, you can start migrating your data from MongoDB to FerretDB.
Traditional tools like `mongodump` and `mongorestore` are excellent – until you start working outside pure MongoDB environments.
FerretDB, DocumentDB, and CosmosDB introduce compatibility edge cases that make binary dumps fragile and unpredictable.

`dsync` provides a modern, wire-protocol-native alternative that's easier to script, debug, and extend.
If you're building CI pipelines or migrating test data into Mongo-compatible services, this approach gives you control without the overhead of managing `.bson` files or battling version mismatches.

It's a great way to ensure your data is in sync and ready for production.
