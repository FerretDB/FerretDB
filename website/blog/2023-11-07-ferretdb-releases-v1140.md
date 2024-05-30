---
slug: ferretdb-releases-v1140
title: FerretDB releases v1.14.0!
authors: [alex]
description: >
  We’re happy to announce the release of FerretDB v.1.14.0 which includes the implementation of the `compact` command, optimized `insert`, and more.
image: /img/blog/ferretdb-v1.14.0.jpg
tags: [release]
---

![FerretDB releases v.1.14.0](/img/blog/ferretdb-v1.14.0.jpg)

We're happy to announce the release of FerretDB v.1.14.0 which includes the implementation of the `compact` command, optimized `insert`, and more.

<!--truncate-->

As we mentioned in our [previous release (v.1.13.0)](https://blog.ferretdb.io/ferretdb-v1130-new-postgresql-backend/), we've migrated to the new PostgreSQL backend, and the old PostgreSQL backend code is now completely removed in this release.
If you're unfamiliar with the new PostgreSQL backend, [here's an overview that should help](https://blog.ferretdb.io/ferretdb-v1-10-production-ready-sqlite/#the-new-architecture).

It's been an exciting time at FerretDB and the open source community as we rounded up the last of the Hactoberfest PRs.
We celebrate all the new contributors to FerretDB and their support for a truly open source alternative to MongoDB.
In this latest release, we had 3 new contributors to FerretDB: [@ShatilKhan](https://github.com/ShatilKhan), [@rubiagatra](https://github.com/rubiagatra), and [@gen1us2k](https://github.com/gen1us2k).

Let's check out the new features.

## New features and enhancements

In the latest release, we've implemented the `compact` command.
The command defragments and reorganizes the data within a collection to free up storage and improve efficiency.

We've also optimized `insert` performance by batching, so instead of inserting documents one after the other, the new approach batches multiple documents into a single `Collection.InsertAll` call for both PostgreSQL and SQLite backends.

We're also currently working on supporting capped collections, and we should have more updates on this in future releases, so keep an eye out for it.

For a complete list of the new features and enhancements, [check out the release notes](https://github.com/FerretDB/FerretDB/releases/tag/v1.14.0)

## Our incredible community

This past month, we've had numerous first-time code contributions, bug reports, feature requests, community discussions, and feedback.
The support we've received on FerretDB has been incredible, and we're so happy to have such a vibrant community.

At [FerretDB](https://www.ferretdb.com/), we are committed to providing an open source document database with MongoDB compatibility, which is now available on several cloud platforms.
In addition to [Civo](https://www.civo.com/marketplace/FerretDB) and [Scaleway](https://www.scaleway.com/en/betas/#managed-document-database), we're delighted to announce that FerretDB is [now available on Vultr](https://www.vultr.com/docs/ferretdb-managed-database-guide/), too.

If you have questions or feedback on FerretDB, [contact us on our community channels](https://docs.ferretdb.io/#community).
