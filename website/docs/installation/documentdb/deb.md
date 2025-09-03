---
sidebar_position: 3
---

# DEB package

FerretDB uses PostgreSQL with [DocumentDB extension](https://github.com/documentdb/documentdb) as a database engine.

We provide different DocumentDB `.deb` packages for various deployments on our [release page](https://github.com/FerretDB/documentdb/releases/).

- For most use cases, we recommend using the production package (e.g., `documentdb.deb`).
- For debugging purposes, use the development package (contains either `-dev` or `-dbgsym` suffix e.g., `documentdb-dev.deb`/`documentdb-dbgsym.deb`).
  It includes features that significantly slow down performance and is not recommended for production use.

## Installation

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

<!-- Keep in sync with https://github.com/FerretDB/documentdb/blob/ferretdb/ferretdb_packaging/10-preload.sh -->

```text
shared_preload_libraries                      = 'pg_cron,pg_documentdb_core,pg_documentdb'
cron.database_name                            = 'postgres'

documentdb.enableCompact                      = true

documentdb.enableLetAndCollationForQueryMatch = true
documentdb.enableNowSystemVariable            = true
documentdb.enableSortbyIdPushDownToPrimaryKey = true

documentdb.enableSchemaValidation             = true
documentdb.enableBypassDocumentValidation     = true

documentdb.enableUserCrud                     = true
documentdb.maxUserLimit                       = 100
```

Ensure to restart PostgreSQL for the changes to take effect.

Then create the extension by running the following SQL command within the `postgres` database:

```sql
CREATE EXTENSION documentdb CASCADE;
```

You can now go ahead and set up FerretDB by following [this installation guide](../ferretdb/deb.md).

## Updating to a new version

Before [updating to a new FerretDB release](../ferretdb/docker.md#updating-to-a-new-version), it is critical to install the matching DocumentDB package first.

The following steps are critical to ensuring a successful update.

Download the new `.deb` package that matches the FerretDB release you are updating to from the release page.

Then install the new package using `dpkg`:

```sh
sudo dpkg -i /path/to/<new-documentdb-package.deb>
```

Replace `/path/to/<new-documentdb-release.deb>` with the actual path and filename of the downloaded `.deb` package.

After installing the new package, you need to update the DocumentDB extension in your PostgreSQL database.
To do this, run the following command from within the `postgres` database:

```sh
sudo -u postgres psql -d postgres -c 'ALTER EXTENSION documentdb UPDATE;'
```

Next, verify or update your `postgresql.conf` to include the correct extension libraries on startup (same as listed in the Installation section above).
Restart PostgreSQL to apply changes.

Once the DocumentDB update is complete, proceed with the FerretDB update steps.
