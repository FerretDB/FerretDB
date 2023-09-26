---
slug: ferretdb-releases-v1110
title: FerretDB releases v1.11.0
authors: [alex]
description: >
  We just released FerretDB v1.11.0, which includes several improvements for the SQLite backend.
image: /img/blog/ferretdb-v1.11.jpg
tags: [release]
---

![FerretDB v.1.11.0 - new release](/img/blog/ferretdb-v1.11.jpg)

We just released FerretDB v1.11.0, which includes several improvements for the SQLite backend.

<!--truncate-->

This comes on the back of our previous release, which included full support for the SQLite backend as well as the new FerretDB architecture ([see here for more](https://blog.ferretdb.io/ferretdb-v1-10-production-ready-sqlite/)).

We've continued work on moving the PostgreSQL backend to the new architecture, which should make it easier to add new features much faster.

Besides, the feedback from the community on the SQLite backend has been very positive, and we really appreciate all the support from every corner.
We are committed to providing more features for FerretDB and resolving any issues you might encounter.
And in this release, we've addressed some of these bugs, and also provided a helpful guide for migrating data from MongoDB to FerretDB.

Let's check them out!

## Fixed bugs and enhancements

Like every release, there is always room for improvement; in this release, we've addressed some of the bugs discovered in the previous release.

The issue with `$collStats` failing to return the correct count of documents for SQLite, particularly when there are a large number of documents inserted, has been resolved.

We've also enabled `dropIndexes` to update the metadata table for the SQLite backend.
Similarly, the SQLite backend is now able to return statistics of indexes for both `collStats` and `dbStats`.

## Documentation

Many of our users have asked about a guide to efficiently migrate their data from MongoDB to FerretDB, and now it's available.

We've created a pre-migration guide that should help you prepare and test your application with FerretDB; we also describe the process and tools for migrating your MongoDB data to FerretDB.
[See them here](https://docs.ferretdb.io/category/migrating-to-ferretdb/).

## Other changes

For other changes in this release, please [check out our release notes](https://github.com/FerretDB/FerretDB/releases/tag/v1.11.0).

The community has been a key component in our growth, and we appreciate every single support we've received either via bug reports, code contributions, feature requests, suggestions, and more.

As always, if you have any questions, feedback, or requests, we would love to hear from you!
Reach out to us on any of our [community channels](https://docs.ferretdb.io/#community).
