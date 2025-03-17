---
slug: new-ferretdb-v2-rc2-release-with-debian-packages-and-better-indexing-support
title: FerretDB releases v2.0.0-rc.2 with new Debian packages and better indexing support
authors: [alex]
description: >
  We have released FerretDB v2.0.0-rc.2, with an embeddable Go package, new Debian packages, improved stats and logging, and better indexing support.
image: /img/blog/ferretdb-v2-rc2.jpg
tags: [release]
---

![FerretDB v2.0.0-rc.2](/img/blog/ferretdb-v2-rc2.jpg)

FerretDB v2 keeps getting better!
Based on your feedback, we are making improvements with an embeddable Go package, new Debian packages,
improved logging, and better indexing support.

<!--truncate-->

After we released [the first release candidate of FerretDB v2](https://blog.ferretdb.io/ferretdb-releases-v2-faster-more-compatible-mongodb-alternative/),
we received lots of amazing feedback and shoutouts from the community.
So many of our users have been impressed with its overall performance improvements and range of compatible features,
including vector search, full-text search, replication, and better aggregation pipeline support, among others.

We've just released [FerretDB v2.0.0-rc.2](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0-rc.2),
with new features and improvements that make it even better,
and we're excited to share them with you!

## Docker image tags update

We've made some changes to our Docker image tags.
The `latest` tag now points to v2, which means if you're pulling directly without specifying a tag
(e.g., `ghcr.io/ferretdb/ferretdb`), you're getting FerretDB v2 by default.
If you're running FerretDB in production, we strongly recommend explicitly specifying the full version tag
(e.g., `ghcr.io/ferretdb/ferretdb:2.0.0-rc.2`) to ensure consistency across deployments.

## Embeddable Go package

FerretDB is now even easier to integrate into other applications with the new
[embeddable Go package](https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/ferretdb).
This means you can run FerretDB within your own Go applications, giving you more flexibility in deployment and usage.

## Debian packages for DocumentDB

We now provide `.deb` packages for users who want to install the DocumentDB PostgreSQL extension
on Debian-based systems.
This makes installation and upgrades simpler for Debian and Ubuntu users.
You can now install it from the [DocumentDB repository](https://github.com/FerretDB/documentdb/releases).

## TTL indexes and `reIndex` command

Indexes are essential parts of any database, and based on your feedback, we have fixed TTL indexes and
implemented the `reIndex` command.

In the previous release, there was an issue with TTL indexes where expired documents were not removed.
With the new release, TTL indexes now function correctly; expired documents will now automatically be removed as expected.

```js
db.runCommand({
  createIndexes: 'collection',
  indexes: [{ key: { createdAt: 1 }, expireAfterSeconds: 3600 }]
})
```

Some users reported issues with indexes not working as expected, not only the TTL indexes.
They've now been fixed, and we recommend rebuilding all indexes with the `reIndex` command to ensure they are in the best state.
You can do that with the following command:

```js
db.runCommand({ reIndex: 'collection' })
```

## New `dbStats` command

`dbStats` is a crucial command for many UI applications and monitoring tools.
It provides important information about database size, storage efficiency, and more.
With the new release, you can now get better insights into your database with the `dbStats` command.

```js
db.runCommand({ dbStats: 1 })
```

## Mongo-compatible logging format

We have added a MongoDB-style log format, making it easier for users and tools to parse and analyze them.

## Thank you!

With the new FerretDB v2 release candidate, we are enabling more applications to run their MongoDB workloads
on open source, using all the familiar features and commands.
We are grateful for all the feedback, suggestions, and contributions from the community,
and we look forward to hearing more from you as we continue to improve FerretDB.

If you're still on v1, you're missing out on all these new features and improvements.
FerretDB v2 is the most feature-complete open source alternative to MongoDB – see our
[releases page](https://github.com/FerretDB/FerretDB/releases) to try it out!

If you have any questions about FerretDB, please feel free to reach out on any of our channels listed
[here](https://docs.ferretdb.io/#community).
