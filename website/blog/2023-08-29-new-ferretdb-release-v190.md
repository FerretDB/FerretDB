---
slug: new-ferretdb-release-v190
title: FerretDB v1.9.0 – New release
authors: [alex]
description: >
  We're happy to announce the release of FerretDB v1.9.0, which comes with a couple of enhancements, fixes, and a stronger focus on delivering the most performant and optimized FerretDB architecture.
image: /img/blog/ferretdb-v1.9.0.png
tags: [release]
---

![FerretDB v.1.9.0 - new release](/img/blog/ferretdb-v1.9.0.png)

We're happy to announce the release of FerretDB v1.9.0, which comes with a couple of enhancements, fixes, and a greater focus on delivering the most performant and optimized [FerretDB](https://www.ferretdb.io) architecture.

<!--truncate-->

Recently, we've been working on a new backend architecture, along with the addition of the SQLite backend.
We are confident that these improvements will help us deliver a better optimized [open source database alternative to MongoDB](https://blog.ferretdb.io/open-source-is-in-danger/), and also allow more developers to take advantage of FerretDB.

The massive support and significant contributions from the community cannot be understated in this progress.
In this release alone, we had 4 new contributors – [@aenkya](https://github.com/aenkya), [@pratikmota](https://github.com/pratikmota), [@durgakiran](https://github.com/durgakiran), and [@slavabobik](https://github.com/slavabobik) – which is so amazing!
Surely, this is a sign of our growing community and the incredible potential of FerretDB.

Let's have a look at some of the key changes in this release.

## New changes

We're happy to announce that we've published Docker images on [Quay.io](https://quay.io/), and [you can check it out here](https://quay.io/organization/ferretdb).

We've also refactored our aggregation operators and accumulators to ensure that they take values as arguments rather than documents that they need to parse.

As part of the process in providing production-ready access to the SQLite database backend, we've also added some new metrics to the SQLite backend pool statistics.

You can check out all the [other changes in this release here](https://github.com/FerretDB/FerretDB/releases/tag/v1.9.0).

## What next?

We understand the importance of having a fully optimized and efficient database, which is also our goal.
The main focus of our work for the past few releases was the new architecture that will significantly improve the performance of existing backends and make it easier to add new ones.
The SQLite backend is a part of that work.
It should be available for production workloads soon.
We really can't wait for you to try it out!

We appreciate all the contributions, feedback, and support from the community since the project, and it's been an exciting experience every step of the way.
We'd love to hear from you if you have any questions.
Our community channels are open – [contact us here](https://docs.ferretdb.io/#community).
