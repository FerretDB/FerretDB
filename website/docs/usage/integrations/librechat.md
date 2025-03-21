---
sidebar_position: 2
slug: /integrations/librechat
description: Learn how to set up LibreChat with FerretDB.
---

# Setting up LibreChat with FerretDB

FerretDB is a truly open-source document database alternative to MongoDB that uses Postgres with DocumentDB as the backend.
It enables you to use familiar syntax and commands – without vendor lock-in or proprietary concerns.

You can set up your FerretDB instance with LibreChat in two ways:

## Add FerretDB and PostgreSQL via Docker Compose Override

If you don't already have a running FerretDB instance, you can replace the default `mongo` service in the `docker-compose.override.yml` file by adding FerretDB and PostgreSQL with DocumentDB extension.

Here's an example of how you can set up FerretDB and PostgreSQL with DocumentDB extension in your `docker-compose.override.yml` file:

```yaml
services:
  api:
    environment:
      - MONGO_URI=mongodb://ferretdb:27017/LibreChat

  postgres:
    image: ghcr.io/ferretdb/postgres-documentdb:17-0.102.0-ferretdb-2.0.0
    platform: linux/amd64
    restart: on-failure
    environment:
      - POSTGRES_USER=username
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=postgres
    volumes:
      - ./ferret-data:/var/lib/postgresql/ferret-data

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb:2.0.0
    restart: on-failure
    ports:
      - 27017:27017
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://username:password@postgres:5432/postgres
      - FERRETDB_AUTH=false
```

## Connect to an existing FerretDB instance

With your FerretDB instance running, you can specify your connection URI in the `docker-compose.override.yml` file.
This will allow LibreChat to connect directly to it.

```yaml
services:
  api:
    environment:
      - MONGO_URI=mongodb://ferretdb:27017/LibreChat
```

Learn more about FerretDB setup here:

- [FerretDB installation guide](https://docs.ferretdb.io/installation/ferretdb/)
- [FerretDB authentication](https://docs.ferretdb.io/security/auth/)

:::note
If you're using `npm` to set up LibreChat, update the `MONGO_URI` in your `.env` file to point to your FerretDB instance instead of MongoDB.
:::
