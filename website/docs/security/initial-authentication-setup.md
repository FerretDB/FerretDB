---
sidebar_position: 3
slug: /security/initial-authentication-setup/
description: Learn to set up authentication from scratch
---

# Initial authentication setup

You can secure your connections right from scratch by setting up an initial authentication credential using the following dedicated flags or environment variables.

- `--setup-username`/`FERRETDB_SETUP_USERNAME`: Specifies the username to be created.
- `--setup-password`/`FERRETDB_SETUP_PASSWORD`: Specifies the password for the user (can be empty).
- `--setup-database`/`FERRETDB_SETUP_DATABASE`: Specifies the initial database that will be created.

:::note
`--test-enable-new-auth`/`FERRETDB_TEST_ENABLE_NEW_AUTH` must be set to `true` to enable the authentication setup.
:::

Once the flags/environment variables are passed, FerretDB will create the specified user with the given password and the given database.

## Initial authentication setup with Postgres backend

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

## Initial authentication setup with SQLite backend

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
