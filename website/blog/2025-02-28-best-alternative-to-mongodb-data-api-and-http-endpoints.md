---
slug: best-alternative-to-mongodb-data-api-and-http-endpoints
title: 'The Best Alternative for MongoDB Data API and HTTP Endpoints'
authors: [alex]
description: >
  MongoDB's deprecation of Atlas Data API left many developers without a suitable replacement. FerretDB v2 now provides a compatible alternative for the deprecated Atlas Data API.
image: /img/blog/mongodb-data-api.jpg
tags: [open source, sspl, product, document databases, community]
---

![The Best Alternative for MongoDB Data API and HTTP Endpoints](/img/blog/mongodb-data-api.jpg)

In September 2024, MongoDB announced the deprecation of Atlas Data API – among other features – leaving many developers affected, and without a suitable replacement.
We are happy to announce that [FerretDB v2](https://blog.ferretdb.io/ferretdb-releases-v2-faster-more-compatible-mongodb-alternative) now provides a compatible replacement for it.

<!--truncate-->

The Data API deprecation affected many developments, especially as it came without alternative migration options.
This is a common theme for many proprietary (or license-restrictive) projects, and not just MongoDB – [read more on the uncertainties of proprietary solutions here](https://blog.ferretdb.io/why-open-source-important-proprietary-uncertainties).
Open source solutions like FerretDB offer control, community support, and long-term stability, so you're never left stranded.

The good thing is: with the release of FerretDB v2, you can successfully replace MongoDB Data API and perform database operations on FerretDB using a direct HTTP-based method to access and interact with their databases.

In this post, we'll walk you through how to use FerretDB's Data API, demonstrating how you can find, insert, update, and delete documents – all without needing a MongoDB server.

## Why Data API matters

Some programming environments don't have a native MongoDB driver or simply can't use one – whether due to platform limitations, security constraints, or architectural decisions.
Without a driver, you'd normally have to wrap your logic inside another application or depend on a separate service just to send or retrieve data.
That's extra work and dependencies.

A Data API changes that – suddenly, your database is accessible through simple HTTP requests.
A basic `curl` command can fetch or modify data.
So instead of needing a full-fledged backend just to store something in a database, a workflow can send REST requests to a Data API.

For developers already familiar with REST, Data API is essential.
There's no new learning curve – you can interact with your database via HTTP requests.
It follows the defined [Data API OpenAPI documentation here](https://github.com/FerretDB/FerretDB/blob/main/internal/dataapi/api/openapi.json).

With FerretDB's Data API stepping in as an alternative to MongoDB's deprecated service, we are ensuring developers can interact with their data without issues, no matter the stack they're working with.

## FerretDB Data API

FerretDB is an open-source MongoDB-compatible database that runs on PostgreSQL.
With MongoDB's Data API deprecated, we saw the need for a new solution that offers similar functionality while remaining open and accessible.
In our case, the FerretDB Data API is built directly into FerretDB – it's not a standalone service but an integrated part of the database process.

FerretDB's Data API provides:

- Seamless HTTP-based database access – no need for drivers; interact via simple REST requests.
- MongoDB-compatible queries – use familiar JSON-based queries and commands.
- PostgreSQL-backed storage – leverage the reliability of PostgreSQL with a MongoDB-like API.

To access Data API on FerretDB, set the environment variable or flag (`FERRETDB_LISTEN_DATA_API_ADDR`/`--listen-data-api-addr`) to your desired port when starting FerretDB.

## Setting up FerretDB Data API

### Get your FerretDB instance running

You can set it up via Docker compose.
Here's an example:

```yaml
services:
  postgres:
    image: ghcr.io/ferretdb/postgres-documentdb:16
    restart: on-failure
    environment:
      - POSTGRES_USER=username
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=postgres
    volumes:
      - ./data:/var/lib/postgresql/data

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb:2
    restart: on-failure
    ports:
      - 27017:27017
      - 8080:8080
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://username:password@postgres:5432/postgres
      - FERRETDB_LISTEN_DATA_API_ADDR=:8080
networks:
  default:
    name: ferretdb
```

Note that the FerretDB v2 release requires a PostgreSQL with DocumentDB extension as the backend.
Once your FerretDB instance is running, the Data API endpoint will be available at `http://localhost:8080`.

### Using FerretDB's Data API for CRUD operations

Since you've enabled FerretDB's Data API (`FERRETDB_LISTEN_DATA_API_ADDR=:8080`), let's find, insert, update, and delete a single document using `curl`.

#### Insert a document

Insert a document into a collection called `users` in a `test` database.

```sh
curl -X POST http://127.0.0.1:8080/action/insertOne \
  -H "Content-Type: application/json" \
  -u username:password \
  -d '{
        "database": "test",
        "collection": "users",
        "document": {
          "name": "Andrew",
          "email": "andrew@example.com",
          "age": 25
        }
      }'
```

Response:

```json
{ "n": 1.0 }
```

#### Find a document

Next, let's retrieve the document we just inserted.

```sh
curl -X POST http://127.0.0.1:8080/action/find \
  -H "Content-Type: application/json" \
  -u username:password \
  -d '{
        "database": "test",
        "collection": "users",
        "filter": { "name": "Andrew" }
      }'
```

Response:

```json
{
  "documents": [
    {
      "_id": { "$oid": "67a2e9b70a5e9467a00918b0" },
      "name": "Andrew",
      "email": "andrew@example.com",
      "age": 25
    }
  ]
}
```

You can also try this via Postman or any other OpenAPI client.
This is how it looks in Postman:

![FerretDB Data API find operation](/img/blog/data-api-find.png)

#### Update a document

Let's update Andrew's email:

```sh
curl -X POST http://127.0.0.1:8080/action/updateOne \
  -H "Content-Type: application/json" \
  -u username:password \
  -d '{
        "database": "test",
        "collection": "users",
        "filter": { "name": "Andrew" },
        "update": { "$set": { "email": "andrew.new@example.com" } }
      }'
```

Response:

```json
{ "matchedCount": 1, "modifiedCount": 1 }
```

Let's also try this on Postman:

![FerretDB Data AP update operation](/img/blog/data-api-updateone.png)/Users/alexanderfashakin/projects/ferret/FerretDB/website/static/img/blog/data-api-deleteone.png

#### Delete a document

And to complete the CRUD operations, let's delete Andrew's document:

```sh
curl -X POST http://127.0.0.1:8080/action/deleteOne \
  -H "Content-Type: application/json" \
  -u username:password \
  -d '{
        "database": "test",
        "collection": "users",
        "filter": { "name": "Andrew" }
      }'
```

Response:

```json
{ "deletedCount": 1 }
```

And in Postman:

![FerretDB Data API delete operation](/img/blog/data-api-deleteone.png)

## Easily replace MongoDB Data API with FerretDB

With just a few months left before MongoDB's Data API and HTTPS endpoints reach end-of-life, it's crucial to migrate to a suitable alternative.
FerretDB's Data API provides a seamless replacement, allowing developers to interact with their databases via HTTP.
You can easily swap out your MongoDB instance for FerretDB and continue working with your data without any interruptions.
Besides, FerretDB is an open source solution – you can host it on your infrastructure, ensuring control, community support, and long-term stability.

Ready to migrate?
[Check out our migration guide on how you can get started](https://docs.ferretdb.io/migration/) and start using the Data API today!

And if you have any questions, bugs, or features that are unsupported for your use case, [please feel free to reach out to us](https://docs.ferretdb.io/#community)!
