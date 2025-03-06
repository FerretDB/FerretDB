---
sidebar_position: 4
---

# Systemd Unit

:::note
This feature is still experimental.
If you encounter any problem, please [join our community](/#community) to report it.
:::

With both DEB and RPM package we ship the systemd unit, to start FerretDB automatically.
If FerretDB is not installed yet, please refer to the [`.deb`](deb.md) or [`.rpm`](rpm.md) installation pages.

The unit file provides some basic environment variables as an example.
They should be overwritten with the proper [configuration](../../configuration/flags.md).

```systemd
[Service]
ExecStart=/usr/bin/ferretdb
Restart=on-failure

# Configure the FerretDB service with `systemctl edit ferretdb`.
# For more configuration options check https://docs.ferretdb.io/configuration/flags/

Environment="FERRETDB_POSTGRESQL_URL=postgres://127.0.0.1:5432/postgres"
```

You can modify them by using `systemctl edit ferretdb` command.
This'll open a text editor that'll allow you to override the systemd options of choice.
For example, if we want to listen on the Unix domain socket, we could write something like:

```systemd
### Editing /etc/systemd/system/ferretdb.service.d/override.conf
### Anything between here and the comment below will become the new contents of the file

[Service]
Environment="FERRETDB_LISTEN_UNIX=/var/lib/ferretdb/ferretdb.sock"

### Lines below this comment will be discarded
...
```
