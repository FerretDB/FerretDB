---
slug: fastnetmon-picked-ferretdb-saas-offering
title: 'Why FastNetMon picked FerretDB for their SaaS offering'
author: Pavel Odintsov
author_title: Co-founder and CTO at FastNetMon
author_image_url: /img/blog/ferretdb-fastnetmon.jpg
description: >
  One of the early adopters of FerretDB, Pavel Odintsov describes the experience of using FerretDB and why FastNetMon picked it for their SaaS offering.
image: /img/blog/ferretdb-vultr.jpg
tags: [open source, community, document databases, compatible applications]
---

![Why FastNetMon picked FerretDB for their SaaS offering](/img/blog/ferretdb-fastnetmon.jpg)

_One of the early adopters of FerretDB, Pavel Odintsov describes the experience of using FerretDB and why FastNetMon picked it for their SaaS offering._

<!--truncate-->

[FastNetMon LTD](https://fastnetmon.com/) is a leading vendor for on-premise DDoS detection for Telecom networks.
To reduce the time needed to deploy products when networks are under attack, they're working on a new offering called FastNetMon Cloud, which will allow customers to get traffic visibility and attack detection in a matter of minutes without needing to deploy new hardware and install the product.

To offer the best quality of service and at the same time to keep maintenance costs low, we need a database that does not require significant attention from us and offers outstanding performance.
At the same time, we need to have trust in the future of the database and see that the database has a clear and very predictable roadmap and a bright future backed by a great team.

## FastNetMon motivation

We've used MongoDB for more than six years and accumulated extensive operational experience with it.
Schema-less databases allow our developers to focus on product logic and ignore any database-specific issues.

Unfortunately, MongoDB-specific issues were one of the most popular among queries to our support team.
Most of the issues were caused by the Linux distribution upgrade, which involved a complicated MongoDB upgrade process that had to be executed version by version, and it was challenging to do it correctly.

Another issue with MongoDB is that support of new Linux distributions is significantly delayed up to many months, and we had to wait to offer our product just because of MongoDB's unavailability for new platforms.

One of the main reasons why we started looking for alternatives was the unconditional introduction of the AVX requirement in a new version of MongoDB, [which is explained here](https://pavel.network/please-do-not-require-avx-for-your-software/).
Due to this issue, we could only conduct upgrades for some of our customers using hardware with AVX support.

As we offer an on-premise product, we did not have any issues with the SSPL license introduction.
Still, the cloud offering poses significant potential risks with SSPL and requires extensive consultation with highly qualified legal experts.

### Technical details

We had two potential options to use [FerretDB](https://www.ferretdb.com/): [with PostgreSQL](https://fastnetmon.com/docs-fnm-advanced/using-fastnetmon-advanced-with-ferretdb-and-postgresql-instead-of-mongodb/) and [SQLite](https://fastnetmon.com/using-fastnetmon-advanced-with-ferretdb-and-sqlite-backend-instead-of-mongodb/) and we decided to stick with SQLite option as the amount of data is very low.
We clearly do not want to keep PostgreSQL daemon as an additional dependency.

FerretDB is implemented in the Go language and can work on all supported platforms.

As part of the cost evaluation, we decided to prefer ARM64 Graviton 2 and Graviton 3 CPUs for our Cloud platform, and FerretDB added support for them on the same day we asked for it.
Performance was exceptional, and with the release of [FerretDB v 1.12](https://blog.ferretdb.io/ferretdb-v112-available/#arm64-binaries-now-available), there is now official support for ARM64.
We have no AVX requirement, and we can use the same FerretDB binary for most platforms we use for our product.

We can see that the team behind FerretDB is top-notch, and we have great trust in the project's future.

## Conclusion

FerretDB offers complete compatibility with MongoDB interfaces and features simpler maintenance, fast release cycles, and a community-focused view.

That's why we're looking to pick up FerretDB as the foundation for our FastNetMon Cloud Platform.

### About FastNetMon

FastNetMon is a team of professionals and enthusiasts working in the network security area.
The company was founded in 2016 in London and operates worldwide, protecting businesses from cyber threats.

### About FerretDB

FerretDB is a truly open-source alternative to MongoDB built on Postgres.
FerretDB allows you to use MongoDB drivers seamlessly with PostgreSQL as the database backend.
Use all tools, drivers, UIs, and the same query language and stay open-source.
Our mission is to enable the open-source community and developers to reap the benefits of an easy-to-use document database while avoiding vendor lock-in and faux pen licenses.
We are not affiliated, associated, authorized, endorsed by, or in any way officially connected with MongoDB Inc. or any of its subsidiaries or its affiliates.
