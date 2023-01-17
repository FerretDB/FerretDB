---
slug: new-ferretdb-minor-release-0-5-1
title: "New FerretDB minor release - 0.5.1"
author: Peter Farkas
image: ../static/img/blog/rescue-kitten-komari-ferret-brothers-47.jpg
tags: [release]
date: 2022-07-27
---

![photo credit: boredpanda.com](../static/img/blog/rescue-kitten-komari-ferret-brothers-47.jpg)

<!--truncate-->

Today we released FerretDB 0.5.1, a minor release which adds some new features, but mostly improvements and fixes.

As you may have noticed, our release schedule is every two weeks.
A major version 0.y.z every month,, followed by a minor or patch release two weeks later.
Our plan is to release 1.0.0 by the end of the year, which would be the first FerretDB release recommended to be used as a replacement for MongoDB.
You can [check out our roadmap on GitHub](http://www.github.com/orgs/FerretDB/projects/2).

In this month’s patch release, we added features, some of the notable ones are:

**Array Query Operators, now all 3 of them!**

By adding support for the `$all` Array Query operator, which matches arrays containing all the elements specified in the query, FerretDB now supports all Array Query operators: `$all`, `$size`, and `$elemMatch`.

Note that there is a known limitation where $elemMatch may use other operators recursively, however, not all of those operators are supported in FerretDB as of now.

## Diagnostic commands

With 0.5.1, we added support for the getLog and explain commands, adding these to the long list of supported MongoDB diagnostic commands, [see the full list here](https://github.com/FerretDB/FerretDB/issues/228).

Please check out the Release notes for the full list of added features.

### Support count for Tigris handler

We added support of count method for Tigris handler.
If you run FerretDB with Tigris, as of now you can use count.

## Fixes

In this release, we provide fixes such as implementing changes in the handling of non-existing databases and collections to better align with MongoDB.
We also fixed the behavior of ModifyCount in certain edge cases.

## Other changes

We enhanced our testing and development environment with creating and improving tests.

As always, please share your feedback with us!
