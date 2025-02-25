---
slug: new-ferretdb-v2-rc2-release-with-debian-packages-and-more-indexing-support
title: FerretDB releases v2.0.0-rc.2 with new Debian packages and more indexing support
authors: [alex]
description: >
  We have released FerretDB v2.0.0-rc.2, with an embeddable package, new Debian packages, improved stats and logging, and more indexing support.
image: /img/blog/ferretdb-v2-rc2.jpg
tags: [release]
---

![FerretDB v2.0.0-rc.2](/img/blog/ferretdb-v2-rc2.jpg)

[FerretDB](https://www.ferretdb.com/) v2 keeps getting better!
Based on your feedback, we are making improvements with an embeddable Go package, new Debian packages, improved logging, and better indexing support.

<!--truncate-->

After [we released FerretDB v2](https://blog.ferretdb.io/ferretdb-releases-v2-faster-more-compatible-mongodb-alternative/), we received lots of amazing feedbacks and shoutouts from the community.
So many of our users have been impressed with its overall performance improvements and range of compatible features, including vector search, full-text search, replication, and better aggregation pipeline support, among others.

We've just released FerretDB v2.0.0-rc.2, with new features and improvements that make it even better, and we're excited to share them with you!

## Docker image tags update

We've made some changes to our Docker image tags.
The `latest` tag now points to v2, which means if you're pulling directly without specifying a tag (e.g. `ghcr.io/ferretdb/ferretdb`), you're getting FerretDB v2 by default.
If you're running FerretDB in production, we strongly recommend explicitly specifying the full version tag (e.g. `ghcr.io/ferretdb/ferretdb:2.0.0-rc.2`) to ensure consistency across deployments.

## Embeddable package

FerretDB is now even easier to integrate into other applications with the new embeddable package.
This means you can run FerretDB within your own Go applications, giving you more flexibility in deployment and usage.

## Debian packages for DocumentDB

We now provide `.deb` packages for users who want to install FerretDB with the DocumentDB PostgreSQL extension on Debian-based systems.
This makes installation and upgrades simpler for Debian and Ubuntu users.
You can now install it from our repository; you can also check out the installation guide in our documentation.

## Support for TTL index and `reIndex`

Indexes are essential parts of any database, and based on your feedback, we now have full support for TTL indexes and the `reIndex` command.
You can try it out as shown below:

TTL indexes now function correctly; Expired documents will now automatically be removed as expected.

```sh
db.collection.createIndex( { "createdAt": 1 }, { expireAfterSeconds: 3600 } )
```

Similarly, the `reIndex` command has been fixed, so you can rebuild indexes without issues.

```sh
db.collection.reIndex()
```

## New `dbStats` command

`dbStats` is a crucial command for many UI applications and monitoring tools.
It provides important information about database size, storage efficiency, and more.
With the new release, you can now get better insights into your database with the `dbStats` command.

```sh
db.stats()
```

## Mongo-compatible logging format

Logs now follow the MongoDB-style format, making it easier for users and tools to parse and analyze them.

## Thank you!

With the release of FerretDB v2, we are enabling more applications to run their MongoDB workloads in open source, using all the familiar features and commands.
We are grateful for all the feedback, suggestions, and contributions from the community, and we look forward to hearing more from you as we continue to improve FerretDB.

If you're still on v1, you're missing out on all these new features and improvements.
FerretDB v2 is the most feature-complete open source alternative to MongoDB – [try it out here](https://github.com/FerretDB/FerretDB/releases)!

If you have any questions about FerretDB, please feel free to [reach out on any of our channels here](https://docs.ferretdb.io/#community).
