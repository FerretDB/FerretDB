---
sidebar_position: 1
slug: /security/ # referenced in README.md
description: authentication
---

# Authentication

FerretDB does not store authentication information (usernames and passwords) itself but uses the backend's authentication mechanisms.
The default username and password can be specified in FerretDB's connection string,
but the client could use a different user by providing a username and password in MongoDB URI.
For example, if the server was started with `postgres://user1:pass1@postgres:5432/ferretdb`,
anonymous clients will be authenticated as user1,
but clients that use `mongodb://user2:pass2@ferretdb:27018/ferretdb?tls=true&authMechanism=PLAIN` MongoDB URI will be authenticated as user2.
Since usernames and passwords are transferred in plain text,
the use of TLS is highly recommended.

## PostgreSQL backend with default username and password

PostgreSQL server may start with default username and password.

In the example, default username and password are specified in FerretDB's connection string `user1:pass1`.
See more about [creating PostgreSQL user](https://www.postgresql.org/docs/current/sql-createuser.html)
and [PostgreSQL authentication methods](https://www.postgresql.org/docs/current/auth-methods.html).

### Using `ferretdb` package

Start `ferretdb` by specify `--postgresql-url` with default username and password.

```sh
ferretdb --postgresql-url=postgres://user1:pass1@postgres:5432/ferretdb
```

An anonymous client is authenticated with default `user1` from `--postgresql-url`.

```sh
mongosh `mongodb://127.0.0.1/ferretdb`
```

A client that specify username and password in MongoDB URI as below is authenticated as `user2`.
See how to [create user](https://www.postgresql.org/docs/current/sql-createuser.html).

```sh
mongosh `mongodb://user2:pass2@127.0.0.1/ferretdb?authMechanism=PLAIN`
```

### Using docker

For using docker, specify `FERRETDB_POSTGRESQL_URL` with default username and password.

```yml
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
docker compose up -d
```

An anonymous client is authenticated with `user1` from `FERRETDB_POSTGRESQL_URL`.
Use following command to run `mongosh` inside the temporary MongoDB container,
attached to the same Docker network.

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo `mongodb://ferretdb/ferretdb`
```

Clients that specify username and password in MongoDB URI such as below is authenticated as `user2`.

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo \
`mongodb://user2:pass2@ferretdb/ferretdb?authMechanism=PLAIN`
```
