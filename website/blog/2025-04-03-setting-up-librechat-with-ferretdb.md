---
slug: setting-up-librechat-with-ferretdb
title: '# Setting up LibreChat with FerretDB'
authors: [alex]
description: >
  This guide will walk you through setting up LibreChat with FerretDB.
image: /img/blog/ferretdb-openziti.jpg
tags: [tutorial, community, open source]
---

![Setting up LibreChat with FerretDB](/img/blog/ferretdb-openziti.jpg)

LibreChat is an free, open-source application that provides a user-friendly, and customizable interface for interacting with various AI providers.
It allows users to connect with providers like OpenAI, Azure, Anthropic, Google, and more.

For users of LibreChat interested in remaining in the open-source ecosystem, FerretDB is a great alternative to MongoDB.
With FerretDB, you can run your AI applications without worrying about vendor lock-in or proprietary concerns.
It uses Postgres with DocumentDB extension as the backend while enabling you to use familiar syntax and commands.

This guide will walk you through setting up LibreChat with FerretDB.

<!--truncate-->

## Setting up FerretDB with LibreChat

You can set LibreChat to use FerretDB in any of the following two ways:

### Connect to an existing FerretDB instance

This setup assumes you already have a running FerretDB instance.
If you don't, set it up using any of the methods mentioned in the [FerretDB installation guide](https://docs.ferretdb.io/installation/ferretdb/).

If you are setting up LibreChat with Docker, replace the MongoDB connection string by specifying the FerretDB URI in the `docker-compose.override.yml` file.
This will allow LibreChat to connect directly to it instead of MongoDB.
Here's an example of how you can set up the connection in your `docker-compose.override.yml` file:

```yaml
services:
  api:
    environment:
      - MONGO_URI=mongodb://<username>:<password>@<host>:<porrt>/LibreChat // your FerretDB instance
```

If you're using `npm` to set up LibreChat, update the `MONGO_URI` in your `.env` file to point to your FerretDB instance instead of MongoDB.

For example:

```text
MONGO_URI=mongodb://<username>:<password>@<host>:<porrt>/LibreChat // your FerretDB instance
```

### Add FerretDB to LibreChat `docker-compose.override.yml` file

If you want to set up FerretDB together with LibreChat in the same Docker Compose environment, you can do so by adding FerretDB and PostgreSQL with DocumentDB extension to your `docker-compose.override.yml` file.

The LibreChat `MONGO_URI` environment variable will point to the FerretDB instance.

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

Conclusion

You can improve your FerretDB and LibreChat setup by checking out the following resources:

- [Setup authentication for FerretDB](https://docs.ferretdb.io/security/auth/)
- [Troubleshooting FerretDB](https://docs.ferretdb.io/troubleshooting/)
- [LibreChat Docker setup](https://www.librechat.ai/docs/local/docker)

If you have any questions on FerretDB, feel free to reach out to us on any of [our community channels](https://docs.ferretdb.io/#community).
