---
sidebar_position: 3
---

# Migrating from FerretDB v1.x

If you are running an older version of FerretDB (v1.24 or earlier versions),
we recommend upgrading to the latest version.

## Major differences

These are the major differences between FerretDB v1.x and v2.x:

1. Unlike v1.x that provides options for PostgreSQL and SQLite as backend,
   FerretDB v2.x **requires a PostgreSQL with DocumentDB extension** as the backend,
   which provides much better compatibility and performance.
   Find the installation guide for PostgreSQL with DocumentDB extension
   [here](../installation/documentdb/docker.md).
2. **`PLAIN` authentication is no longer supported** in FerretDB v2.x.
   FerretDB v2.x uses `SCRAM-SHA-256` for authentication.
   So if your connection string in v1.x looks like
   `mongodb://username:password@localhost:27017/ferretdb?authMechanism=PLAIN`,
   use `mongodb://username:password@localhost:27017/` with 2.x.
   Find out more about FerretDB authentication [here](../security/authentication.md).
3. OpLog is not supported yet.

## Migrating from FerretDB v1.x to v2.x

The migration process follows the same steps as MongoDB to FerretDB.
Follow [this guide to export and import your data](migrating-from-mongodb.md).
