---
sidebar_position: 4
---

# Systemd Unit

With both DEB and RPM package we ship the systemd unit, to start FerretDB automatically.
If FerretDB is not installed yet, please refer to the [`.deb`](https://docs.ferretdb.io/quickstart-guide/deb/) or [`.rpm`](https://docs.ferretdb.io/quickstart-guide/rpm/) installation pages.

The unit file provides some basic environment variables as an example. 
They should be overwritten with the proper configuration, with the [correct flags](../configuration/flags.md).

```
[Service]
ExecStart=/usr/bin/ferretdb
Restart=on-failure

# Configure the FerretDB service with `systemctl edit ferretdb`.
# For more configuration options check https://docs.ferretdb.io/configuration/flags/

Environment="FERRETDB_POSTGRESQL_URL=postgres://username:password@127.0.0.1:5432/ferretdb"
Environment="FERRETDB_LISTEN_ADDR=:27017"
Environment="FERRETDB_DEBUG_ADDR=:8088"
```

You can modify them by using `systemctl edit ferretdb` command.
This'll open a text editor that'll allow you to
override the systemd options of choice.
For example, if we want to use SQLite backend instead of PostgreSQL, we could write something like:

```
### Editing /etc/systemd/system/ferretdb.service.d/override.conf
### Anything between here and the comment below will become the new contents of the file

[Service]
Environment="FERRETDB_SQLITE_URL=file:/var/lib/ferretdb/data/"
Environment="FERRETDB_HANDLER=sqlite"

Environment="FERRETDB_POSTGRESQL_URL="

### Lines below this comment will be discarded
...
```

