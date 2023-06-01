---
slug: ferretdb-v-1-2-0-minor-release
title: FerretDB v1.2.0. Minor Release
authors: [alex]
description: >
  FerretDB version 1.2.0 has just been released with minor changes that fixes the bugs in the previous release and enhance existing features and processes.
image: /img/blog/ferretdb-1-2-0-minor-release.jpg
tags: [release]
---

![FerretDB v.1.2.0](/img/blog/ferretdb-1-2-0-minor-release.jpg)

FerretDB v1.2.0 has just been released with minor changes that address some of the bugs in previous versions and enhance existing features and processes.

<!--truncate-->

One of the major updates we're excited about is the initial progress towards supporting [SQLite](https://www.sqlite.org/) database backend.
This release includes a highly experimental SQLite backend.
While this isn't ready yet, there are a lot of promising developments on the way that are sure to appeal to [FerretDB](https://www.ferretdb.io/) users and admirers.

The potential implementation of SQLite reflects our goal to provide an open-source database alternative to MongoDB and also enable more database backend support beyond [PostgreSQL](https://www.postgresql.org/).

Although this release doesn't come with any major feature additions, we are proud of our growing list of contributors, especially the two amazing new contributors in this release: [@adetunjii](https://github.com/adetunjii) and @[christiano](https://github.com/christiano).
We appreciate every single line of code, bug discovery, comments, and feedback from the open-source community.

Let's take a look at some of the significant changes in this release.

## Fixed bugs and enhancements

We've fixed a bug with unset field sorting.
This bug interfered with the proper sorting of documents that had unset fields.
Interestingly, this bug didn't show up consistently but only under specific conditions.
With a single sort field, everything worked as expected, but in situations with multiple sort fields and the unset field wasn't the first, the sorting process didn't function correctly.

In addition to that, we've found and resolved a bug with `dbStats` and `collStats` operations, ensuring that they return `int64` values, enabling them to handle large databases and collections effectively.

We've also fixed the compatibility issue with the C# driver by allowing the driver to complete a server handshake and prevent it from sending a `getLastError`.

Another feature we've improved on in this release is to enable multiple document inserts within a single transaction for the `insertMany` command.
Initially, transactions were created for each inserted document.
This change should potentially decrease insertion time and transaction overhead for documents.

Furthermore, we've added support for dot notation in query projections.

## Other Changes

Our documentation has also been updated with a new section that explains the [installation details for .rpm packages](https://docs.ferretdb.io/quickstart-guide/rpm/).

[See here to learn more about all the other changes](https://github.com/FerretDB/FerretDB/releases/latest) that come with FerretDB v.1.2.0.

Special appreciation goes to all FerretDB contributors, users, and admirers – we're so product and delighted for all the contributions.
Plus, it's been an impressive period since we released our [first GA version back in April](https://blog.ferretdb.io/ferretdb-1-0-ga-opensource-mongodb-alternative/), with a rising list of contributors, users, comments, and followers.
We're committed and even more enthusiastic about providing the defacto open-source alternative for MongoDB.

If you have any questions or want to know more about FerretDB, please [reach out to us](https://docs.ferretdb.io/#community).
