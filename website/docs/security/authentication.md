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

Having user management within FerretDB allows creations and deletions of users via command and hides the necessity to have the knowledge of the specific backend to manage users.
However, FerretDB does not support authorization yet which means fine-grained access control is not yet possible.

## PostgreSQL backend authentication

For a postgreSQL backend, it uses provided username and password to connect to the postgreSQL backend directly.
The supported mechanism is `PLAIN` which passes the password in a plain text.

When starting FerretDB, the default username and password can be specified in FerretDB's connection string,
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

Upon enabling the flag `--test-enable-new-auth` or the environment variable `TEST_ENABLE_NEW_AUTH`,
FerretDB enables user management and authentication against the stored credentials in the `admin` database `system.users` collection.
The users are managed using the `usersInfo`, `createUser`, `updateUser`, `dropUser` and `dropAllUsersFromDatabase` commands.
The users may be created and authenticated using `SCRAM-SHA-256`, `SCRAM-SHA-1` and `PLAIN` mechanisms.
The `SCRAM` mechanisms are the algorithms used to store hashed passwords.
FerretDB also stores hashed password using `PLAIN` mechanism.

Please note that FerretDB does not support authorization yet, so created users currently have access to all databases.

To connect to FerretDB, if a client use `mongodb://user:pass@ferretdb:27018/ferretdb?tls=true&authMechanism=SCRAM-SHA-256`,
`user` is authenticated against its credential stored in the `admin.system.users` collection using `SCRAM-SHA-256` authentication mechanism.

Before the first user is created in FerretDB, the credentials provided in the MongoDB connection string are used to connect directly to the PostgreSQL backend via passthrough.
For instance, when the `admin.system.users` collection is empty and
a client connects as `mongodb://pguser:pgpass@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN`,
it uses `pguser` to connect to the postgreSQL backend.
When authenticating directly to the backend, only `PLAIN` mechanism is supported.
Please note this exception no longer applies once the first user is created.

When usernames and passwords are transferred in plain text or on an unsecured network,
the use of [TLS](../security/tls-connections.md) is highly recommended.

## Example using experimental authentication

Start `ferretdb` by specifying `--postgresql-url` with username and password.
All authenticated clients use `pguser` user to query PostgreSQL backend.

```sh
ferretdb --postgresql-url=postgres://pguser:pgpass@localhost:5432/ferretdb
```

A client connects as the user `user`, and authenticated using credentials stored in the `admin.system.users` collection.

```sh
mongosh 'mongodb://user:pass@127.0.0.1/ferretdb?authMechanism=SCRAM-SHA-256'
```
