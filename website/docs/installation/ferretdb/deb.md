---
sidebar_position: 2
---

# DEB package

To install the `.deb` packages for FerretDB on your Debian, Ubuntu, and other `.deb`-based systems,
you can use `dpkg` tool.

Download the latest FerretDB `.deb` package from [our release pages](https://github.com/FerretDB/FerretDB/releases/latest),
rename it to `ferretdb.deb`,
then run the following command in your terminal:

```sh
sudo dpkg -i ferretdb.deb
```

You can check that FerretDB was installed by running

```sh
ferretdb --version
```

The `.deb` package ships with the systemd unit for starting FerretDB automatically.
For more information about its configuration, please take a look at [systemd configuration guide](systemd.md).

FerretDB does not automatically install PostgreSQL and DocumentDB extension,
see DocumentDB extension DEB package [installation guide](../documentdb/deb.md).

:::tip
The latest FerretDB package works best with [DocumentDB v0.102.0-ferretdb-2.0.0-rc.2](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.0.0-rc.2).
:::

Find out more about:

- [getting logs](../../configuration/observability.md#logging).
