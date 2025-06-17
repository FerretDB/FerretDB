---
sidebar_position: 5
---

# Systemd unit

:::note
This feature is still experimental.
If you encounter any problem, please [join our community](/#community) to report it.
:::

With both DEB and RPM package we ship the systemd unit, to start FerretDB automatically.
If FerretDB is not installed yet, please refer to the [`.deb`](deb.md) or [`.rpm`](rpm.md) installation pages.

The unit file connects to a local postgresql server by default.
You can overwrite the [configuration](../../configuration/flags.md)
by using the `systemctl edit ferretdb` command.
This'll open a text editor that'll allow you to override the systemd options of choice.
For example, you can reference an external file for your postgres URL and password:

```systemd
### Editing /etc/systemd/system/ferretdb.service.d/override.conf
### Anything between here and the comment below will become the new contents of the file

[Service]
Environment="FERRETDB_POSTGRESQL_URL_FILE=/etc/my_protected_ferretdb_url"

### Lines below this comment will be discarded
...
```
