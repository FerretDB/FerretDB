---
sidebar_position: 1
slug: /security/authentication/ # referenced in error messages
description: Learn to use authentication mechanisms
---

# Authentication

FerretDB does not store authentication information (usernames and passwords) itself but uses the backend's authentication mechanisms.
The default username and password can be specified in FerretDB's connection string,
but the client could use a different user by providing a username and password in MongoDB URI.
For example, if the server was started with `postgres://user1:pass1@postgres:5432/ferretdb`,
anonymous clients will be authenticated as user1,
but clients that use `mongodb://user2:pass2@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN` MongoDB URI will be authenticated as user2.
Since usernames and passwords are transferred in plain text,
the use of [TLS](../security/tls-connections.md) is highly recommended.

## PostgreSQL backend with default username and password

In following examples, default username and password are specified in FerretDB's connection string `user1:pass1`.
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

## Experimental authentication

FerretDB supports experimental `SCRAM-SHA-256` and `SCRAM-SHA-1` authentication mechanisms.
It is enabled by the flag `--test-enable-new-auth` or the environment variable `TestEnableNewAuth`.
A user is created using the `createUser` command and stored in the `admin` database `system.users` collection.

FerretDB uses the provided username and password to authenticate against the stored credentials.
For example, if a client connects as `mongodb://user1:pass1@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN`,
`user1` is authenticated against its credential stored in the `admin.system.users` collection.

Before the first user is created in FerretDB, the credentials provided in the MongoDB connection string are used to connect directly to the PostgreSQL backend via passthrough.
For instance, when the `admin.system.users` collection is empty and
a client connects as `mongodb://pguser1:pgpass1@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN`,
it uses `pguser1` to connect to the postgreSQL backend.
The `PLAIN` mechanism is used for this case.
Please note this exception no longer applies once the first user is created.

When usernames and passwords are transferred in plain text or on an unsecured network,
the use of [TLS](../security/tls-connections.md) is highly recommended.

#### Example using experimental authentication

Start `ferretdb` by specifying `--postgresql-url` with username and password.
All authenticated clients use `pguser1` user to query PostgreSQL backend.

```sh
ferretdb --postgresql-url=postgres://pguser1:pgpass1@localhost:5432/ferretdb
```

A client connects as the user `user1`, and authenticated using credentials stored in the `admin.system.users` collection.

```sh
mongosh 'mongodb://user1:pass1@127.0.0.1/ferretdb?authMechanism=PLAIN'
```
