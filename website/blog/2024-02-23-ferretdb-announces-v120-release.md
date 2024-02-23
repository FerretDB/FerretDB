---
slug: ferretdb-announces-v120-release
title: FerretDB announces v1.20 release
authors: [alex]
description: >
  We have released FerretDB v1.20, and it includes a couple of changes that will be crucial to enable and build further increased compatibility, performance, and security for FerretDB.
image: /img/blog/ferretdb-v1.20.0.jpg
tags: [release]
---

![FerretDB announces v1.20](/img/blog/ferretdb-v1.20.jpg)

We have released FerretDB v1.20, and there are a couple of changes that will be crucial to enable and build further increased compatibility, performance, and security for [FerretDB](https://www.ferretdb.com/).

<!--truncate-->

There are some interesting changes in this release that form a critical part of upcoming features, such as SCRAM-SHA-256 and SCRAM-SHA-1 authentication mechanisms, support for [Compass UI](https://www.mongodb.com/products/tools/compass), and our plans for FerretDB v2.0.

In this release post, we will delve into our ongoing efforts, and how some of the recent changes play a pivotal role for future releases.

## Upcoming features

Many of the latest changes in this release – and the last couple of releases as well – serve as a base ground to enabling FerretDB for even more use cases.
For instance, we are working on enabling more user management commands and should soon have SCRAM-SHA-256 authentication support.
Work on adding support for SCRAM-SHA-1 is also in progress and both authentication mechanisms could be ready by the next release.

They are currently available for experimental purposes and can be enabled using the `--test-enable-new-auth` flag or`FERRETDB_TEST_ENABLE_NEW_AUTH` environment variable.

At the moment, you can experiment with any of the supported user management commands.
A user is created using the `createUser` command and its details are stored in the `system.users` collection of the`admin` database.
Then FerretDB will use the provided username and password to authenticate against the stored credentials of the user.

In addition, we have been working on enabling support for Compass UI; it should be available in the next release.
This is an exciting development as it will make FerretDB available to more users who have been asking about Compass support for quite some time.
We have also been investigating sessions and can't wait to have it available.

Authorization is another noteworthy feature that's in our immediate roadmap.
We are currently researching ways to make authorization work on a FerretDB level.

## The journey to FerretDB v.2.0

These updates are just a glimpse of what's to come.
The latest release is going to be one of the last few releases of FerretDB v1.x.

Since we started working on FerretDB, we've had an immense response from the open source community and people looking to move away from the vendor-lock-in restrictions caused by MongoDB's license change.
No one loves having a rug pulled from underneath them, and that's what many users felt and echoed when MongoDB switched to a more proprietary license.

We understand the need to make it easy for users to switch by offering a truly open source alternative that is compatible with many of their existing needs.

FerretDB v2.0 is just around the corner, and it will probably be released later this year.
One of the motivations for v2.0 is that certain use cases require even higher performance and compatibility.
So we believe FerretDB v1.20 should perform to these standards so more people can use it as their MongoDB alternative.

The open source community has always been a big part of everything we do at FerretDB.
There's been several impactful contributions from the community and we would love to appreciate @[ahmethakanbesel](https://github.com/ahmethakanbesel) as a first-time contributor to FerretDB in this release.

Check out the [release notes for the complete list](https://github.com/FerretDB/FerretDB/releases/tag/v1.20.1) of changes.
If you have any question at all about FerretDB, please feel free to [reach out on any of our channels here](https://docs.ferretdb.io/#community).
