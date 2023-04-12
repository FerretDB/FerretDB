---
slug: ferretdb-1-0-ga-opensource-mongodb-alternative
title: "Announcing FerretDB 1.0 GA - a truly Open Source MongoDB alternative"
authors: [peter]
description: >
    After many months of development, FerretDB is now production-ready. We are excited to announce the general availability of FerretDB v1.0.
image: /img/blog/ferretdb-v1.0.jpg
tags: [document database, mongodb alternative, mongodb compatible]
---

![Announcing FerretDB 1.0 GA - MongoDB compatibility](/img/blog/ferretdb-v1.0.jpg)

## Consider starring us on GitHub: [https://github.com/FerretDB/FerretDB](https://github.com/FerretDB/FerretDB)

After several months of development, [FerretDB](https://www.ferretdb.io/) is now production-ready.
We are excited to announce the general availability of FerretDB, a truly Open Source MongoDB alternative, built on PostgreSQL, and released under the Apache 2.0 license.

<!--truncate-->

MongoDB is [no longer open source](https://blog.opensource.org/the-sspl-is-not-an-open-source-license).
We want to bring MongoDB database workloads back to its open source roots.
We are enabling [PostgreSQL](https://www.postgresql.org/) and other database backends to run MongoDB workloads, retaining the opportunities provided by the existing ecosystem around MongoDB.

* Deploy anywhere + stay in control of your data
* Use it freely for your cloud-based projects
* Use your existing PostgreSQL infra to run MongoDB workloads

## How to get started

We offer Docker images for production use and development, as well as RPM and DEB packages.
If you would like to test FerretDB, we provide an All-in-one Docker image, containing everything you need to evaluate FerretDB with PostgreSQL.
[Get started with the All-in-one package here.](https://github.com/FerretDB/FerretDB#quickstart)

Additionally, thanks to our partners, FerretDB is available on two cloud providers for testing:

* Scaleway ([see their blog post for more information](https://www.scaleway.com/en/blog/ferretdb-open-source-alternative-mongodb/))
* On the [Civo Marketplace](https://www.civo.com/marketplace/FerretDB)

## Main feature additions to GA

In this GA release, FerretDB now supports the `createIndexes` command.
This will enable you to specify the fields you want to index, and also the type of index to use (e.g. ascending, descending, or hashed).

For instance, suppose you have a `users` collection containing several fields, including "age", "name", and "email", and you want to create an index on the "age" field.
You can now run the following command:

```js
db.users.createIndex({ age: 1 })
```

This will create an ascending index on the "age" field, which will speed up any queries that filter on that field.

We've also added the `dropIndexes` command, which allows you to remove an index from a collection.
Here's an example:

```js
db.users.dropIndex({ age: 1 })
```

This will remove the index from the "users" collection.

We have expanded our aggregation pipeline functionality to include additional stages, such as `$unwind`, `$limit` and `$skip`, in addition to the `$sum` accumulator within the `$group` stage.
With these additions, we can perform more refined calculations and manipulations of collection data.
In addition to these, we also added support for `count` and `storageStats` fields in `$collStats` aggregation pipeline stage.

To help you gather more information about your collections, databases, and server performance, we've enabled partial support for several server commands, including `collStats`, `dbStats`, and `dataSize`.

To retrieve statistics about a collection, use the `collStats` command like this:

```js
db.runCommand({ collStats: "users" })
```

If the statistics is about the database, run the command below:

```js
db.runCommand({ dbStats: 1 })
```

For the total data size of a collection, run the following command:

```js
db.runCommand({ dataSize: "<database>.<collection>" })
```

## So where are we now?

With the release of FerretDB 1.0 GA, no breaking changes will be introduced in the upcoming minor versions.

We are also proud to announce that FerretDB now has:

* 👨🏻‍💻 Over 40 code contributors with more than 130 merged pull requests from our community (see our thank you notes below)
* ⭐️ Over 5300 Stars and 200 Forks on GitHub
* 🔥 Over 100 running instances with [telemetry enabled](https://docs.ferretdb.io/telemetry/)
* ⏫ Over 10000 FerretDB downloads

With the release of FerretDB 1.0, these numbers will only continue to grow.

We are executing on our [roadmap](https://github.com/orgs/FerretDB/projects/2), and are planning to add more significant features in the coming months.
Get started with FerretDB 1.0 GA [here](https://github.com/FerretDB/FerretDB#quickstart).

## Where are we headed?

We are creating a new standard for document databases with [MongoDB compatibility](https://blog.ferretdb.io/mongodb-compatibility-whats-really-important/).
FerretDB is a drop-in replacement for MongoDB, but it also aims to set a new standard which not only brings easy to use document databases back to its open source roots, but also enables different database engines to run document database workloads in using a standardized interface.

Also, FerretDB will offer the flexibility of document databases, with the possibility to run queries on the same data set using SQL.

## What is FerretDB GA capable of?

FerretDB 1.0 GA includes all the essential features capable of running document database workloads.

We are testing FerretDB against real-world applications, like [FastNetMon](https://fastnetmon.com/docs-fnm-advanced/using-fastnetmon-advanced-with-ferretdb-and-postgresql-instead-of-mongodb/), [the Enmeshed Connector](https://enmeshed.eu/blog/announcing-ferretdb-compatibility), [BigBlueButton](https://bigbluebutton.org/), [Strapi](https://strapi.io/), or frameworks such as [MeteorJS](https://www.meteor.com/), among others.

We also confirmed that popular database management tools such as `mongosh`, [MongoDB Compass](https://www.mongodb.com/products/compass), [NoSQL Booster](https://nosqlbooster.com/), [Mingo](https://mingo.io/) are able to leverage the current feature set of FerretDB.

It’s like managing a MongoDB database, but it is FerretDB (and open source) under the hood.
We think this is insanely cool!

## What database backends does FerretDB support?

FerretDB offers a fully pluggable architecture, we support other backends as well, and support for these can be contributed by the community.

### PostgreSQL

We are building FerretDB on PostgreSQL and we envision that this is going to be our main database backend in the foreseeable future.
This is the backend which will get the newest features of FerretDB.
We are implementing features adding MongoDB compatibility to Postgres mainly using its JSONB capabilities.
However, we recognize that we will need to depart from this approach to increase performance by creating our own extension or through other methods.

### Tigris

We partnered with [Tigris Data](https://www.tigrisdata.com/) to add support for Tigris, a backend which offers a fully managed solution and [an alternative to MongoDB Atlas](https://blog.ferretdb.io/how-to-keep-control-of-your-infra-using-ferretdb-and-tigris/).
Backend support for Tigris is now maintained by Tigris Data.

You can try it out on [their website](https://www.tigrisdata.com/).

### SAP HANA

Our friends at [SAP](https://www.sap.com/index.html) are currently working on adding [SAP HANA](https://www.sap.com/products/technology-platform/hana.html) compatibility to FerretDB, which we are very excited about.
It is also great to see SAP’s commitment to open source.

### SQLite and future database backends

We are open to adding support to other backends.
Currently, we are in the process of adding basic support for [SQLite](https://www.sqlite.org/).

## How can I help?

* Provide feedback
* Contribute
* Partner with us

We are working with software publishers, infrastructure providers and the maintainers of popular JS frameworks to create compatibility with their software.

If you think that your software would benefit from having an alternative to MongoDB - [please let us know](https://www.ferretdb.io/contact/).
We are happy to work with you.

We are available on our Community Slack, on [GitHub](https://github.com/FerretDB/FerretDB), and we are also ready to jump on a call with you if you have ideas to discuss.
[Contact us today.](https://docs.ferretdb.io/#community)

## Thank you

We extend our sincerest thanks to our dedicated community for all the contributions and feedback leading to FerretDB GA.

Thank you to [Instaclustr](https://www.instaclustr.com/) for their code contributions and feedback, and [Tigris](https://www.tigrisdata.com/) as a supporter and first user of FerretDB in production.

Special thanks go to: [seeforschauer](https://github.com/seeforschauer), [ribaraka](https://github.com/ribaraka), [ekalinin](https://github.com/ekalinin), [fenogentov](https://github.com/fenogentov), [OpenSauce](https://github.com/OpenSauce), [GinGin3203](https://github.com/GinGin3203), [DoodgeMatvey](https://github.com/DoodgeMatvey), [pboros](https://github.com/pboros), [lucboj](https://github.com/lucboj), [nicolascb](https://github.com/nicolascb), [ravilushqa](https://github.com/ravilushqa), [fcoury](https://github.com/fcoury), [AlphaB](https://github.com/AlphaB), [peakle](https://github.com/peakle), [jyz0309](https://github.com/jyz0309), [klokar](https://github.com/klokar), [thuan1412](https://github.com/thuan1412), [kropidlowsky](https://github.com/kropidlowsky), [yu-re-ka](https://github.com/yu-re-ka), [jkoenig134](https://github.com/jkoenig134), [ndkhangvl](https://github.com/ndkhangvl), [codingmickey](https://github.com/codingmickey), [zhiburt](https://github.com/zhiburt), [ronaudinho](https://github.com/ronaudinho), [si3nloong](https://github.com/si3nloong), [folex](https://github.com/folex), [GrandShow](https://github.com/GrandShow), [narqo](https://github.com/narqo), [taaraora](https://github.com/taaraora), [muyouming](https://github.com/muyouming), [junnplus](https://github.com/junnplus), [agneum](https://github.com/agneum), [radmirnovii](https://github.com/radmirnovii), [ae-govau](https://github.com/ae-govau), [hugojosefson](https://github.com/hugojosefson).

There are others, of course, who have helped in many ways by identifying bugs, providing feedback, and testing FerretDB, we thank you all!

We are happy to welcome new members to our growing community and encourage everyone to join us on [GitHub](https://github.com/FerretDB/FerretDB/), and explore the many possibilities FerretDB has to offer.
