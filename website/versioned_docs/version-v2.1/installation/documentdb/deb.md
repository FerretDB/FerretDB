---
sidebar_position: 3
---

# DEB package

FerretDB uses PostgreSQL with [DocumentDB extension](https://github.com/microsoft/documentdb) as a database engine.

We provide different DocumentDB `.deb` packages for various deployments on our [release page](https://github.com/FerretDB/documentdb/releases/).

- For most use cases, we recommend using the production package (e.g., `documentdb.deb`).
- For debugging purposes, use the development package (contains either `-dev` or `-dbgsym` suffix e.g., `documentdb-dev.deb`/`documentdb-dbgsym.deb`).
  It includes features that significantly slow down performance and is not recommended for production use.

Download the appropriate DocumentDB `.deb` package from our release page.

Before installing the DocumentDB extension, you need to install PostgreSQL and additional dependencies required by the DocumentDB extension.

With the dependencies installed, you can install the DocumentDB extension using `dpkg`.
For example, to install the `documentdb.deb` package, run the following command:

```sh
sudo dpkg -i /path/to/documentdb.deb
```

Ensure to replace `/path/to/documentdb.deb` with the actual path and filename of the downloaded `.deb` package.

Once installed, update your `postgresql.conf` to load the extension libraries on startup into the default `postgres` database.
Add the following lines to `postgresql.conf`:

```text
shared_preload_libraries = 'pg_cron,pg_documentdb_core,pg_documentdb'
cron.database_name       = 'postgres'
```

Ensure to restart PostgreSQL for the changes to take effect.

Then create the extension by running the following SQL command within the `postgres` database:

```sql
CREATE EXTENSION documentdb CASCADE;
```

You can now go ahead and set up FerretDB by following [this installation guide](../ferretdb/deb.md).
