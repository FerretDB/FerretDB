---
sidebar_position: 2
---

# DEB package

FerretDB uses PostgreSQL with [DocumentDB extension](https://github.com/microsoft/documentdb) as a database engine.

Download the DocumentDB `.deb` package from [our release page](https://github.com/FerretDB/documentdb/releases/),
you can use `dpkg` tool to install it.

For example, to install the `documentdb.deb` package, run:

```bash
sudo dpkg -i documentdb.deb
```

:::tip
For more information on the best FerretDB version to use, see the [DocumentDB release notes](https://github.com/FerretDB/documentdb/releases/).
:::

You need to install PostgreSQL and additional dependencies required by the DocumentDB extension.

After installing the package, you need to create the extension in your database.

Ensure to update the `postgresql.conf` file with the following settings:

```conf
shared_preload_libraries = 'pg_cron,pg_documentdb_core,pg_documentdb'
cron.database_name       = 'postgres'
```

Then create the extension by running the folllowing inside the PostgreSQL shell:

```sql
CREATE EXTENSION documentdb CASCADE;
 ```

