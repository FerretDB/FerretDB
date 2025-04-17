---
slug: migrate-mongodb-data-ferretdb-dsync
title: 'Migrating MongoDB Data to FerretDB with dsync'
authors: [alex]
description: >
  Learn how to migrate your MongoDB data to FerretDB using dsync — with zero downtime and no data loss.
image: /img/blog/ferretdb-migration-dsync.jpg
tags: [open source, sspl, document databases, community]
---

![Migrating MongoDB Data to FerretDB with dsync](/img/blog/ferretdb-migration-dsync.jpg)

If you're looking to migrate data from MongoDB to FerretDB, ensuring a smooth transition is crucial.

<!--truncate-->

Traditional migration workflows often rely on static dumps and manual restoration steps, which can often lead to complications.
Skipped collections.
Metadata mismatches.
Data loss.
And worst of all – no clue where things went wrong.

[dsync](https://github.com/adiom-data/dsync/) by [adiom](https://adiom.com/) is a tool that connects directly to both the source and destination of MongoDB-compatible services and streams data in real time.
It handles both the initial sync and live replication, continuously monitoring for any changes and updating the destination database accordingly.
Using dsync, you can migrate your data from MongoDB to FerretDB.

In this post, you'll learn to use dsync to migrate the data from MongoDB to a running FerretDB instance.

## Prerequisites

Make sure to have the following ready before you start:

- [Running FerretDB instance](https://docs.ferretdb.io/installation/ferretdb/).
- Running MongoDB instance (local or remote).

## Download dsync

Ensure to download the latest release of dsync from the [GitHub Releases page](https://github.com/adiom-data/dsync/releases/latest).
For Mac users, you may need to configure a security exception to execute the binary by [following these steps](https://support.apple.com/en-ca/guide/mac-help/mh40616/mac).

Alternatively, you can build dsync from the source code.

```sh
git clone https://github.com/adiom-data/dsync.git
cd dsync
go build
```

## Run dsync to migrate into FerretDB

To migrate data from your local MongoDB instance to FerretDB, you need to specify the source and destination connection strings.

Let's say:

- MongoDB is running locally at `mongodb://localhost:27018/`
- FerretDB is running at `mongodb://<username>:<password>@localhost:27017`

Set the following environment variables for the source and destination connection strings and run dsync to migrate all data:

```sh
export MDB_SRC='mongodb://localhost:27018/'
export FERRETDB_DEST='mongodb://<username>:<password>@localhost:27017/'

./dsync --progress --logfile dsync.log "$MDB_SRC" "$FERRETDB_DEST"
```

Replace `<username>` and `<password>` with your FerretDB credentials.
If you're running FerretDB without authentication enabled, you can omit them.

When you run dsync, it opens a live-change monitoring session in your terminal to track the progress of the migration.

```text
Dsync Progress Report : ChangeStream
Time Elapsed: 00:12:50        1/1 Namespaces synced
Processing change stream events
```

The session will remain open for as long as dsync is running.
So even if new data is added to the source MongoDB instance, dsync will keep track of it and replicate it to FerretDB.

Confirm that the data has been migrated successfully by connecting to your FerretDB instance and checking the data.

## Run your workloads in open source with FerretDB

With dsync, start migrating your data from MongoDB to FerretDB – without any downtime, or data loss.

FerretDB lets you run your MongoDB workloads in open source, without fear of vendor lock-in or restrictive licenses like SSPL.
Your data is yours, and you can run it wherever you want.
With PostgreSQL with DocumentDB extension as the backend, FerretDB is designed to be a drop-in replacement for MongoDB, so you can keep using your existing tools and libraries without any changes.

Have any questions about the migration process?
[Contact us on any of our community channels](https://docs.ferretdb.io/#community) – we're happy to help.
