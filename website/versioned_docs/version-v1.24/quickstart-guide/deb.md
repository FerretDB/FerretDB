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

FerretDB does not automatically install PostgreSQL or other backends.
To install PostgreSQL, run the following commands:

```sh
sudo apt update
sudo apt install -y postgresql
```

The `.deb` package ships with the systemd unit for starting FerretDB automatically.
For more information about its configuration, please take a look at [systemd configuration guide](systemd.md).

Find out more about:

- [getting logs](../configuration/observability.md#logging).
