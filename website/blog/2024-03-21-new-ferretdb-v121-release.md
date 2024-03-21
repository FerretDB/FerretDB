---
slug: new-ferretdb-v121-release
title: FerretDB releases v1.21 with experimental support for SCRAM-SHA-1/SCRAM-SHA-256
authors: [alex]
description: >
  We have released FerretDB v1.21, and it includes experimental support for the SCRAM-SHA-1 and SCRAM-SHA-256 authentication mechanisms.
image: /img/blog/ferretdb-v1.21.0.jpg
tags: [release]
---

![FerretDB v1.21](/img/blog/ferretdb-v1.21.0.jpg)

In the latest [FerretDB](https://www.ferretdb.com/) v1.21 release, we added experimental support for the `SCRAM-SHA-1`/`SCRAM-SHA-256` authentication mechanisms.

<!--truncate-->

This release blog post will delve into how the new authentication mechanisms work, and our ongoing efforts to improve FerretDB.

## Experimental support for SCRAM-SHA-1/SCRAM-SHA-256

Earlier this year, we began work on enabling support for the `SCRAM-SHA-1` and `SCRAM-SHA-256` authentication mechanisms, and these are now available for experimental purposes.

Enable the new authentication mode by running FerretDB with the flag `--test-enable-new-auth`/`FERRETDB_TEST_ENABLE_NEW_AUTH` environment variable.

Once a user is created using the `createUser` command, you can use the created user credential in the connection string to connect to FerretDB.

We encourage you to try out the new authentication mechanisms and let us know what you think.

## Bug fixes and enhancements

In this release, we've reorganized `upsert` handling in `update` and `findAndModify` commands, fixing an issue where filter fields were incorrectly ignored and not appended when `upsert: true`.

In addition to that, we improved the cleanup logic for capped collections in FerretDB to correctly handle document deletion when collections are configured with a size parameter without a `max` option.

Check out our latest [release notes for the complete list](https://github.com/FerretDB/FerretDB/releases/tag/v1.21.0) of changes in this release.

## FerretDB v2.0: it's coming!

Almost a year ago, we released FerretDB 1.0 GA as the first open source alternative to MongoDB, based on Postgres.
Our aim with 1.x was twofold: first, we wanted to provide a production-ready alternative to encourage the open source community to try FerretDB and provide feedback for us on their experience across a variety of different use cases and platforms.
Today, with thousands of running FerretDB around the world, we are able to understand better what the expectations are.
Your feedback, contributions and telemetry data helped us to fine-tune our roadmap.

While we continued working on FerretDB 1.x releases, but in the background, we also turned our attention to addressing the elephant in the room: performance.

The latest release is going to be one of the last releases of FerretDB v1.x.
FerretDB v2.0 is just around the corner, with drastically improved performance and compatibility.
FerretDB 2.0 will be a departure from our current architecture, which enables us to roll out these enhancements.

We understand the need to make it easy for users to switch by offering a truly open source alternative that is compatible with many of their existing needs.

So we believe FerretDB v2.0 should perform even better, enable even greater scalability and compatibility to support more application use cases.

We hope to have it available for our users very soon.

The open source community plays an integral role in everything we do at FerretDB, and we appreciate all the support from everyone, including partners, code contributors, and well-wishers.
Particularly for this release, we are grateful to [@farit2000](https://github.com/farit2000), [@sbshah97](https://github.com/sbshah97) as first-time contributors to FerretDB.

If you have any question at all about FerretDB, please feel free to [reach out on any of our channels here](https://docs.ferretdb.io/#community).
