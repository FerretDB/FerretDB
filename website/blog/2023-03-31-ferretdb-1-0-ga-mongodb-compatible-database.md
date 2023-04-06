---
slug: ferretdb-1-0-ga-mongodb-compatible-database
title: "Announcing FerretDB 1.0 GA - the Open-Source, MongoDB-Compatible Database"
authors: [peter]
description: >
    After many months of development, FerretDB is now production-ready. We are excited to announce the general availability of FerretDB v1.0.
image: /img/blog/ferretdb-monogdb.jpg
tags: [document database, mongodb alternative, mongodb compatible]
---

![Announcing FerretDB 1.0 GA - MongoDB compatibility](/img/blog/ferretdb-v1.0.jpg)

After several months of development, [FerretDB](https://www.ferretdb.io/) is now production-ready.
We are excited to announce the general availability of FerretDB 1.0.

<!--truncate-->

This release is a tremendous milestone representing all the efforts and hard work our amazing team has put in over the past year.

Our mission remains the same: we are building FerretDB, an open source document database with MongoDB compatibility.
For those who are new to our story, let me explain why:

[MongoDB](https://www.mongodb.com/) is a technology with excellent developer experience and very strong ecosystem.
However, since it was abruptly [moved from a fully open source license to a proprietary one (SSPL)](https://blog.ferretdb.io/open-source-is-in-danger/), the market has been searching for an alternative.
While there is a wide selection of open source document databases, none of them are compatible with MongoDB.

This is [why we founded FerretDB](https://blog.ferretdb.io/mangodb-overwhelming-enthusiasm-for-truly-open-source-mongodb-replacement/): we want to bring MongoDB database workloads back to its open source grounds.
We are enabling [PostgreSQL](https://www.postgresql.org/) and other database backends to run MongoDB workloads, retaining the opportunities provided by the existing ecosystem around MongoDB.

We extend our sincerest thanks to our dedicated community for all the contributions and feedback leading to the FerretDB GA.

For their code contributions, special thanks go to: [seeforschauer](https://github.com/seeforschauer), [ribaraka](https://github.com/ribaraka), [ekalinin](https://github.com/ekalinin), [fenogentov](https://github.com/fenogentov), [OpenSauce](https://github.com/OpenSauce), [GinGin3203](https://github.com/GinGin3203), [DoodgeMatvey](https://github.com/DoodgeMatvey), [pboros](https://github.com/pboros), [lucboj](https://github.com/lucboj), [nicolascb](https://github.com/nicolascb), [ravilushqa](https://github.com/ravilushqa), [fcoury](https://github.com/fcoury), [AlphaB](https://github.com/AlphaB), [peakle](https://github.com/peakle), [jyz0309](https://github.com/jyz0309), [klokar](https://github.com/klokar), [thuan1412](https://github.com/thuan1412), [kropidlowsky](https://github.com/kropidlowsky), [yu-re-ka](https://github.com/yu-re-ka), [jkoenig134](https://github.com/jkoenig134), [ndkhangvl](https://github.com/ndkhangvl), [codingmickey](https://github.com/codingmickey), [zhiburt](https://github.com/zhiburt), [ronaudinho](https://github.com/ronaudinho), [si3nloong](https://github.com/si3nloong), [folex](https://github.com/folex), [GrandShow](https://github.com/GrandShow), [narqo](https://github.com/narqo), [taaraora](https://github.com/taaraora), [muyouming](https://github.com/muyouming), [junnplus](https://github.com/junnplus), [agneum](https://github.com/agneum), [radmirnovii](https://github.com/radmirnovii), [ae-govau](https://github.com/ae-govau), [hugojosefson](https://github.com/hugojosefson).

There are others, of course, who have helped in many ways by identifying bugs, providing feedback, and testing FerretDB, we thank you all!

We are happy to welcome new members to our growing community and encourage everyone to join us on [GitHub](https://github.com/FerretDB/FerretDB/), and explore the many possibilities FerretDB has to offer.

## So where are we now?

With the release of FerretDB 1.0 GA, we are now production-ready!

This means you can now use FerretDB in your production environments, and we'll be here to support you along the way.

A bit more than a year since we started working on FerretDB full time, we've grown to a team of 9 people, working full-time and remotely from all over the world.

We are also proud to announce that FerretDB now has:

* 👨🏻‍💻 Over 30 code contributors with more than 130 merged pull requests from our community of contributors
* ⭐️ over 5.3k Stars on GitHub
* ⏫ More than 180 Docker hub downloads
* 🔥 Over 60 running instances
* ⏫ Over 10k FerretDB downloads

And the best part: with the release of FerretDB 1.0, these numbers will only continue to grow.

We are executing on our [roadmap](https://github.com/orgs/FerretDB/projects/2), and are planning to add more significant features in the coming months.
Get started with FerretDB 1.0 GA [here](https://docs.ferretdb.io/quickstart-guide/).

## Where are we headed?

We would like to create a new standard for document databases, with [MongoDB compatibility](https://blog.ferretdb.io/mongodb-compatibility-whats-really-important/).
FerretDB is a drop-in replacement for MongoDB, but it also aims to set a new standard which not only brings easy to use document databases back to its open source roots, but also enables different database engines to run document database workloads in a standardized way.

Also, we are building an easy to use database offering the flexibility of document databases, with the possibility to run queries on the same database using SQL.

## What is FerretDB GA capable of?

FerretDB 1.0 GA includes all the essential features capable of running document-type workloads.

Moreover, we are testing FerretDB against real-world applications, like [BigBlueButton](https://bigbluebutton.org/), [Strapi](https://strapi.io/), or frameworks such as [MeteorJS](https://www.meteor.com/), among others.
We also confirmed that popular database management software such as [Mongosh/MongoDB Compass](https://www.mongodb.com/docs/compass/current/embedded-shell/), [NoSQL Booster](https://nosqlbooster.com/downloads), [Mingo](https://mingo.io/) are able to leverage the current feature set of FerretDB.

[TODO - Add a Quote here]

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

[TODO - Add a Quote here]

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
