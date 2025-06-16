---
sidebar_position: 3
---

# DEB package

We provide different `.deb` packages for various deployments on [our release page](https://github.com/FerretDB/FerretDB/releases/).

- For most use cases, we recommend using the production package (e.g., `ferretdb.deb`).
- For debugging purposes, use the development package (contains a `-dev` suffix e.g., `ferretdb-dev.deb`).
  It includes features that significantly slow down performance and is not recommended for production use.

## Installation

Download the appropriate FerretDB `.deb` package from our release page,
rename it to `ferretdb.deb`.

To install the `ferretdb.deb` package on your Debian, Ubuntu, and other `.deb`-based systems,
you can use `dpkg` tool, as shown below:

```sh
sudo dpkg -i ferretdb.deb
```

You can check that FerretDB was installed by running:

```sh
ferretdb --version
```

The `.deb` package ships with the systemd unit for starting FerretDB automatically.
For more information about its configuration, please take a look at [systemd configuration guide](systemd.md).

FerretDB does not automatically install PostgreSQL and DocumentDB extension,
see DocumentDB extension DEB package [installation guide](../documentdb/deb.md).

:::tip
Ensure to [enable telemetry](../../telemetry.md) to receive notifications on the latest versions.
For more information on the best DocumentDB version to use, see the [corresponding release notes for the FerretDB package](https://github.com/FerretDB/FerretDB/releases/).
:::

Find out more about:

- [getting logs](../../configuration/observability.md#logging).

## Updating to a new version

Before updating to a new FerretDB version, make sure to update to the matching DocumentDB package first.
Following the [DocumentDB update guide](../documentdb/docker.md#updating-to-a-new-version) is critical for a successful update.

Download the new `.deb` package from the release page.
Then, install the new package using `dpkg`:

```sh
sudo dpkg -i /path/to/<new-ferretdb-package>.deb
```

Then verify that the new version is installed by running:

```sh
ferretdb --version
```
