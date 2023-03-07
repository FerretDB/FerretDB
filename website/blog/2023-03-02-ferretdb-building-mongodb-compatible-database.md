---
slug: ferretdb-building-mongodb-compatible-database
title: "Building an Open-Source, MongoDB-Compatible Database - Our Journey to 1.0 and Beyond"
authors: [peter]
description: >
    We are building FerretDB, an open source document database with MongoDB compatibility - and this is why
image: /img/blog/ferretdb-monogdb.jpg
tags: [document database, mongodb alternative, mongodb compatible]
---

![FerretDB - Building MongoDB compatibility](/img/blog/ferretdb-monogdb.jpg)

We are building FerretDB, an open source document database with MongoDB compatibility.
For those who are new to our story, let me explain why:

<!--truncate-->

[MongoDB](https://www.mongodb.com/) is a technology with excellent developer experience and very strong ecosystem.
However, since it was abruptly [moved from a fully open source license to a proprietary one (SSPL)](https://blog.ferretdb.io/open-source-is-in-danger/), the market has been searching for an alternative.
While there is a wide selection of open source document databases, none of them are compatible with MongoDB.

This is why we founded [FerretDB](https://www.ferretdb.io/): we are on a mission to bring MongoDB database workloads back to its open source grounds.
We are enabling [PostgreSQL](https://www.postgresql.org/) and other database backends to run MongoDB workloads, retaining the opportunities provided by the existing ecosystem around MongoDB.

## So where are we now?

A bit more than a year after we started working on FerretDB full time, we are a team of 9 people, working full-time and remotely from all over the world.
We are executing on our [roadmap](https://github.com/orgs/FerretDB/projects/2), and are planning to release our 1.0 GA by the end of March.

## Where are we headed?

We would like to create a new standard for document databases, with [MongoDB compatibility](https://blog.ferretdb.io/mongodb-compatibility-whats-really-important/).
FerretDB is a drop-in replacement for MongoDB, but it also aims to set a new standard which not only brings easy to use document databases back to its open source roots, but also enables different database engines to run document database workloads in a standardized way.

Also, we are building an easy to use database offering the flexibility of document databases, with the possibility to run queries on the same database using SQL.

## What is the FerretDB GA going to be capable of?

FerretDB 1.0 GA includes all the essential features capable of running document-type workloads.

We are testing FerretDB 0.9.2 against real-world applications, like [BigBlueButton](https://bigbluebutton.org/), [Strapi](https://strapi.io/), or frameworks such as [MeteorJS](https://www.meteor.com/), among others.
We also confirmed that popular database management software such as [Mongosh/MongoDB Compass](https://www.mongodb.com/docs/compass/current/embedded-shell/), [NoSQL Booster](https://nosqlbooster.com/downloads), [Mingo](https://mingo.io/) are able to leverage the current feature set of FerretDB.
It’s like managing a MongoDB database, but it is FerretDB (and open source) under the hood.
We think this is insanely cool!

## What database backends does FerretDB support?

FerretDB offers a fully pluggable architecture, we support other backends as well, and these can be contributed by the community.

### PostgreSQL

We are building FerretDB on PostgreSQL and we envision that this is going to be our main database backend in the foreseeable future.
This is the backend which will get the newest features of FerretDB.
We are implementing features adding MongoDB compatibility to Postgres mainly using its JSON capabilities.
However, we recognize that we need to depart from this approach to increase performance by creating our own extension or through other methods.

### Tigris

We partnered with [Tigris Data](https://www.tigrisdata.com/) to add support for Tigris, a backend which offers a fully managed solution and [an alternative to MongoDB Atlas](https://blog.ferretdb.io/how-to-keep-control-of-your-infra-using-ferretdb-and-tigris/).
You can try it out on [their website](https://www.tigrisdata.com/).

### SAP HANA

Our friends at [SAP](https://www.sap.com/index.html) are currently working on adding [SAP HANA](https://www.sap.com/products/technology-platform/hana.html) compatibility to FerretDB, which we are very excited about.
It is also great to see SAP’s commitment to open source.

### Future database backends

We are open to adding other backends, currently playing with the idea of adding basic support for [SQLite](https://www.sqlite.org/), opening the possibility of using FerretDB in a low footprint, embedded environment.

## Where can I run FerretDB?

FerretDB can be run locally, and we are also partnering with cloud providers to make it available to you as a preview.
Since FerretDB is released under Apache 2.0, unlike MongoDB, it can be run anywhere for free.
See [our guide on how to get started.](https://docs.ferretdb.io/category/quickstart/)

## What features are you focusing on?

We think that building out support for aggregation pipelines is our most important task, as the vast majority of MongoDB applications naturally expect this feature.
Also, implementing support for indexes as any real database should have that feature.

## How can I help?

* Provide feedback
* Contribute
* Partner with us

We are working with software publishers, infrastructure providers and the maintainers of popular JS frameworks to create compatibility with their software.

If you think that your software would benefit from having an alternative database solution it can run with - [please let us know](https://www.ferretdb.io/contact/).
We are happy to work with you.

We are available on our Community Slack, on GitHub, and we are also ready to jump on a call with you if you have ideas.
[Contact us today.](https://docs.ferretdb.io/#community)
