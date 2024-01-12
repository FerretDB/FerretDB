---
sidebar_position: 4
slug: /configuration/oplog-support/
---

# OpLog support

FerretDB currently has a basic implementation of the OpLog (operations log).

The OpLog is a special capped collection which stores all operations that modify your data.
A capped collection is a fixed-sized collection that overwrites its entries when it reaches its maximum size.
Naturally, OpLog is a capped collection so as to ensure that data does not grow unbounded.

:::note
At the moment, only basic OpLog tailing is supported.
Replication is not supported yet.
:::

Oplog support is critical for the Meteor framework to build real-time applications.
Such applications require notifications on real-time events and can use the OpLog to build a simple pub/sub system.

## Enabling OpLog functionality

FerretDB will not create the oplog automatically; you must do so manually.

To enable OpLog functionality, manually create a capped collection named `oplog.rs` in the `local` database.

```js
// use local
db.createCollection('oplog.rs', { capped: true, size: 536870912 })
```

You may also need to set the replica set name using [`--repl-set-name` flag / `FERRETDB_REPL_SET_NAME` environment variable](flags.md#general).

:::tip
**`--repl-set-name` flag / `FERRETDB_REPL_SET_NAME`** environment variable allow clients and drivers to perform an initial replication handshake.
We do not perform any replication but clients and drivers will assume that the replication protocol is being used.
The purpose of this flag is to allow access to the OpLOg.
:::

```sh
docker run -e FERRETDB_REPL_SET_NAME=rs0 ...
```

To query the OpLog:

```js
db.oplog.rs.find()
```

To query OpLog for all the operations in a particular namespace (`test.foo`), run:

```js
db.oplog.rs.find({ ns: 'test.foo' })
```

If something does not work correctly or you have any question on the OpLog functionality, [please inform us here](https://github.com/FerretDB/FerretDB/issues/new?assignees=ferretdb-bot&labels=code%2Fbug%2Cnot+ready&projects=&template=bug.yml).
