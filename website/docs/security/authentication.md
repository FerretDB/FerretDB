---
sidebar_position: 1
slug: /security/authentication/ # referenced in error messages
description: Learn to use authentication mechanisms
---

# Authentication

FerretDB provides authentication via the backend's authentication mechanisms and the experimental authentication mode.

For the backend's authentication mechanism, the default username and password can be specified in FerretDB's connection string,
but the client could use a different user by providing a username and password in MongoDB URI.

For example, if the server was started with `postgres://user1:pass1@postgres:5432/ferretdb`,
anonymous clients will be authenticated as user1,
but clients that use `mongodb://user2:pass2@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN` MongoDB URI will be authenticated as user2.
Since usernames and passwords are transferred in plain text,
the use of [TLS](../security/tls-connections.md) is highly recommended.

The FerretDB experimental authentication mode allows you to create user credentials for authenticated connections.
See [experimental authentication mode](#experimental-authentication-mode) for more.

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
    restart: on-failure
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

## Experimental authentication mode

FerretDB provides a new experimental authentication mode that supports the `SCRAM-SHA-1` and `SCRAM-SHA-256` authentication mechanisms.
This mode enables FerretDB to manage users by itself and also provide support for more user management commands.

You can enable this mode by setting the `FERRETDB_TEST_ENABLE_NEW_AUTH` flag or `--test-enable-new-auth` to `true`.
See [flags](../configuration/flags.md) for more information.

For example:

```sh
ferretdb --test-enable-new-auth=true
```

With this new authentication mode, you can create user credentials for authenticated connections using the `createUser` command and also access other user management commands such as `dropAllUsersFromDatabase`, `dropUser`, `updateUser`, and `usersInfo`.

This mode also enables you to set up initial authentication credentials for your instance.

### Initial authentication setup

You can secure your connections right from scratch by setting up an initial authentication credential using the following [dedicated flags or environment variables](../configuration/flags.md)

- `--setup-username`/`FERRETDB_SETUP_USERNAME`: Specifies the username to be created.
- `--setup-password`/`FERRETDB_SETUP_PASSWORD`: Specifies the password for the user (can be empty).
- `--setup-database`/`FERRETDB_SETUP_DATABASE`: Specifies the initial database that will be created.

:::note
`--test-enable-new-auth`/`FERRETDB_TEST_ENABLE_NEW_AUTH` must be set to `true` to enable the authentication setup.
:::

Once the flags/environment variables are passed, FerretDB will create the specified user with the given password and the given database.

#### Initial authentication setup with Postgres backend

A typical setup for a local Postgres database with an initial user setup would look like this:

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
      - FERRETDB_POSTGRESQL_URL=postgres://username:password@postgres:5432/ferretdb
      - FERRETDB_TEST_ENABLE_NEW_AUTH=true
      - FERRETDB_SETUP_USERNAME=user
      - FERRETDB_SETUP_PASSWORD=pass
      - FERRETDB_SETUP_DATABASE=ferretdb

networks:
  default:
    name: ferretdb
```

You can then start the services with `docker compose up -d`.
Once the services are up and running, you can connect to the FerretDB instance using the authencation credentials created during the setup.

```sh
mongosh "mongodb://user:pass@ferretdb/ferretdb"
```

#### Initial authentication setup with SQLite backend

You can configure your instance to be created with an initial user for authentication.

A typical setup would look like this:

```yaml
services:
  ferretdb:
    image: ghcr.io/ferretdb/ferretdb
    restart: on-failure
    ports:
      - 27017:27017
    environment:
      - FERRETDB_HANDLER=sqlite
      - FERRETDB_TEST_ENABLE_NEW_AUTH=true
      - FERRETDB_SETUP_USERNAME=user
      - FERRETDB_SETUP_PASSWORD=pass
      - FERRETDB_SETUP_DATABASE=ferretdb
    volumes:
      - ./state:/state

networks:
  default:
    name: ferretdb
```

You can start the services and connect to the instance as shown in the Postgres example above.
