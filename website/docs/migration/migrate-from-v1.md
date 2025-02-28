---
sidebar_position: 3
---

# Migrating from FerretDB v1.x

If you are running an older version of FerretDB (v1.24, or any other version), we recommend upgrading to the latest version.
This guide will help you migrate your data from FerretDB v1.x to v2.x.

:::note
FerretDB v2.x requires a PostgreSQL with DocumentDB extension as the backend.
:::

## Major differences between FerretDB v1.x and v2.x

### PostgreSQL with DocumentDB extension backend

Unlike v1.x that provides options for PostgreSQL and SQLite as backend, FerretDB v2.x requires a PostgreSQL with DocumentDB extension as the backend.
This represents a significant change in the architecture of FerretDB; running FerretDB v2.x with just PostgreSQL will not work.

With PostgreSQL with DocumentDB extension, you get better performance, features, and compatibility for your MongoDB workloads.
You can find the installation guide for PostgreSQL with DocumentDB extension [here](../installation/postgresql-documentdb.md).

### Authentication

`PLAIN` authentication is no longer supported in FerretDB v2.x.
FerrretDB v2.x uses `SCRAM-SHA-256` for authentication.
As such, you will need to update your connection string to use `SCRAM-SHA-256` for authentication.

If your connection string in v1.x looks like this (`mongodb://username:password@localhost:27017/ferretdb?authMechanism=PLAIN`), you will need to update it to this (`mongodb://username:password@localhost:27017/`) in v2.x.
Otherwise, you will encounter an authentication error.

For more information on authentication in FerretDB, please refer to the [authentication documentation](../security/authentication.md).

## Migrate your data from FerretDB v1.x to v2.x

To migrate your data, you will need to have the following in place:

- `mongodump` utility
- `mongorestore` utility
- Connection string to a running instance of FerretDB v1.x
- Connection string to a running instance of FerretDB v2.x with PostgreSQL with DocumentDB extension as the backend

### Backup your data

Before migrating your data from FerretDB v1.x to v2.x, ensure you have a backup of your data.
You can create a backup of your data using the `mongodump` utility.

```sh
mongodump --uri="mongodb://username:password@localhost:27017/ferretdb?authMechanism=PLAIN"
```

This command will create a dump of all your data, consisting of BSON files of all the collections in your FerretDB instance.

### Set up a running instance of FerretDB v2.x

To set up a running instance of FerretDB v2.x, follow the installation guide [here](../installation/install-ferretdb.md).
Ensure that it's running with a PostgreSQL with DocumentDB extension as the backend.

### Migrating data

To migrate your data from to v2.x, use the `mongorestore` utility to restore the data from the backup you created earlier.

```sh
mongorestore --uri="mongodb://username:password@localhost:27017/"
```

If you are restoring all the data from the same directory as the backup, you can run the command without any additional parameters.
Otherwise, you can specify the database/collection you want to restore from the `dump` folder:

```sh
mongorestore --uri="mongodb://username:password@localhost:27017/" dump/ferretdb/collection.bson
```

This should migrate all your data successfully to FerretDB v2.x.

If you encounter any issues during the migration process, please send us a message along with the error message and FerretDB logs to any of our community channels.
