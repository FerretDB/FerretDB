---
slug: ferretdb-v-1-5-0-beta-support-sqlite
title: FerretDB v1.5.0 New Release with Beta-level Support for the SQLite Backend
authors: [alex]
description: >
  We are delighted to announce the release of FerretDB v1.5.0 which includes beta-level support for the SQLite backend.
image: /img/blog/ferretdb-v1.5.0.jpg
tags: [release]
---

![FerretDB v.1.5.0 includes beta-level support for the SQLite Backend](/img/blog/ferretdb-v1.5.0.jpg)

After a great deal of work from the [FerretDB](https://www.ferretdb.io/) team and our open-source community, we are delighted to announce the release of FerretDB v1.5.0, which includes beta-level support for the [SQLite](https://www.sqlite.org/) backend.

<!--truncate-->

Important note to [Tigris](https://www.tigrisdata.com/) users: FerretDB 1.5.0 is the last release of FerretDB which includes support for the Tigris backend.
If you wish to use Tigris, please do not update FerretDB beyond v1.5.0.
Earlier versions of FerretDB with Tigris support will still be available on [GitHub](https://github.com/FerretDB/FerretDB/releases).

In the past couple of weeks, we've been working on providing support for the SQLite backend.
This is very much a work in progress and we'll be adding more functionalities in our next releases, but we are excited for FerretDB enthusiasts to try it out – we can't wait to see what you achieve with it!

Let's check out the rest!

## New features and enhancements

This release has several exciting updates, especially the provision of beta support for the SQLite backend, which includes enabled support for cursors and the `count` command.

With improved support for cursors for SQLite (and PostgreSQL) backend in the latest version release, FerretDB users can now use commands like `find` and `aggregate` to return large data sets more efficiently.

We've also implemented the `count` command for the SQLite backend.

For other important updates in this release, please see the [release notes](https://github.com/FerretDB/FerretDB/releases/tag/v1.5.0)

## We Thank You

We've received lots of amazing feedback from the community in the past couple of weeks and we're just happy to thank everyone for their contribution, either through code additions, bug reports, blog posts, and product feedback.
We also have a first contribution from [@Matthieu68857](https://github.com/Matthieu68857).
Thank you!

Going forward, we will be adding more features to enable full support for the SQLite backend.
This is very exciting!

We can't wait for you to try out FerretDB with the SQLite backend.
And when you do, please share with us all your discoveries, feedback, questions, bugs, anything.
Every little bit counts!

You can reach out to us on any of [our community channels](https://docs.ferretdb.io/#community), or [find us on GitHub](https://github.com/FerretDB/FerretDB/).
