---
slug: ferretdb-v1130-new-postgresql-backend
title: FerretDB releases v1.13.0 with new PostgreSQL backend
authors: [alex]
description: >
  The release of FerretDB v1.13.0 brings new changes, which include the new PostgreSQL backend – now enabled by default – `arm/v7` packages, and a default SQLite directory for Docker images.
image: /img/blog/ferretdb-v1.13.0.png
tags: [release]
---

![FerretDB v.1.13.0 - new release with new PostgreSQL backend](/img/blog/ferretdb-v1.13.0.png)

The release of FerretDB v1.13.0 brings new changes, which include the new PostgreSQL backend – now enabled by default – `arm/v7` packages, and a default SQLite directory for Docker images.

<!--truncate-->

Let's see what's changed.

## New PostgreSQL Backend

In the past couple of months, we've been working on the new PostgreSQL backend, and we're thrilled to finally make it available by default.
Nevertheless, if you still want to enable the old backend, you can do so with the `--postgresql-old` flag or `FERRETDB_POSTGRESQL_OLD=true` environment variable.
Please note however that this will be removed in the next release.

## Default SQLite directory for Docker images and `arm/v7` packages now available

The new release includes changes to the default SQLite directory for Docker images.
Our Docker images now use `/state` directory for the SQLite backend.
That directory is also a Docker volume, so data will be preserved after the container restart by default.

Please note that this change does not affect binaries and `.deb`/`.rpm` packages.

In addition, the new version of FerretDB now offers `linux/arm/v7` binaries, Docker images, and `.deb`/`.rpm` packages.
[See them here](https://github.com/FerretDB/FerretDB/releases/).

## Other changes

In line with improving performance on FerretDB, we've implemented more pushdowns for the PostgreSQL backend in this release.
We've also added a filter pushdown for `_id: <string>` for the SQLite backend.
See here to [learn about pushdowns on FerretDB](https://blog.ferretdb.io/ferretdb-v-0-9-3-improved-aggregation-pipeline-support/).

We've also implemented some of the missing fields for `collStats`, `dbStats`, and `aggregate` `$collStats`.

Additionally, we've resolved the bug caused by invalid validation for the `_id` field, where attempting to insert a document such as this `db.test.insertOne({_id: 1, v:{_id: ["foo", "bar"]}})` returns an error that the "`_id` value cannot be of type array".

This release also includes basic logging for the PostgreSQL backend.

There were a lot of changes in this release, and while we've not touched on everything in this blog post, you can [find the rest of the changes here](https://github.com/FerretDB/FerretDB/releases/tag/v1.13.0).

## An eventful season for open source!

As we wind down the Hacktoberfest event, it's been an incredible period for the open-source community as many new entrants made their first contribution to an open-source project.
At FerretDB, we've seen a remarkable uptick in contributors, with many contributing for the first time.
In this new release, we had 5 new contributors: [@Akhil-2001](https://github.com/Akhil-2001), [@sid-js](https://github.com/sid-js), [@codenoid](https://github.com/codenoid), [@chanon-mike](https://github.com/chanon-mike), and [@pvinoda](https://github.com/pvinoda).
We appreciate all the contributions, bug reports, and feedback from everyone.

Please know you are always welcome to contribute to FerretDB any time, and we can't wait to welcome more open-source enthusiasts in the coming weeks and months.
If you don't know what to start with, [reach out to us on any of our community channels](https://docs.ferretdb.io/#community) and we'll be happy to help you get started.
