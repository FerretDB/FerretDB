---
sidebar_position: 3
---

# RPM package

To install the `.rpm` packages for FerretDB on your RHEL, CentOS, and other `.rpm`-based systems,
you can use `rpm` tool.

Download the latest FerretDB `.rpm` package from [our release pages](https://github.com/FerretDB/FerretDB/releases/latest),
then run the following command in your terminal:

```sh
sudo rpm -i ferretdb.rpm
```

You can check that FerretDB was installed by running

```sh
ferretdb --version
```

FerretDB does not automatically installs PostgreSQL or other backends.
To install PostgreSQL, run the following commands:

```sh
sudo yum install -y postgresql
```

Currently, our `.rpm` package does not provide a SystemD unit for starting FerretDB automatically.
You have to do it manually by running `ferretdb` binary with the [correct flags](../configuration/flags.md).

Find out more about:

* [getting logs](../configuration/logging.md#binary-executable-logs).
