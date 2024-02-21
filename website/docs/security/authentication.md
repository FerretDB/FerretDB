---
sidebar_position: 1
slug: /security/authentication/ # referenced in error messages
description: Learn to use authentication mechanisms
---

# Authentication

FerretDB supports `PLAIN`, `SCRAM-SHA-256` and `SCRAM-SHA-1` authentication mechanisms.
A user with supported mechanism is created by `createUser()` command and stored in the `admin.system` database `users` collection.

FerretDB uses passed username and password to authenticate against stored credentials.
For example, if a client connects as `mongodb://user1:pass1@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN`,
`user1` is authenticated against its credential in `admin.system` database `users` collection.

Before the first user is created in FerretDB, the credential passed in the MongoDB connection string is used to connect directly to the PostgreSQL backend via passthrough.
For example, when `admin.system` database `users` collection is empty and
a client connects as `mongodb://pguser1:pgpass1@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN`,
it uses `pguser1` to connect to the postgreSQL backend. The `PLAIN` mechanism is used for this case.
Please note this exception no longer applies once the first user is created.

When usernames and passwords are transferred in plain text,
the use of [TLS](../security/tls-connections.md) is highly recommended.

## PostgreSQL backend

In the following examples, username and password are specified in FerretDB's connection string `pguser1:pgpass1`.
Ensure `pguser1` is a PostgreSQL user with necessary
[privileges](https://www.postgresql.org/docs/current/sql-grant.html).
See more about [creating PostgreSQL user](https://www.postgresql.org/docs/current/sql-createuser.html)
and [PostgreSQL authentication methods](https://www.postgresql.org/docs/current/auth-methods.html).

### Using `ferretdb` package

Start `ferretdb` by specifying `--postgresql-url` with username and password.
All authenticated clients use `pguser1` user to query PostgreSQL backend.

```sh
ferretdb --postgresql-url=postgres://pguser1:pgpass1@localhost:5432/ferretdb
```

A client connects as user `user1` that is authenticated using credentials stored in `admin.system` database `users` collection.

```sh
mongosh 'mongodb://user1:pass1@127.0.0.1/ferretdb?authMechanism=PLAIN'
```

### Using Docker

For Docker, specify `FERRETDB_POSTGRESQL_URL` with username and password.

```yaml
services:
  postgres:
    image: postgres
    environment:
      - POSTGRES_USER=username
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=ferretdb
    volumes:
      - ./data:/var/lib/postgresql/data

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb
    restart: on-failure
    ports:
      - 27017:27017
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://pguser1:pgpass1@postgres:5432/ferretdb

networks:
  default:
    name: ferretdb
```

To start `ferretdb`, use docker compose.

```sh
docker compose up
```

A client connects as user `user1` that is authenticated using credentials stored in `admin.system` database `users` collection.

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh \
  mongo 'mongodb://user1:pass1@ferretdb/ferretdb?authMechanism=PLAIN'
```

## Authentication Handshake

:::note
Some drivers may still use the legacy `hello` command to complete a handshake.
:::

If you encounter any issues while authenticating with FerretDB, try setting the Stable API version to V1 on the client as this may prevent legacy commands from being used.
Please refer to your specific driver documentation on how to set this field.

If this does not resolve your issue please file a bug report [here](https://github.com/FerretDB/FerretDB/issues/new?assignees=ferretdb-bot&labels=code%2Fbug%2Cnot+ready&projects=&template=bug.yml).
