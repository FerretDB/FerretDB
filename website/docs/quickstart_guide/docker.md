---
sidebar_position: 1
---

# Docker

These steps describe a quick local setup.
They are not suitable for most production use-cases because they keep all data inside containers.

<!-- markdownlint-disable MD029 -->

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
      - POSTGRES_DB=ferretdb
      - POSTGRES_HOST_AUTH_METHOD=trust

  postgres_setup:
    image: postgres:14
    container_name: postgres_setup
    restart: on-failure
    entrypoint: ["sh", "-c", "psql -h postgres -U user -d ferretdb -c 'CREATE SCHEMA IF NOT EXISTS test'"]

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb:latest
    container_name: ferretdb
    restart: on-failure
    ports:
      - 27017:27017
    command: ["-listen-addr=:27017", "-postgresql-url=postgres://user@postgres:5432/ferretdb"]

networks:
  default:
    name: ferretdb
```

* `postgres` container runs PostgreSQL 14 that would store data.
* `postgres_setup` container creates a PostgreSQL schema `test` that would act like a FerretDB database of the same name.
* `ferretdb` runs FerretDB.

2. Start services with `docker-compose up -d`.

3. If you have `mongosh` installed, just run it to connect to FerretDB database `test`.
If not, run the following command to run `mongosh` inside the temporary MongoDB container, attaching to the same Docker network:

```sh
docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo:5 mongodb://ferretdb/
```

<!-- markdownlint-enable MD029 -->

You can also install with FerretDB with the `.deb` and `.rpm` packages
provided for each [release](https://github.com/FerretDB/FerretDB/releases).
