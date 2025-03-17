---
sidebar_position: 2
---

# DEB package

:::note
We provide different `.deb` packages for various deployments.

- The development package (contains either `-dev` or `-dbgsym` suffix e.g., `documentdb-dev.deb`/`documentdb-dbgsym.deb`) is for debugging purposes.
  It includes features that significantly slow performance and is not recommended for production use.
- For other use cases, we recommend the production package (e.g., `documentdb.deb`).

:::

FerretDB uses PostgreSQL with [DocumentDB extension](https://github.com/microsoft/documentdb) as a database engine.

Download the DocumentDB `.deb` package from [our release page](https://github.com/FerretDB/documentdb/releases/),
you can use `dpkg` tool to install it.

:::tip
For more information on the best FerretDB version to use, see the [DocumentDB release notes](https://github.com/FerretDB/documentdb/releases/).
:::

You need to install PostgreSQL and additional dependencies required by the DocumentDB extension.
