---
slug: ferretdb-releases-v1150
title: FerretDB releases v1.15.0!
authors: [alex]
description: >
  FerretDB is delighted to announce the release of v1.15.0 with support for `showRecordId`in `find` and JSON format for logging.
image: /img/blog/ferretdb-v1.15.0.jpg
tags: [release]
---

![FerretDB releases v.1.15.0](/img/blog/ferretdb-v1.15.0.jpg)

FerretDB is delighted to announce the release of v1.15.0 with support for `showRecordId`in `find` and JSON format for logging.

<!--truncate-->

With the new release, we've changed our artifacts naming scheme; our binaries and packages now include `linux` as part of their file names.
The purpose of this is to prepare for providing artifacts for other operating systems.

There are other enhancements and changes in this release, including enabling the use of existing PostgreSQL schema, and making it possible to use FeretDB without a state directory.

Let's check them out!

## New features and enhancements

FerretDB v1.15.0 comes with support for `showRecordId` in `find` command.
This will enable you to view each document's unique storage ID along with its data.
We've also included support for JSON format for logging.
You can find out more details on this in our [Observability docs](https://docs.ferretdb.io/configuration/observability/).

Along with these new additions, we now provide an option for disabling the `--debug-addr` flag.
You can now disable the debug handler if the value of that flag is set to an empty string or `-`.
See [here for more](https://docs.ferretdb.io/configuration/flags/#interfaces).

Similarly, we've enabled the use of FerretDB without a state directory.
Like the debug handler, you can also disable it by setting `--state-dir` to an empty string or `-`.

Some users may prefer to use default schemas like `public`, so with this release, we've enabled the use of existing PostgreSQL schema.
We've also made it possible to generate SQL queries with comments for `find` operations.

## Documentation

In this release, we've updated our pre-migration page to include extra details on how to specify the `--listen-addr` and `--proxy-addr` flags or set the `FERRETDB_LISTEN_ADDR` and `FERRETDB_PROXY_ADDR` environment variables.

In addition to this, we've now enabled interactive code playground in FerretDB via [Codapi](https://codapi.org/).
This will make it possible for users to run FerretDB commands directly from our blog or documentation.
Check out [this blog post to see how it works](https://blog.ferretdb.io/mongodb-crud-operations-with-ferretdb/).

## Other updates

Please see [our release notes](https://github.com/FerretDB/FerretDB/releases/tag/v1.15.0) for more on this release and all its associated packages.

We continue to be overwhelmed by the massive support from the open source community, as we deliver the truly open-source document database alternative to MongoDB.
Nowhere is this more evident than in our ever-increasing number of contributors.
In this release, we had 4 new contributors to FerretDB: [@mrusme](https://github.com/mrusme), [@cosmastech](https://github.com/cosmastech), [@chumaumenze](https://github.com/chumaumenze), and [@ksankeerth](https://github.com/ksankeerth).
Over the past three months, we've had 24 different community contributors, which is incredible!

Aside from this, FerretDB now has:

- More than 20.9k [gchr.io](https://github.com/FerretDB/FerretDB/pkgs/container/ferretdb) downloads; 500+ [downloads on Docker](https://hub.docker.com/r/ferretdb/ferretdb/tags)
- More cloud providers (like [Vultr](https://www.vultr.com/docs/ferretdb-managed-database-guide/), [Civo](https://www.civo.com/marketplace/FerretDB), and [Scaleway](https://www.scaleway.com/en/betas/#managed-document-database)) offering FerretDB as a managed service
- Over 7.9k stars on GitHub

Your support is greatly appreciated!

Everyone is welcome to contribute to FerretDB via code, bug reports, feature request, or documentation ([see our contribution guide for more](https://docs.ferretdb.io/contributing/)).
If you have questions or feedback on FerretDB, [contact us on our community channels](https://docs.ferretdb.io/#community).
