---
slug: ferretdb-new-release-v180
title: FerretDB v1.8.0 – New release
authors: [alex]
description: >
  FerretDB v1.8.0 has just been released with new features, bug fixes, and improvements to the SQLite and PostgreSQL backend.
image: /img/blog/ferretdb-v1.8.0.jpg
tags: [release]
---

![FerretDB v.1.8.0 - new release](/img/blog/ferretdb-v1.8.0.jpg)

FerretDB v1.8.0 has just been released with new features, bug fixes, and improvements to the SQLite and PostgreSQL backend.

<!--truncate-->

In the last few weeks we've added support for the `$group` stage `_id` expression and the `$expr` evaluation query operator, and also made several optimizations and resolved issues with some of the previously implemented commands, our integration tests, and among others.

These are quite times for many people across the globe, including FerretDB maintainers and contributors, as we enjoy the summer vacations and the warm weather.

At [FerretDB](https://www.ferretdb.io), we see this period as an exciting one!
For instance, with the latest improvements and additions in FerretDB v1.8.0, we are getting closer to the production-ready release of the SQLite backend for FerretDB, which should be available soon.
In addition to that, we'll be attending [Postgres Ibiza 2023 (August 29-31, 2023)](https://pgibz.io/) and [Civo Navigate (September 5-6, 2023)](https://www.civo.com/navigate).
Our CEO, Peter Farkas, will be giving a presentation at both events – we hope to see you there!

Without further ado, let's check out some of the notable features in this release!

## New features

In the new release, we've added support for the `$group` stage `_id` expression.

With this update, users can now use more complex document structures for the `_id` field in `$group` aggregations; you can now group data using an expression like `{"$group": {_id: {"v": "$v"}}}`.

We've also added support for the `$expr` evaluation query operator.
This operator allows you to use aggregation expressions within the `$match` aggregation stage and the `find` command.
For example, you can now execute queries like `db.values.find( { $expr: { $gt: [ "$v" , 2] } } )`.
For the `$match` aggregation stage, you can use it like this: `{ $match: { $expr: { <aggregation expression> } } }`.

## Bug fixes

We've fixed a bug that caused the `findAndModify` command to return an immutable `_id` error when upserting the same `_id`.
Previously, one of our users experienced an error when executing an `updateOne` operation that upserts the same `_id`, for example, `db.test.updateOne({ _id: 1}, { $set: { _id: 1, a: 1 }})`.
With this bug fix, users can now execute such commands without encountering an error.

## Other changes

In this release, we've improved on our SQLite backend by implementing metadata caching.
Additionally, we've integrated transaction locks for collection creation and drops, ensuring database integrity during these critical operations.
This comes as part of our commitment to enabling and improving the new backend architecture.
In the coming weeks, we'll be adding more features to the SQLite backend, so stay tuned!

Of course there are other important changes in this release that we haven't mentioned here.
You can check out the full list of changes in the [release notes](https://github.com/FerretDB/FerretDB/releases/tag/v1.8.0)

We're delighted with all the support and contributions from everyone in the open source community, and we hope you're enjoying your summer!
If you can, do catch us at Postgres Ibiza 2023 and Civo Navigate – we'd love to meet you!

If you have any questions for us, and you'd like to get in touch, please [reach out to us on any of our community channels](https://docs.ferretdb.io/#community) – we can't wait to hear from you!
