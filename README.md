# FerretDB

[![Go Reference](https://pkg.go.dev/badge/github.com/FerretDB/FerretDB/ferretdb.svg)](https://pkg.go.dev/github.com/FerretDB/FerretDB/ferretdb)

[![Go](https://github.com/FerretDB/FerretDB/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/go.yml)
[![Integration](https://github.com/FerretDB/FerretDB/actions/workflows/integration.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/integration.yml)
[![codecov](https://codecov.io/gh/FerretDB/FerretDB/branch/main/graph/badge.svg?token=JZ56XFT3DM)](https://codecov.io/gh/FerretDB/FerretDB)

[![Security](https://github.com/FerretDB/FerretDB/actions/workflows/security.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/security.yml)
[![Packages](https://github.com/FerretDB/FerretDB/actions/workflows/packages.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/packages.yml)
[![Docs](https://github.com/FerretDB/FerretDB/actions/workflows/docs.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/docs.yml)

FerretDB was founded to become the de-facto open-source substitute to MongoDB.
FerretDB is an open-source proxy, converting the MongoDB 6.0+ wire protocol queries to SQL -
using PostgreSQL as a database engine.

## Why do we need FerretDB?

MongoDB was originally an eye-opening technology for many of us developers,
empowering us to build applications faster than using relational databases.
In its early days, its ease-to-use and well-documented drivers made MongoDB one of the simplest database solutions available.
However, as time passed, MongoDB abandoned its open-source roots;
changing the license to SSPL - making it unusable for many open source and early-stage commercial projects.

Most MongoDB users do not require any advanced features offered by MongoDB;
however, they need an easy-to-use open-source document database solution.
Recognizing this, FerretDB is here to fill that gap.

## Scope and current state

FerretDB is compatible with MongoDB drivers and popular MongoDB tools.
It functions as a drop-in replacement for MongoDB 6.0+ in many cases.
Features are constantly being added to further increase compatibility and performance.

We welcome all contributors.
See our [public roadmap](https://github.com/orgs/FerretDB/projects/2/views/1),
a list of [known differences with MongoDB](https://docs.ferretdb.io/diff/),
and [contributing guidelines](CONTRIBUTING.md).

## Quickstart

```sh
docker run -d --rm --name ferretdb -p 27017:27017 ghcr.io/ferretdb/all-in-one
```

This command will start a container with FerretDB, PostgreSQL, and MongoDB Shell for testing and experiments.
However, it is unsuitable for production use cases because it keeps all data inside and loses it on shutdown.
See our [Docker quickstart guide](https://docs.ferretdb.io/quickstart-guide/docker/) for instructions
that don't have those problems.

With that container running, you can:

* Connect to it with any MongoDB client application using MongoDB URI `mongodb://127.0.0.1:27017/`.
* Connect to it using MongoDB Shell by just running `mongosh`.
  If you don't have it installed locally, you can run `docker exec -it ferretdb mongosh`.
* Connect to PostgreSQL running inside the container by running `docker exec -it ferretdb psql -U username ferretdb`.
  FerretDB uses PostgreSQL schemas for MongoDB databases.
  So, if you created some collections in the `test` database using any MongoDB client,
  you can switch to it by running `SET search_path = 'test';` query
  and see a list of PostgreSQL tables by running `\d` `psql` command.

You can stop the container with `docker stop ferretdb`.

We also provide binaries and packages for various Linux distributions.
See [our documentation](https://docs.ferretdb.io/quickstart-guide/) for more details.

## Building and packaging

We strongly advise users not to build FerretDB themselves.
Instead, use binaries, Docker images, or `.deb`/`.rpm` packages provided by us.

If you want to package FerretDB for your operating system or distribution,
the recommended way to build the binary is to use the `build-release` task;
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
