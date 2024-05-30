---
slug: ferretdb-meteor-mongodb-alternative
title: 'Meteor.js and FerretDB: Using an Open Source MongoDB Alternative for Your Meteor.js Apps'
authors: [alex]
description: >
  Here we explore the possible synergy and compatibility of FerretDB in Meteor.js and how you can build your applications without any concern for vendor lock-in.
image: /img/blog/ferretdb-meteor.jpg
tags: [javascript frameworks, compatible applications, open source, community]
---

![Meteor.js and FerretDB](/img/blog/ferretdb-meteor.jpg)

[Meteor.js](https://www.meteor.com/) has gained immense popularity as a full-stack JavaScript platform for developing modern web and mobile applications, thanks to its seamless data synchronization and full-stack capabilities.
Its flexibility, real-time capabilities, and rich ecosystem have made it a developer's favorite.

<!--truncate-->

However, the default database for Meteor.js applications, MongoDB, presents significant concerns about vendor lock-in, especially since its move away from its open-source roots.
To mitigate this, open-source alternatives like [FerretDB](https://www.ferretdb.io/) offer the opportunity to seamlessly replace MongoDB without compromising the commands and syntax you already know.

In this article, we'll explore the possible synergy between FerretDB and Meteor.js and how you can build your applications without any concern for lock-in.

## Brief History of Meteor.js

Meteor.js is an open-source, all-in-one JavaScript framework, available publicly under the MIT license.
Built with Node.js, Meteor.js uses MongoDB as its core database for applications, Blaze handlebars for efficient templating, a powerful pub/sub method for real-time data synchronization, and seamless integration with various Meteor.js plugins, and `npm` packages you need.

This extensive ecosystem empowers developers to rapidly create quick cross-platform prototypes or large, complex applications with ease.
With its open-source nature and availability under the MIT license, Meteor.js has gained popularity among not only mid-sized and enterprise companies like [IKEA](https://www.ikea.com/), [Workpop](https://www.betterteam.com/workpop), [Accenture](https://www.accenture.com/us-en) but also individual developers and startups.

The framework has been responsible for the development of remarkable apps, such as [Rocket.Chat](https://rocket.chat/), [Apify](https://apify.com/), and [Chatra](https://chatra.com/), among others.
Its ability to efficiently handle real-time communication and collaboration make it an ideal platform for applications requiring instant updates and dynamic content.

Meteor.js uses MongoDB as its standard database.
However, [several concerns about MongoDB](https://blog.ferretdb.io/open-source-is-in-danger/) over the years have led some to seek (or demand) an open-source alternative that sufficiently replaces MongoDB without any hassles.

Since MongoDB's license switch to SSPL from its former open-source license, there have been growing concerns about the nature of the license and how it may affect developers and companies building their applications with MongoDB.
Several Meteor.js application developers are also in this corner.

And this is what makes [the announcement of the General Availability version](https://blog.ferretdb.io/ferretdb-1-0-ga-opensource-mongodb-alternative/) of FerretDB – an open-source replacement to MongoDB – so exciting!

## Introducing FerretDB

FerretDB is the defacto open-source MongoDB alternative that converts MongoDB wire protocols with a backend on [PostgreSQL](https://www.postgresql.org/).
Besides PostgreSQL, FerretDB is also working to provide support for other database backends, such as [Tigris](https://www.tigrisdata.com/), [SAP Hana](https://www.sap.com/africa/products/technology-platform/hana.html), [SQLite](https://sqlite.org/), and many more to come.

With the FerretDB GA released in April 2023, users can now leverage FerretDB for their production workloads.
Users familiar with MongoDB can seamlessly transition since it uses the same query language and syntax.
This makes FerretDB a highly attractive option for Meteor.js developers looking to [escape the vendor lock-in challenges associated with MongoDB](https://blog.ferretdb.io/5-ways-to-avoid-database-vendor-lock-in/) while still benefiting from a familiar environment.

Some of the popular database management tools, like `mongosh`, [MongoDB Compass](https://www.mongodb.com/products/compass), [Studio3T](https://studio3t.com/), [Mingo](https://mingo.io/), and [NoSQLBooster](https://nosqlbooster.com/), are already capable of taking full advantage of the current set of features on FerretDB.

Apart from these tools, FerretDB is being tested against real-world applications, and some of the major ones are applications built with Meteor.js since they use MongoDB as their default database.

Some of the benefits of using FerretDB with your Meteor.js applications include the following:

- **Open-source nature:** With a low barrier to entry, you can get started with FerretDB and take advantage of a community of developers and engineers that's always ready to help and support you every step of the way, without any fear of vendor lock-in.
- **User-friendliness:** FerretDB lets you jump in and use the same syntax you're familiar with, so there's no need to learn a new language or syntax; it's the same familiar environment you're already used to.
- **PostgreSQL backend:** FerretDB's database engine, PostgreSQL, is renowned for its open-source nature, reliability, and robust community, as well as its ability to handle large amounts of data and complex queries, which can be an added advantage.

## Replacing MongoDB With FerretDB in Your Meteor.js Apps

FerretDB has made significant strides in ensuring compatibility with Meteor.js apps, and the development team continues to work on enhancing this compatibility further.
As of the latest release, FerretDB demonstrates compatibility with many common MongoDB use cases encountered in Meteor.js applications.

The current level of compatibility enables Meteor.js developers, already accustomed to MongoDB, to seamlessly transition to FerretDB with minimal friction.
The same query language and syntax can be directly applied to FerretDB, and this compatibility also extends to MongoDB management tools like Compass, Mingo UI, Studio 3T, etc.

However, please note that while FerretDB strives to achieve comprehensive compatibility with Meteor.js apps, there may be certain edge cases or specific MongoDB features that are not yet fully supported.
The FerretDB team is actively addressing these limitations.
To stay up to date with the compatibility status and ongoing efforts, please refer to the [FerretDB documentation](https://docs.ferretdb.io/reference/supported-commands/), or take a look at this [issue where FerretDB aims to achieve compatibility with Meteor.js](https://github.com/FerretDB/FerretDB/issues/2414) examples.

While dealing with these compatibility issues, one of the [issues the FerretDB team has grappled with is the challenge of supporting `OpLog` tailing](https://github.com/meteor/meteor/discussions/12150).
The introduction of `ChangeStreams` in MongoDB 3.6 and later allows real-time data access that bypasses the complications associated with `OpLog` tailing.

Even though Meteor.js traditionally supports MongoDB's `OpLog` tailing for real-time data sync, Meteor.js falls back to polling when `OpLog` is not available; however, while this is not ideal for many production use cases, it can occasionally do the job.
At this stage, it's unclear whether the hassles of implementing `OpLog` will be worth it, [especially with a few Meteor.js apps, like Rocket.Chat, moving to `ChangeStreams`](https://github.com/FerretDB/FerretDB/issues/1993#issuecomment-1518978149).

Some of the other planned feature additions to help improve compatibility with Meteor.js include support for dot notation in query projections, `createIndexes` for unique indexes, support for partial indexes, and many more.
In addition to these feature additions, FerretDB will also be publishing more blog posts that showcase how different applications can use FerretDB in Meteor.js.

## Get Started With FerretDB and Meteor.js

As FerretDB continues to evolve and mature, it is poised to become an even more robust and reliable MongoDB alternative for Meteor.js developers.
The FerretDB team is dedicated to ongoing development, bug fixes, and feature implementation to address the needs and feedback of the Meteor.js community.

Check out the [FerretDB installation guide](https://docs.ferretdb.io/quickstart-guide/) or [contact us to get started](https://docs.ferretdb.io/#community).
