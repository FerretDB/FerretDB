---
sidebar_position: 1
slug: /security/authentication/ # referenced in error messages
description: Learn to use authentication mechanisms
---

# Authentication

FerretDB supports `PLAIN`, `SCRAM-SHA-256` and `SCRAM-SHA-1` authentication mechanisms.
A user with supported mechanism is created by `createUser()` command and stored in `admin.system` database `users` collection.

FerretDB uses passed username and password to authenticate against stored credentials.
For example, if a client connects as `mongodb://user1:pass1@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN`,
`user1` is authenticated against its record in `admin.system` database `users` collection.

Before the first user is created, the credential passed in the connection string is used to authenticate directly to the postgreSQL backend.
For example, when `admin.system` database `users` collection is empty and
a client connects as `mongodb://dbuser1:dbpass1@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN`,
it uses `dbuser1` to authenticate directly to the postgreSQL backend.
Please note this exception no longer applies once the first user is created.

When usernames and passwords are transferred in plain text,
the use of [TLS](../security/tls-connections.md) is highly recommended.

## PostgreSQL backend with default username and password

In following examples, username and password are specified in FerretDB's connection string `user1:pass1`.
Ensure `user1` is a PostgreSQL user with necessary
[privileges](https://www.postgresql.org/docs/current/sql-grant.html).
See more about [creating PostgreSQL user](https://www.postgresql.org/docs/current/sql-createuser.html)
and [PostgreSQL authentication methods](https://www.postgresql.org/docs/current/auth-methods.html).

### Using `ferretdb` package

Start `ferretdb` by specifying `--postgresql-url` with default username and password.

```sh
ferretdb --postgresql-url=postgres://user1:pass1@localhost:5432/ferretdb
```

An anonymous client is authenticated with default `user1` from `--postgresql-url`.

```sh
mongosh 'mongodb://127.0.0.1/ferretdb'
```

A client that specify username and password in MongoDB URI as below is authenticated as `user2`.

```sh
mongosh 'mongodb://user2:pass2@127.0.0.1/ferretdb?authMechanism=PLAIN'
```

### Using Docker

For Docker, specify `FERRETDB_POSTGRESQL_URL` with default username and password.

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
      - FERRETDB_POSTGRESQL_URL=postgres://user1:pass1@postgres:5432/ferretdb

networks:
  default:
    name: ferretdb
```

To start `ferretdb`, use docker compose.

```sh
docker compose up
```

An anonymous client is authenticated with `user1` from `FERRETDB_POSTGRESQL_URL`.
Use following command to run `mongosh` inside the temporary MongoDB container,
attached to the same Docker network.

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh \
  mongo 'mongodb://ferretdb/ferretdb'
```

A client that specify username and password in MongoDB URI as below is authenticated as `user2`.

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh \
  mongo 'mongodb://user2:pass2@ferretdb/ferretdb?authMechanism=PLAIN'
```

## Authentication Handshake

:::note
Some drivers may still use the legacy `hello` command to complete a handshake.
:::

If you encounter any issues while authenticating with FerretDB, try setting the Stable API version to V1 on the client as this may prevent legacy commands from being used.
Please refer to your specific driver documentation on how to set this field.

If this does not resolve your issue please file a bug report [here](https://github.com/FerretDB/FerretDB/issues/new?assignees=ferretdb-bot&labels=code%2Fbug%2Cnot+ready&projects=&template=bug.yml).
