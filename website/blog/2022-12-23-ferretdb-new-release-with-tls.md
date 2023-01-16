---
slug: ferretdb-new-release-with-tls
title: "FerretDB v0.7.1 - Now offering basic TLS support"
author: Alexander Fashakin
date: 2022-12-23
---

![FerretDB 0.7.1 release](https://i0.wp.com/www.swhosting.com/blog/wp-content/uploads/2014/10/TLS.png)

<!--truncate-->

We just released FerretDB 0.7.1, and this release - and the previous release (v0.7.0)- includes several exciting new features, updates, and bug fixes lined up for you.
Here's what's new:

## New features

The biggest change in this new version is the implementation of basic TLS support for FerretDB.
And with this addition, our users can now configure FerretDB to listen on a TLS port.
Starting from v0.7.0, we've included support for `filter` in `listCollections`, meaning you can now use a filter to limit the results returned in a `listCollections` set.
For Tigris users, we've implemented the `explain` command.

## Bug fixes

In the latest release, we’ve also addressed issues with `unset` comparisons, ensuring it works in a similar way as MongoDB.
This is in addition to previous bug fixes from the last release.
For example, the issue with using the greater than and less than operators on array values comparison has been resolved.
Furthermore, we've fixed the issue with parallel inserts causing concurrent database creation to fail when the database doesn't exist.

## Documentation

Our documentation has undergone a few improvements as well.
We've added more details to the supported commands table.
Users can now view the list of available commands for diagnostics, authentication and role management, and session and free monitoring operations.

## Other enhancements and changes

Among other improvements, starting from v.0.7.0, we now allow the use of dash `(-)` in database names for both PostgreSQL and Tigris.
We've also simplified and improved the approach to fetch documents for `delete`, `count`, `find`, `findAndModify`, and `update`.

You can find more details on all these improvements and more in the [FerretDB](https://github.com/FerretDB/FerretDB/blob/main/CHANGELOG.md "") changelog.
If you have any questions, feel free to [contact us](https://docs.ferretdb.io/intro/#community).
