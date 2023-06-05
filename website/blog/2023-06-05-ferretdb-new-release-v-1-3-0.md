---
slug: ferretdb-new-release-v-1-3-0
title: FerretDB v1.3.0. Release
authors: [alex]
description: >
  We’ve just announced the release of a new version of FerretDB v1.3.0, which now includes new feature additions, such as the `logout` command and positional operator in projection.
image: /img/blog/ferretdb-v1.3.0.jpg
tags: [release]
---

![FerretDB v.1.3.0](/img/blog/ferretdb-v1.3.0.jpg)

We've just announced the release of FerretDB v1.3.0, which includes new feature additions, such as the `logout` command and positional operator in projection.

<!--truncate-->

Since our [GA release](https://blog.ferretdb.io/ferretdb-1-0-ga-opensource-mongodb-alternative/), we've worked on improving all aspects of [FerretDB](https://www.ferretdb.io/) with the help of our active and growing community; you've all been truly amazing!

We've seen a high demand for FerretDB and its application as an open-source replacement for MongoDB workloads and applications.
This is exciting news, and we look forward to enabling more developers build their applications with FerretDB, starting with the latest release.

Want to know what's new in FerretDB v.1.3.0?
Let's find out.

## New features

In this release, we've added the `logout` command.
Even though it's deprecated in MongoDB, it's necessary in some circumstances.
This feature addition was made possible thanks to one of our contributors, @[@kropidlowsky](https://github.com/kropidlowsky).

Similarly, we've been improving our projection features and just added a positional operator which should help users retrieve the first matching element from an array if the specified query conditions are met.

## Fixed bugs

We've also improved on some important issues with the previous release, especially with regard to version update notifications and error handling for authentication issues.

Previously, when an update was no longer available, FerretDB would still report an update being available.

In this release, we've fixed this issue, and now FerretDB should accurately reflect the status of available updates correctly via [telemetry](https://docs.ferretdb.io/telemetry/).

Also, we noticed that error messages related to authentication were quite unclear.
This release addresses this by returning more insightful error messages and also provides updates on our documentation about authentication and TLS connections.

We also had an issue with setting version fields for `.deb` and `.rpm` packages for FerretDB.
This issue has been resolved; both packages should now accurately reflect the correct versions.

Beyond this, we've also fixed an issue found when using `distinct` commands with a `null` query which resulted in a type mismatch.
Also, in cases where multiple `update` operators were applied to the same path, there were path collisions that would arise and prevent successful update operations.
That has been fixed, and update operations should work correctly now without conflicts.

## Documentation

To improve the ease of use of FerretDB, we've expanded our documentation to include detailed explanations of FerretDB authentication and TLS connections ([See documentation for more](https://docs.ferretdb.io/category/security/)) and basic explanations of aggregation operations.

In addition to that, we've added a Markdown formatter which should help format tables and ensure consistency in our documentation.

## Other changes

Please see [our release notes](https://github.com/FerretDB/FerretDB/releases/tag/v1.3.0) for the complete list of changes and keep up to date with us as we work on enabling more MongoDB workloads to take advantage of FerretDB.

Ensure to upgrade to this version of FerretDB to take advantage of the new features and bug fixes.
We thank all our contributors, partners, users, and the entire open-source community; your contributions have been invaluable!
We look forward to more contributions and feedback from you.

Remember, if you have any questions at all, we'd love to hear them.
[Contact us here!](https://docs.ferretdb.io/#community)
