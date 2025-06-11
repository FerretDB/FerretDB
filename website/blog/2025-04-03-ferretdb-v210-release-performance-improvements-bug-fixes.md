---
slug: ferretdb-v210-release-performance-improvements-bug-fixes
title: 'FerretDB releases v2.1.0 with performance and bug fixes'
authors: [alex]
description: >
  FerretDB v2.1.0 brings performance improvements and key bug fixes.
image: /img/blog/ferretdb-v2.1.0.jpg
tags: [release]
---

![FerretDB releases v2.1.0 with performance improvements and key bug fixes](/img/blog/ferretdb-v2.1.0.jpg)

FerretDB v2.1.0 is now available.

<!--truncate-->

This release builds on the [FerretDB v2.0 GA milestone](https://blog.ferretdb.io/ferretdb-v2-ga-open-source-mongodb-alternative-ready-for-production/) with performance improvements and key bug fixes.

## Important update information

Due to incompatibilities in our previous release, DocumentDB can't be updated directly.
Users must backup their current databases and conduct a clean installation of FerretDB v2.1.0 and the DocumentDB extension before restoring their data.

To update to FerretDB v2.1.0, please follow the same instructions as in our [migration guide](https://docs.ferretdb.io/migration/migrating-from-mongodb/) to:

- Back up your data using `mongodump` or `mongoexport`.
- Remove existing FerretDB and DocumentDB images, containers, and volumes (or Debian packages and data directories).
- Install [FerretDB v2.1.0](https://github.com/FerretDB/FerretDB/releases/tag/v2.1.0) and [DocumentDB extension](https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.1.0).
- Restore your data with `mongorestore` or `mongoimport`.

This is a one-time manual process.
Future updates will be much smoother.

## What's new in v2.1.0

### Indexing issue resolved

We fixed an issue related to indexing that was present in previous versions.
With this fix, you can expect more reliable and efficient index operations.

We recommend using the `reIndex` command to rebuild your indexes after updating to FerretDB v2.1.0.
This will ensure that all indexes are up to date and functioning correctly.

### Fixed embeddable package

The [embeddable Go package](https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/ferretdb), broken in v2.0.0, now works as expected, making it easier to use FerretDB as a library in Go applications.

### Other changes

FerretDB v2.1.0 provides better performance thanks to fixes in DocumentDB that improve how document updates are handled.

The `console` logger now supports colorized log levels.
We also made improvements to the `--help` output to align with our documentation, which should offer clearer guidance for users setting up or troubleshooting FerretDB.

To better integrate with Docker Secrets, FerretDB now supports reading the PostgreSQL connection URL from a file using the `--postgresql-url-file` flag or the `FERRETDB_POSTGRESQL_URL_FILE` environment variable.

See the [release notes for all changes in this release](https://github.com/FerretDB/FerretDB/releases/tag/v2.1.0).

## Looking ahead

Since the release of FerretDB v2.0.0, we have been working to address the feedback and issues reported by our users.
We are committed to providing a truly open-source alternative to MongoDB that's highly performant, compatible, and reliable for all your database needs.

Visit [our GitHub](https://github.com/FerretDB) and [our website](https://www.ferretdb.com) to download,
contribute, or explore enterprise solutions.

If you have any questions, reach out to us on [our community channels](https://docs.ferretdb.io/#community).
