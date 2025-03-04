---
sidebar_position: 3
---

# Migrating from FerretDB v1.x

If you are running an older version of FerretDB (v1.24 or earlier versions), we recommend upgrading to the latest version.

These are the major differences between FerretDB v1.x and v2.x:

1. Unlike v1.x that provides options for PostgreSQL and SQLite as backend, FerretDB v2.x **requires a PostgreSQL with DocumentDB extension** as the backend.
   Find the [installation guide for PostgreSQL with DocumentDB extension here](../installation/documentdb/).

2. **`PLAIN` authentication is no longer supported** in FerretDB v2.x.
   FerrretDB v2.x uses `SCRAM-SHA-256` for authentication.
   So if your connection string in v1.x looks like this (`mongodb://username:password@localhost:27017/ferretdb?authMechanism=PLAIN`), in v2.x, (`mongodb://username:password@localhost:27017/`).
   [Find out more on FerretDB authentication here](../security/authentication.md).

## Backup and restore data from FerretDB v1.x to v2.x

The migration process follows the same steps as MongoDB to FerretDB.
Follow [this guide to dump and restore your data](migrating-from-mongodb.md).
