---
slug: ferretdb-v1-10-production-ready-sqlite
title: FerretDB v1.10 – Production-ready SQLite support and more
authors: [aleksi]
description: >
  We are happy to announce the full support for the SQLite backend and discuss our new architecture.
image: /img/blog/post-cover-image.jpg # FIXME
tags: [release]
---

<!-- FIXME banner image -->

We just released the new version of FerretDB – v1.10, which is [available in the usual places](https://docs.ferretdb.io/quickstart-guide/).
With it, the SQLite backend support is officially out of beta, on par with our PostgreSQL backend, and fully supported.

<!--truncate-->

In the past two weeks, we added support for many missing commands and features,
including aggregation pipelines, indexes, query explains, and collection renaming.
Please refer to the [full release notes](https://github.com/FerretDB/FerretDB/releases/tag/v1.10.1) for details.
There are still opportunities for improvement, such as support for more [query pushdowns](https://docs.ferretdb.io/pushdown/),
but they will not break the compatibility.
You can start to use the SQLite backend in production now!
It is ideal if you need a MongoDB-compatible database that does need high performance but could work in environments
[where MongoDB can't](https://github.com/FerretDB/FerretDB/discussions/1266) and without a separate process.

You may wonder how we made such huge progress after a few relatively quiet releases.
That's because we worked on the new FerretDB architecture in the background, and the new SQLite backend is the first part of it.
Let's talk about it.

## The new architecture

When FerretDB (née MangoDB) was [first released](2021-11-05-mangodb-overwhelming-enthusiasm-for-truly-open-source-mongodb-replacement.md),
it supported a single PostgreSQL backend.
The simplified version of the architecture looked like this:

![Old MangoDB architecture](/img/blog/new-arch/image1.png)

The protocol handling module was (and still is) responsible for handling client connections, MongoDB wire protocol,
and BSON documents encoding/decoding.
The commands handling module was responsible for handling protocol commands such as `find`, `update`, and `delete`.

Over time, we understood that there is some common code between command handlers.
We extracted a part responsible for accessing PostgreSQL and running queries into the backend package:

![Old FerretDB architecture](/img/blog/new-arch/image2.png)

When we added another backend, we could not figure out if there was a common interface for different storages as they were so different.
So we decided to make another commands handling module that was similar in things like command parameters extraction
but very different in the way it interacted with storage.
Similar parts were extracted into a set of common packages:

![FerretDB with Tigris architecture](/img/blog/new-arch/image3.png)

Over time the common part grew, but we still had very different backend packages.
When we decided to add SQLite support, we wanted to revisit our previous decisions –
could we design a common storage interface after all?
We did not expect many people to use SQLite with a lot of data, so performance could have a bit lower priority,
and the interface could be relatively simple.
At the same time, we had a lot of practical knowledge of how FerretDB is used with PostgreSQL backend,
what is important and what is not, and what problems the current backend has and how to avoid them.
So we settled on this architecture:

![FerretDB with SQLite architecture](/img/blog/new-arch/image4.png)

The common backend interface is relatively small to make it easy to add new backends.
For example, there is a single method that backends should implement to insert given documents.
The code for processing ordered and unordered insertions is written once in the command handler;
there is no need to repeat that logic for every backend.
Of course, if there is some storage that could perform this operation more efficiently,
there could be a separate optional method that could be implemented.
Then the command handler could check if that method is implemented by the current backend and use it.

That architecture also allows us to add optional functionality to all backends at once.
For example, there is already a _very early_ prototype of the OpLog functionality.
It works by wrapping `insert`, `update`, and `delete` methods of any backend and inserting documents
into operation log collection when wrapped methods succeed.
The more advanced backend-specific methods of implementing that functionality could be implemented if needed.
For example, PostgreSQL could use triggers for achieving better performance.
That leads us to the discussion of the FerretDB future.

## The future

We will continue working on the SQLite backend, improving performance and compatibility.
We will continue working on the soon-to-be-only commands handler module.
And we just started working on the new PostgreSQL backend:

![New FerretDB architecture](/img/blog/new-arch/image5.png)

We don't plan to introduce any breaking changes there.
Soon you will be able to update FerretDB to the latest version that will use the same data layout in PostgreSQL,
but under the hood, the new backend code will be working.
After that, we will start working on bringing that prototype to production.
Working on a feature like that will be much easier for us with more backends getting it «for free» with less code to maintain.
And for you, that means more features delivered faster!

You can always see what else we planned in our [public roadmap](https://github.com/orgs/FerretDB/projects/2/views/1).
FIXME contributors, hiring
