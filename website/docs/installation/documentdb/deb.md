---
sidebar_position: 2
---

# DEB package

FerretDB uses PostgreSQL with [DocumentDB extension](https://github.com/microsoft/documentdb) as a database engine.

We provide different DocumentDB `.deb` packages for various deployments on our [release page](https://github.com/FerretDB/documentdb/releases/).

- For most use cases, we recommend using the production package (e.g., `documentdb.deb`).
- For debugging purposes, use the development package (contains either `-dev` or `-dbgsym` suffix e.g., `documentdb-dev.deb`/`documentdb-dbgsym.deb`).
  It includes features that significantly slow down performance and is not recommended for production use.

Download the appropriate DocumentDB `.deb` package from our release page.
Then, you can use `dpkg` tool to install it.

You need to install PostgreSQL and additional dependencies required by the DocumentDB extension.

After installing the package, you need to create the extension in your database.

Ensure to update the `postgresql.conf` file with the following settings so it can load the extension on startup:

```conf
shared_preload_libraries = 'pg_cron,pg_documentdb_core,pg_documentdb'
cron.database_name       = 'postgres'
```

You will need to restart the PostgreSQL service to apply the changes.

Then create the extension by running the folllowing inside the PostgreSQL shell:

```sql
CREATE EXTENSION documentdb CASCADE;
```
