---
sidebar_position: 1
---

# Docker

These steps describe a quick local setup.
They are not suitable for most production use-cases because they keep all data inside containers.

1. Store the following in the `docker-compose.yml` file:

   ```yaml
   version: "3"

   services:
     postgres:
       image: postgres
       container_name: postgres
       ports:
         - 5432:5432
       environment:
         - POSTGRES_USER=ferret
         - POSTGRES_DB=ferretdb
         - POSTGRES_HOST_AUTH_METHOD=trust

     ferretdb:
       image: ghcr.io/ferretdb/ferretdb:latest
       container_name: ferretdb
       restart: on-failure
       ports:
         - 27017:27017
       environment:
         - FERRETDB_POSTGRESQL_URL=postgres://ferret@postgres:5432/ferretdb

   networks:
     default:
       name: ferretdb
   ```

   `postgres` container runs PostgreSQL that would store data.
   `ferretdb` runs FerretDB.

2. Start services with `docker-compose up -d`.

3. If you have `mongosh` installed, just run it to connect to FerretDB.
   If not, run the following command to run `mongosh` inside the temporary MongoDB container, attaching to the same Docker network:

   ```sh
   docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo mongodb://ferretdb/
   ```

You can also install with FerretDB with the `.deb` and `.rpm` packages
provided for each [release](https://github.com/FerretDB/FerretDB/releases).
