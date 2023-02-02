---
sidebar_position: 1
---

# Docker

These steps describe a quick local setup.
They are not suitable for most production use-cases because they keep all data
inside containers and don't [encrypt incoming connections](../security.md#securing-connections-with-tls).
For more configuration options check [Configuration flags and variables](../flags.md) page.

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
         - POSTGRES_USER=username
         - POSTGRES_PASSWORD=password
         - POSTGRES_DB=ferretdb

     ferretdb:
       image: ghcr.io/ferretdb/ferretdb:latest
       container_name: ferretdb
       restart: on-failure
       ports:
         - 27017:27017
       environment:
         - FERRETDB_POSTGRESQL_URL=postgres://postgres:5432/ferretdb

   networks:
     default:
       name: ferretdb
   ```

   `postgres` container runs PostgreSQL that would store data.
   `ferretdb` runs FerretDB.

2. Fetch the latest version of FerretDB with `docker compose pull`.
   Afterwards start services with `docker compose up -d`.

3. If you have `mongosh` installed, just run it to connect to FerretDB.
   It will use credentials passed in `mongosh` flags or MongoDB URI
   to authenticate to the PostgreSQL database.
   You'll also need to set `authMechanism` to `PLAIN`.
   The example URI would look like:

   ```text
   mongodb://username:password@127.0.0.1/ferretdb?authMechanism=PLAIN
   ```

   See [Security#Authentication](../security.md#authentication) for more details.

   If you don't have `mongosh`, run the following command to run it inside the temporary MongoDB container, attaching to the same Docker network:

   ```sh
   docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo "mongodb://username:password@ferretdb/ferretdb?authMechanism=PLAIN"
   ```

You can also install FerretDB with the `.deb` and `.rpm` packages
provided for each [release](https://github.com/FerretDB/FerretDB/releases).
