---
slug: 2023-02-14-ferretdb-v-0-9-1-minor-release
title: "FerretDB v0.9.1 - Minor Release"
authors: [alex]
description: "FerretDB v0.9.1 is a minor release that contains upgrades on previous versions, including bug fixes, enhancements, and new features"
image: /img/blog/ferretdb-v0.9.1.jpg
tags: [release]
date: 2023-02-14
---

FerretDB v0.9.1 is a minor release with upgrades on previous versions, including bug fixes, enhancements, and new features.

![FerretDB v0.9.1 - Minor Release](/img/blog/ferretdb-v0.9.1.jpg)

<!--truncate-->

We are pleased to announce the release of FerretDB v0.9.1, a minor release that builds on previous versions and implements new features and bug fixes based on some of the feedback from the community.

Our goal at [FerretDB](https://www.ferretdb.io) is to provide an [open-source MongoDB alternative](https://blog.ferretdb.io/5-database-alternatives-mongodb-2023/) that’s compatible with real-world applications.
We release a new version of FerretDB every two weeks, so be sure to stay up to date with the latest version.

We express our sincere gratitude to our community for their continuous support, contributions, and valuable feedback.
In recent weeks, we’ve seen a marked increase in interest and feedback on FerretDB, which is exciting and will further drive our growth.

This blog post only covers a few changes in this release.
To learn even more, please check here for a [complete list of the changes](https://github.com/FerretDB/FerretDB/releases/tag/v0.9.1) in FerretDB v0.9.1.

## New Features

Similar to previous versions, this release also pushes more filtering queries to the backend, with support for Tigris pushdowns for numbers and queries with dot notation, which should significantly speed up query processes.
We are committed to improving the performance of FerretDB, and this feature is a significant step towards that goal.

## Fixed Bugs

In this release, we’ve identified and addressed some of the bugs in previous releases.
For instance, we’ve fixed the SASL response for PLAIN authentication, which now includes additional fields in the response document to support the PLAIN mechanism's two-step conversation.

We addressed the issue of key ordering during document replacement in this release by fixing the behavior of upsert and non-upsert updates that did not have any operator specified.
Prior to this release, FerretDB sorted data during these operations, which wasn’t the intended behavior.

Additionally, we fixed the `$pop` operator error handling for non-existent path.

All the changes in this release are part of our effort in building the ultimate open-source alternative to MongoDB.
Please see the [release notes on FerretDB v0.9.1](https://github.com/FerretDB/FerretDB/releases/tag/v0.9.1) for a detailed list of all the changes in this version.

We're always here to help you get the most out of FerretDB.
As a result, we encourage our users to share feedback, ask questions, and leave comments.
Get in touch with us [here](https://docs.ferretdb.io/#community)!
