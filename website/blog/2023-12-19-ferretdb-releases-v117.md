---
slug: ferretdb-releases-v117
title: FerretDB releases v1.17.0!
authors: [alex]
description: >
  We just released FerretDB v1.17.0, which now makes it possible to build FerretDB without PostgreSQL or SQLite backend, among other new features.
image: /img/blog/ferretdb-v1.17.0.jpg
tags: [release]
---

![FerretDB releases v.1.17.0](/img/blog/ferretdb-v1.17.0.jpg)

We just released FerretDB v1.17.0, which now makes it possible to build FerretDB without PostgreSQL or SQLite backend, among other new features.

<!--truncate-->

This release rounds up what has been an impressive year at [FerretDB](https://www.ferretdb.com/).
Our mission is clear – to bring MongoDB workloads back to open source.

We've had numerous new features, the addition of the SQLite backend, and improved architecture to support more backends, including MySQL and SAP Hana.
Throughut the year, we made significant improvements to FerretDB performance, enabled better compatibility for many applications and use cases, and made FerretDB available on managed cloud providers.
The FerretDB documentation has also undergone extensive improvements, and with the new release, we've now enabled versioning.

Our contributor list also keeps expanding as we added 3 new contributors ([@wazir-ahmed](https://github.com/wazir-ahmed), [@anunayasri](https://github.com/anunayasri), and @[hungaikev](https://github.com/hungaikev)), our thanks to you all.

Now, let's see what's new in this release.

## New features

Building FerretDB without PostgreSQL and/or SQLite backend is now possible by setting the `ferretdb_no_postgresql` and `ferretdb_no_sqlite` build tags.
This way, you can build FerretDB without any backend.

We've also added support for sorting by `$natural`; you can now set `{$natural: 1}` and `{$natural: -1}` by RecordID for capped collections.
In addition to this change, we've also enabled collection UUID generation, and this is now visible in the `listCollections` output.

## Bug fixes and enhacements

This release fixes `listDatabases` filtering error when `nameOnly` parameter is passed.
We've also improved `validate` diagnostic command, and added more fields to `listCollections.cursor` response (`options.capped`, `options.size`, `options.max`, `info.readOnly`, and `idIndex`).

## We wish you a happy holiday season

Find the complete list of changes for FerretDB v1.17.0 [here](https://github.com/FerretDB/FerretDB/releases/tag/v1.17.0)

Thank you to everyone who contributed to this release, the open source community, and to all our users for your continued support.

We wish you all a happy holiday season and a prosperous new year!

If you have questions or feedback on FerretDB, [contact us on our community channels](https://docs.ferretdb.io/#community).
