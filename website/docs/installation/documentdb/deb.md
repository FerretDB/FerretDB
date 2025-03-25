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

Before installing the DocumentDB extension, you need to install PostgreSQL and additional dependencies required by the DocumentDB extension.

With the dependencies installed, you can install the DocumentDB extension using `dpkg`.
For example, to install the `documentdb.deb` package, run the following command:

```sh
sudo dpkg -i /path/to/documentdb.deb
```

Ensure to replace `/path/to/documentdb.deb` with the actual path and filename of the downloaded `.deb` package.

After that, update the `postgresql.conf` of your PostgreSQL instance with the following settings so it can load the extension on startup:

```text
shared_preload_libraries = 'pg_cron,pg_documentdb_core,pg_documentdb'
cron.database_name       = 'postgres'
```

Note that if the database instance is already running before updating the config file, you may need to restart the PostgreSQL service to apply the changes.

Then create the extension by running the following SQL command within the PostgreSQL instance:

```sql
CREATE EXTENSION documentdb CASCADE;
```

You can now go ahead and set up FerretDB by following [this installation guide](../ferretdb/deb.md).
