---
sidebar_position: 1
slug: /security/authentication/ # referenced in error messages
description: Learn to use authentication mechanisms
---

# Authentication

The authentication is used to verify the identity of the client connecting to FerretDB using username and password.
It blocks the access if the client does not have the valid credentials.

FerretDB provides authentication by relying on the [backend's authentication](#postgresql-backend-authentication) mechanisms by default.
However, there is an [experimental authentication](#experimental-authentication) feature that allows managing and authenticating users within FerretDB.

Having user management within FerretDB allows creations and deletions of users via normal MongoDB commands and hides the necessity to have the knowledge of the specific backend to manage users.

## PostgreSQL backend authentication

A client connects to the PostgreSQL backend by the provided username and password from the FerretDB's connection string or the MongoDB URI.
Only the `PLAIN` mechanism is supported for the backend authentication.

When starting FerretDB, the default username and password can be specified in FerretDB's connection string,
but the client could use a different user by providing a username and password in MongoDB URI.
For example, if the server was started with `postgres://user1:pass1@postgres:5432/ferretdb`,
anonymous clients will be authenticated as user1,
but clients that use `mongodb://user2:pass2@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN` MongoDB URI will be authenticated as user2.
Since usernames and passwords are transferred in plain text,
the use of [TLS](../security/tls-connections.md) is highly recommended.

### PostgreSQL backend with default username and password

In following examples, default username and password are specified in FerretDB's connection string `user1:pass1`.
Ensure `user1` is a PostgreSQL user with necessary
[privileges](https://www.postgresql.org/docs/current/sql-grant.html).
See more about [creating PostgreSQL user](https://www.postgresql.org/docs/current/sql-createuser.html)
and [PostgreSQL authentication methods](https://www.postgresql.org/docs/current/auth-methods.html).

#### Using `ferretdb` package

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

#### Using Docker

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

## Experimental authentication

Upon enabling the experimental authentication by the flag `--test-enable-new-auth` or the environment variable `TEST_ENABLE_NEW_AUTH`,
FerretDB enables user management and authentication against the stored credentials in the `admin` database `system.users` collection.
The users are managed using the `usersInfo`, `createUser`, `updateUser`, `dropUser` and `dropAllUsersFromDatabase` commands.
The users may be created and authenticated using `SCRAM-SHA-256`, `SCRAM-SHA-1` and `PLAIN` mechanisms.

For example, if a client uses `mongodb://user:pass@ferretdb:27018/ferretdb?tls=true&authMechanism=SCRAM-SHA-256`,
`user` is authenticated against its credential stored in the `admin.system.users` collection using `SCRAM-SHA-256` authentication mechanism.

If the `PLAIN` mechanism is used, the use of [TLS](../security/tls-connections.md) is highly recommended since the password is transferred in plain text.

### Example using experimental authentication

Start `ferretdb` by specifying `--postgresql-url`.

```sh
ferretdb --postgresql-url=postgres://pguser:pgpass@localhost:5432/ferretdb
```

A client connects as the user `user`, and authenticates using credentials stored in the `admin.system.users` collection.

```sh
mongosh 'mongodb://user:pass@127.0.0.1/ferretdb?authMechanism=SCRAM-SHA-256'
```

## Authentication Handshake

:::note
Some drivers may still use the legacy `hello` command to complete a handshake.
:::

If you encounter any issues while authenticating with FerretDB, try setting the Stable API version to V1 on the client as this may prevent legacy commands from being used.
Please refer to your specific driver documentation on how to set this field.

If this does not resolve your issue please file a bug report [here](https://github.com/FerretDB/FerretDB/issues/new?assignees=ferretdb-bot&labels=code%2Fbug%2Cnot+ready&projects=&template=bug.yml).
