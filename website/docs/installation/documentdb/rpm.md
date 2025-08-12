---
sidebar_position: 4
---

# RPM package

To install the `.rpm` packages for DocumentDB, you can use the `rpm` tool.

We provide different `.rpm` packages for various Red Hat Enterprise Linux (RHEL) and PostgreSQL major versions on [our release page](https://github.com/FerretDB/documentdb/releases/) (e.g. `rhel8-postgresql16-documentdb-<version>-1.el8.x86_64.rpm`)

## Installation

Before installing the DocumentDB extension, make sure to install PostgreSQL and all additional dependencies required by the DocumentDB extension.

With the dependencies installed, you can install the DocumentDB extension using `rpm`.

Download the appropriate DocumentDB `.rpm` package from the release page, then install it by running the following command in your terminal:

```sh
sudo rpm -i /path/to/documentdb.rpm
```

Ensure to replace `/path/to/documentdb.rpm` with the actual path and filename of the downloaded `.rpm` package.

You can check that DocumentDB was installed by running

```sh
documentdb --version
```

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

You can now go ahead and set up FerretDB by following [this installation guide](../ferretdb/rpm.md).

## Updating to a new version

Before [updating to a new FerretDB release](../ferretdb/rpm.md#updating-to-a-new-version), it is critical to install the matching DocumentDB `.rpm` package first.

The following steps are critical to ensuring a successful update.

Download the new `.rpm` package that matches the FerretDB release you are updating to from the release page.

Then install the new package using `rpm`:

```sh
sudo rpm -i /path/to/<new-documentdb-package.rpm>
```

Replace `/path/to/<new-documentdb-package.rpm>` with the actual path and filename of the downloaded `.rpm` package.

After installing the new package, update the DocumentDB extension in your PostgreSQL database.
To do this, run the following command from within the `postgres` database:

```sh
sudo -u postgres psql -d postgres -c 'ALTER EXTENSION documentdb UPDATE;'
```

Next, verify or update your `postgresql.conf` to include the correct extension libraries on startup (same as listed in the Installation section above).
Restart PostgreSQL to apply changes.

Once the DocumentDB update is complete, proceed with the [FerretDB update steps](../ferretdb/rpm.md#updating-to-a-new-version).
