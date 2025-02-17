---
sidebar_position: 2
---

# DEB package

FerretDB uses PostgreSQL with [DocumentDB extension](https://github.com/microsoft/documentdb) as a database engine.

Ensure you have [PostgreSQL](https://www.postgresql.org/download/linux/debian/) installed.
You may need to install additional dependencies required by the DocumentDB extension.

To install the `.deb` packages for [DocumentDB extension](https://github.com/microsoft/documentdb)
on your Debian, Ubuntu, and other `.deb`-based systems, you can use `dpkg` tool.

Download the latest DocumentDB `.deb` package from [our release pages](https://github.com/FerretDB/documentdb/releases/latest),
rename it to `documentdb.deb`, then run the following command in your terminal:

```sh
sudo dpkg -i documentdb.deb
```
