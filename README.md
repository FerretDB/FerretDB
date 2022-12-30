# FerretDB

[![Go Reference](https://pkg.go.dev/badge/github.com/FerretDB/FerretDB/ferretdb.svg)](https://pkg.go.dev/github.com/FerretDB/FerretDB/ferretdb)
[![Go](https://github.com/FerretDB/FerretDB/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/go.yml)
[![Integration](https://github.com/FerretDB/FerretDB/actions/workflows/integration.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/integration.yml)
[![Docker](https://github.com/FerretDB/FerretDB/actions/workflows/docker.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/docker.yml)
[![codecov](https://codecov.io/gh/FerretDB/FerretDB/branch/main/graph/badge.svg?token=JZ56XFT3DM)](https://codecov.io/gh/FerretDB/FerretDB)

FerretDB (previously MangoDB) was founded to become the de-facto open-source substitute to MongoDB.
FerretDB is an open-source proxy, converting the MongoDB 6.0+ wire protocol queries to SQL -
using PostgreSQL as a database engine.

## Why do we need FerretDB?

MongoDB was originally an eye-opening technology for many of us developers,
empowering us to build applications faster than using relational databases.
In its early days, its ease-to-use and well-documented drivers made MongoDB one of the simplest database solutions available.
However, as time passed, MongoDB abandoned its open-source roots;
changing the license to SSPL - making it unusable for many open source and early-stage commercial projects.

Most MongoDB users do not require any advanced features offered by MongoDB;
however, they need an easy-to-use open-source database solution.
Recognizing this, FerretDB is here to fill that gap.

## Scope and current state

FerretDB will be compatible with MongoDB drivers and will strive to serve as a drop-in replacement for MongoDB 6.0+.

Currently, the project is in its early stages and welcomes all contributors.
See our [public roadmap](https://github.com/orgs/FerretDB/projects/2/views/1),
a list of [known differences with MongoDB](https://docs.ferretdb.io/diff/),
and [contributing guidelines](CONTRIBUTING.md).

## Quickstart

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

2. Start services with `docker compose up -d`.

3. If you have `mongosh` installed, just run it to connect to FerretDB.
   If not, run the following command to run `mongosh` inside the temporary MongoDB container, attaching to the same Docker network:

   ```sh
   docker run --rm -it --network=ferretdb --entrypoint=mongosh mongo mongodb://ferretdb/
   ```

You can also install with FerretDB with the `.deb` and `.rpm` packages
provided for each [release](https://github.com/FerretDB/FerretDB/releases).

## Building and packaging

We strongly advise users not to build FerretDB themselves.
Instead, use Docker images or .deb and .rpm packages provided by us.
If you want to package FerretDB for your operating system or distribution,
the recommended way is to use the `build-release` task;
see our [instructions for contributors](CONTRIBUTING.md) for more details.
FerretDB could also be built as any other Go program,
but a few generated files and build tags could affect it.
See [there](https://pkg.go.dev/github.com/FerretDB/FerretDB/build/version) for more details.

## Documentation

* [Documentation for users](https://docs.ferretdb.io/).
* [Documentation for Go developers about embeddable FerretDB](https://pkg.go.dev/github.com/FerretDB/FerretDB/ferretdb).

## Community

* Website and blog: [https://ferretdb.io](https://ferretdb.io/).
* Twitter: [@ferret_db](https://twitter.com/ferret_db).
* Mastodon: [@ferretdb@techhub.social](https://techhub.social/@ferretdb).
* [Slack chat](https://join.slack.com/t/ferretdb/shared_invite/zt-zqe9hj8g-ZcMG3~5Cs5u9uuOPnZB8~A) for quick questions.
* [GitHub Discussions](https://github.com/FerretDB/FerretDB/discussions) for longer topics.
* [GitHub Issues](https://github.com/FerretDB/FerretDB/issues) for bugs and missing features.
* [Open Office Hours meeting](https://calendar.google.com/event?action=TEMPLATE&tmeid=NjNkdTkyN3VoNW5zdHRiaHZybXFtb2l1OWtfMjAyMTEyMTNUMTgwMDAwWiBjX24zN3RxdW9yZWlsOWIwMm0wNzQwMDA3MjQ0QGc&tmsrc=c_n37tquoreil9b02m0740007244%40group.calendar.google.com&scp=ALL)
  every Monday at 18:00 UTC at [Google Meet](https://meet.google.com/mcb-arhw-qbq).

If you want to contact FerretDB Inc., please use [this form](https://www.ferretdb.io/contact/).
