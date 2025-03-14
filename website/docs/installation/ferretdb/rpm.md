---
sidebar_position: 3
---

# RPM package

We provide different `.rpm` packages for various deployments:

- **a development package**: for debugging problems, and includes features that make it significantly slower.
  For this reason, it is not recommended for production use.
- **a production package**: for all other cases.

:::note
Development packages include a `-dev` suffix (e.g. `ferretdb-dev.rpm`) and are not recommended for production use.
On the other hand, production packages do not include the `-dev` suffix (e.g. `ferretdb.rpm`).
:::

To install the `.rpm` packages for FerretDB on your RHEL, CentOS, and other `.rpm`-based systems,
you can use `rpm` tool.

Download the latest FerretDB `.rpm` package from [our release pages](https://github.com/FerretDB/FerretDB/releases/latest),
rename it to `ferretdb.rpm`,
then run the following command in your terminal:

```sh
sudo rpm -i ferretdb.rpm
```

You can check that FerretDB was installed by running

```sh
ferretdb --version
```

FerretDB does not automatically install PostgreSQL.
To install PostgreSQL, run the following commands:

```sh
sudo yum install -y postgresql
```

The `.rpm` package ships with the systemd unit for starting FerretDB automatically.
For more information about its configuration, please take a look at [systemd configuration guide](systemd.md).

Find out more about:

- [getting logs](../../configuration/observability.md#logging).
