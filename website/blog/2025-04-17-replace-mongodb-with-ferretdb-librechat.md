---
slug: replace-mongodb-with-ferretdb-librechat
title: 'Replacing MongoDB with FerretDB in LibreChat'
authors: [alex]
description: >
  A guide to replacing MongoDB with FerretDB for a fully open-source LibreChat setup.
image: /img/blog/ferretdb-librechat.jpg
tags: [tutorial, community, open source]
---

![Replace MongoDB with FerretDB on LibreChat](/img/blog/ferretdb-librechat.jpg)

[LibreChat](https://www.librechat.ai/) is a free, open-source application that provides a user-friendly and customizable interface for interacting with various AI providers.
It allows users to connect with providers like [OpenAI](https://openai.com/), [Azure](https://azure.microsoft.com/), [Anthropic](https://www.anthropic.com/), and more.

LibreChat users who want to stay within the open-source ecosystem will find FerretDB a fitting replacement for MongoDB.

If you're looking to avoid proprietary databases or vendor lock-in, FerretDB is a drop-in replacement for MongoDB that runs on top of PostgreSQL.
It uses PostgreSQL with DocumentDB extension as the backend while enabling you to use familiar syntax and commands.

This guide shows how to run LibreChat with FerretDB as the database – either by connecting to an existing FerretDB instance or running everything in Docker.

<!--truncate-->

## How to use FerretDB with LibreChat

LibreChat can use FerretDB as its MongoDB-compatible database in two ways:

### Option 1: Connect to an existing FerretDB instance

If you already have FerretDB running, simply connect LibreChat to it using the MONGO_URI environment variable.

For Docker-based setups, update `docker-compose.override.yml` file:

```yaml
services:
  api:
    environment:
      - MONGO_URI=mongodb://<username>:<password>@<host>:<port>/LibreChat
```

For local development with `npm`, update the `.env` file to point to your FerretDB instance instead of MongoDB.

For example:

```text
MONGO_URI=mongodb://<username>:<password>@<host>:<porrt>/LibreChat
```

:::note

If you're new to FerretDB, you can find [installation instructions here](https://docs.ferretdb.io/installation/ferretdb/).

:::

### Option 2: Add FerretDB and PostgreSQL via Docker Compose

To run everything together, add FerretDB and PostgreSQL (with the DocumentDB extension) to your `docker-compose.override.yml` file.

Here's an example:

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

Once set up, run `docker-compose up` to start the entire stack.

## Further resources

Learn more about FerretDB and LibreChat by checking out the following resources:

- [Setup authentication for FerretDB](https://docs.ferretdb.io/security/auth/)
- [Troubleshooting FerretDB](https://docs.ferretdb.io/troubleshooting/)
- [LibreChat Docker setup](https://www.librechat.ai/docs/local/docker)

Need help?
Feel free to reach out to us on any of [our community channels](https://docs.ferretdb.io/#community).
