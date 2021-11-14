# MangoDB

MangoDB is set out to become the de-facto open-source alternative to MongoDB.
MangoDB is an open-source proxy, which converts MongoDB wire protocol queries to SQL, and uses PostgreSQL as a database engine.


## Why do we need MangoDB?

MongoDB is a life-changing technology for many developers, empowering them to build applications faster than using relational databases.
Its easy-to-use and well-documented drivers make MongoDB one of the easiest to use database solutions available.
However, MongoDB abandoned its open-source roots, changing the license to SSPL - making it unusable for many open source and commercial projects.

Most MongoDB users are not in need of many of the advanced features offered by MongoDB; however, they are in need of an open-source database solution.
Recognizing this, MangoDB is here to fill the gap by providing an alternative.


## Scope

MangoDB will be compatible with MongoDB drivers and will work as a drop-in replacement for MongoDB in many cases.


## Current state

What you see here is a tech demo intended to show a proof of concept.
Over the next couple of months, we will be working on adding more.
See [this example](https://github.com/MangoDB-io/example) for a short demonstration.

MangoDB is in a very early stage and welcomes all contributors.
See [CONTRIBUTING.md](CONTRIBUTING.md).


## Quickstart

Those steps describe a quick local setup.
They are not suitable for production use.

1. Store the following in the `docker-compose.yml` file:

```yaml
version: "3"

services:
  postgres:
    image: postgres:14
    container_name: postgres
    ports:
      - 5432:5432
    environment:
      - POSTGRES_USER=user
      - POSTGRES_DB=mangodb
      - POSTGRES_HOST_AUTH_METHOD=trust

  postgres_setup:
    image: postgres:14
    container_name: postgres_setup
    restart: on-failure
    entrypoint: ["sh", "-c", "psql -h postgres -U user -d mangodb -c 'CREATE SCHEMA IF NOT EXISTS test'"]

  mangodb:
    image: ghcr.io/mangodb-io/mangodb:latest
    container_name: mangodb
    restart: on-failure
    ports:
      - 27017:27017
    command: ["--listen-addr=:27017", "--postgresql-url=postgres://user@postgres:5432/mangodb"]
```

* `postgres` container runs PostgreSQL 14 that would store data.
* `postgres_setup` container creates a PostgreSQL schema `test` that would act like a MangoDB database of the same name.
* `mangodb` runs MangoDB.

2. Start services with `docker-compose up -d`.

3. If you have `mongosh` installed, just run it to connect to MangoDB database `test`.
If not, run the following command to run `mongosh` inside the temporary MongoDB container, attaching to the same Docker network:
```
docker run --rm -it --network=mangodb_default --entrypoint=mongosh mongo:5 mongodb://mangodb/
```


## Contact us

Visit us at [www.mangodb.io](https://www.mangodb.io/), get in touch, and sign up for updates on the project.
