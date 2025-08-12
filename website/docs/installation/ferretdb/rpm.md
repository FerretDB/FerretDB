---
sidebar_position: 4
---

# RPM package

To install the `.rpm` packages for FerretDB on your RHEL, CentOS, and other `.rpm`-based systems,
you can use `rpm`, `dnf`, or any other appropriate package manager for your system.

We provide different `.rpm` packages for various deployments on [our release page](https://github.com/FerretDB/FerretDB/releases/).

- For most use cases, we recommend using the production package (e.g., `ferretdb.rpm`).
- For debugging purposes, use the development package (contains a `-dev` suffix e.g., `ferretdb-dev.rpm`).
  It includes features that significantly slow down performance and is not recommended for production use.

## Installation

Download the appropriate FerretDB `.rpm` package from our release page,
rename it to `ferretdb.rpm`,
then run the following command in your terminal:

```sh
dnf install -y ferretdb.rpm
```

You can check that FerretDB was installed by running:

```sh
ferretdb --version
```

:::note
FerretDB does not automatically install PostgreSQL with the DocumentDB extension,
see DocumentDB extension RPM package [installation guide](../documentdb/rpm.md) for more details.
:::

The `.rpm` package ships with the systemd unit for starting FerretDB automatically.
For more information about its configuration, please take a look at [systemd configuration guide](systemd.md).

Find out more about:

- [getting logs](../../configuration/observability.md#logging).

## Updating to a new version

Before updating to a new FerretDB version, make sure to update to the matching DocumentDB package first.
Following the [DocumentDB update guide](../documentdb/rpm.md#updating-to-a-new-version) is critical for a successful update.

Download the new `.rpm` package from the release page.
Then, install the new package using `dnf`:

```sh
dnf install -y /path/to/<new-ferretdb-package>.rpm
```

Then verify that the new version is installed by running:

```sh
ferretdb --version
```
