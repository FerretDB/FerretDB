---
sidebar_position: 1
slug: /quickstart-guide/docker/ # referenced in README.md
description: How to set up FerretDB using Docker
---

# Docker

You provide three Docker images for various deployments:
"all-in-one" for quick testing and experiments,
a development image for debugging problems,
and a production image for all other cases.

All-in-one image is documented in the
[README.md file in the repository](https://github.com/FerretDB/FerretDB#quickstart).
The rest are covered below.

## Production image

Our [production image](https://ghcr.io/ferretdb/ferretdb) `ghcr.io/ferretdb/ferretdb`
is recommended for most deployments.
It does not include PostgreSQL or other backends, so you must run them separately.
You can do that with Docker Compose, Kubernetes, or other means.

### Setup with Docker Compose

The following steps describe a quick local setup:

1. Store the following in the `docker-compose.yml` file:

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
         - FERRETDB_POSTGRESQL_URL=postgres://postgres:5432/ferretdb

   networks:
     default:
       name: ferretdb
   ```

   `postgres` container runs PostgreSQL that would store data in the `./data` directory on the host.
   `ferretdb` runs FerretDB.

2. Start services with `docker compose up -d`.
3. If you have `mongosh` installed, just run it to connect to FerretDB.
   It will use credentials passed in `mongosh` flags or MongoDB URI to authenticate to the PostgreSQL database.
   You'll also need to set `authMechanism` to `PLAIN`.
   The example URI would look like:

   ```text
   mongodb://username:password@127.0.0.1/ferretdb?authMechanism=PLAIN
   ```

   See [Authentication](../security.md#authentication) for more details.

   If you don't have `mongosh`, run the following command to run it inside the temporary MongoDB container,
   attaching to the same Docker network:

   ```sh
   docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo \
     "mongodb://username:password@ferretdb/ferretdb?authMechanism=PLAIN"
   ```

You can improve that setup by:

* [securing connections with TLS](../security.md#securing-connections-with-tls);
* adding backups.

Find out more about:

* [Getting logs](../configuration/logging.md#docker-logs)

## Development image

The development image contains the [debug build](https://pkg.go.dev/github.com/FerretDB/FerretDB/build/version#hdr-Debug_builds)
of FerretDB with test coverage instrumentation, race detector,
and other changes that make it more suitable for debugging problems.
It can be used exactly the same way as the production image, as described above.
