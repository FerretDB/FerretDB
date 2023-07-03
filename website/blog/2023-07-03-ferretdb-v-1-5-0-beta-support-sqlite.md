---
slug: ferretdb-v-1-5-0-beta-support-sqlite
title: FerretDB v1.5.0 New Release with Beta-level Support for SQLite
authors: [alex]
description: >
  We are delighted to announce the release of FerretDB v1.5.0 which includes beta-level support for SQLite backend.
image: /img/blog/ferretdb-v1.5.0.jpg
tags: [release]
---

![FerretDB v.1.5.0 includes beta-level support for SQLite](/img/blog/ferretdb-v1.5.0.jpg)

After a great deal of work from the [FerretDB](https://www.ferretdb.io/) team and our open-source community, we are delighted to announce the release of FerretDB v1.5.0, which includes beta-level support for [SQLite](https://www.sqlite.org/) backend.

<!--truncate-->

For [Tigris](https://www.tigrisdata.com/) backend users, please note that this will be the last release with FerretDB support for the Tigris backend.
Starting from FerretDB v1.6.0, we will no longer provide support for Tigris.
But if you wish to continue using the Tigris backend, please refrain from updating FerretDB beyond the current version, v1.5.0.
Other earlier versions of FerretDB will still offer Tigris backend support, and they are all available on [GitHub](https://github.com/FerretDB/FerretDB/releases).

In the past couple of weeks, we've been working on providing support for SQLite backend.
And while there are still some missing functionalities, we believe it is ready for early adopters who are interested in testing FerretDB with SQLite, and we can't wait to find out what everyone comes up with.

Let's check out the rest!

## New features and enhancements

This release has several exciting updates, especially the provision of beta-like support for SQLite, which includes enabled support for cursors and the `count` command.

With the improved support for cursors in SQLite (and PostgreSQL) in the latest version update, FerretDB users can now use commands like `find` and `aggregate` to return large data sets more efficiently.

We've also implemented the `count` command in SQLite.

For other important updates in this release, please see the [release notes](https://github.com/FerretDB/FerretDB/releases/tag/v1.5.0)

## We Thank You

We've received lots of amazing feedback from the community in the past couple of weeks and we're just happy to thank everyone for their contribution, either through code additions, bug reports, blog posts, and product feedback.
We also have a first contribution from [@Matthieu68857](https://github.com/Matthieu68857).
Thank you!

Going forward, we will be adding more features to enable full support for the SQLite backend.
This is very exciting!

We can't wait for you to try out FerretDB with an SQLite backend.
And when you do, please share with us all your discoveries, feedback, questions, bugs, anything.
Every little bit counts!

You can reach out to us on any of [our community channels](https://docs.ferretdb.io/#community), or [find us on GitHub](https://github.com/FerretDB/FerretDB/).
