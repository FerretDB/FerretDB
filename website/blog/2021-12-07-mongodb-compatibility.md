---
slug: mongodb-compatibility-whats-really-important
title: "MongoDB Compatibility - What’s Really Important ?"
author: Peter Farkas
tags: [Ferretdb]
---

In this recent [TheRegister Article](https://www.theregister.com/2021/12/06/aws_documentdb_not_mongodb_compatible/ "https://www.theregister.com/2021/12/06/aws_documentdb_not_mongodb_compatible/"), there is an interview with MongoDB's CTO, Mark Porter, where the level of compatibility between MongoDB and DocumentDB is discussed.

<!--truncate-->

In the interview, Mark claims that Amazon’s DocumentDB - one of the leading MongoDB alternatives - is just 34% compatible with MongoDB itself.

This looks damning, until you remember the 80/20 rule, which applies here as “80% of applications require just 20% of functionality”.
And if one thing you can count on from AWS is that they are listening to their customers and [choosing functionality](https://docs.aws.amazon.com/documentdb/latest/developerguide/release-notes.html "https://docs.aws.amazon.com/documentdb/latest/developerguide/release-notes.html") to implement wisely.

What MongoDB did really well is - they nailed the developer experience with “native” experience for a variety of programming languages and frameworks, allowing you to seamlessly get persistence for your documents with MongoDB.
This interface and underlying protocol is so good that it is likely to become a standard, similar to how SQL became decades ago or as PostgreSQL wire protocol (used by Yugabyte, CochroachDB, ClickHouse, Spanner, to name a few) became more recently.

As we can see with SQL, few databases have 100% compatibility with most recent standard versions, and databases such as MySQL or Clickhouse reached great adoption despite their early versions having very poor SQL compatibility.
They just provided the right features to be a great fit for certain applications.

We expect that this is how the MongoDB Alternatives Market will evolve -  [DocumentDB](https://aws.amazon.com/documentdb/ "https://aws.amazon.com/documentdb/"), [CosmosDB](https://docs.microsoft.com/en-us/azure/cosmos-db/mongodb/mongodb-introduction "https://docs.microsoft.com/en-us/azure/cosmos-db/mongodb/mongodb-introduction"), [Oracle](https://blogs.oracle.com/database/post/introducing-oracle-database-api-for-mongodb "https://blogs.oracle.com/database/post/introducing-oracle-database-api-for-mongodb"), FerretDB (ourselves) will focus on implementing features which actually matter, not merely trying to achieve some higher compatibility number on arbitrary MongoDB test suites.
Chances are, the companies will come together to standardize and implement useful extensions to MongoDB interfaces, which MongoDB itself is not pursuing.

So do not get scared by [MongoDB’s Amazon DocumentDB bashing web site](https://www.isdocumentdbreallymongodb.com/ "https://www.isdocumentdbreallymongodb.com/")  and self-serving compatibility [“evaluation”](https://www.mongodb.com/atlas-vs-amazon-documentdb/compatibility "https://www.mongodb.com/atlas-vs-amazon-documentdb/compatibility").
Look into the functionality your application uses and plans to use and see whenever any of the MongoDB Alternatives support it.
If they do not - let them know what is important to you .
With FerretDB as an Open Source Project, you can both [file an issue](https://github.com/FerretDB/FerretDB/issues "https://github.com/FerretDB/FerretDB/issues") or scratch your own itch by contributing the feature you need.

Even if you stick to using MongoDB Atlas at this point, do make sure that there are possible Alternative Backends for your application to ensure you are not a hostage and have choices.

At FerretDB, we are in the business of making sure that there is a 100% Open Source choice for your MongoDB alternative needs.
