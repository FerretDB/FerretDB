---
slug: new-release-ferretdb-0_6_1
title: "New release: FerretDB 0.6.1!"
author: Alexander Fashakin
image: ../static/img/blog/six_ferrets-1024x917.jpg
date: 2022-11-11
---

![Six ferrets](../static/img/blog/six_ferrets-1024x917.jpg)

<!--truncate-->

We are happy to inform you that FerretDB is currently in Alpha, and we have a new release - FerretDB 0.6.1.

With the Alpha version, FerretDB has many notable and exciting features, bug fixes, and enhancements.
This couldn't have been possible without the support, feedback, and help of many contributors from the community.
Special thanks go to [@ronaudinho](https://github.com/ronaudinho), [@codingmickey](https://github.com/codingmickey), [@ndkhangvl](https://github.com/ndkhangvl), and [@zhiburt](https://github.com/zhiburt)  for their remarkable contributions.

Starting from 0.6.0, we have added a basic telemetry service that collects anonymous usage data on FerretDB instances.
This will help us quickly identify failed queries, prioritize issues, and achieve greater parity with MongoDB.
And for users that do not want telemetry, we have provided several options to disable in [our documentation](https://docs.ferretdb.io/).

## New features

In the latest version, we now support `$max` field update operators.
We've also enabled simple query pushdown for queries that look like `{_id: _ObjectID()}` in PostgreSQL.
Aside from that, we've migrated FerretDB to Kong, which allows us to provide both environment variables and flags for configuration.

## Fixes

The new version of FerretDB now supports empty document field names, while we've also fixed error messages for invalid `$and/$or/$nor` arguments.

## Documentation

Our [documentation page](https://docs.ferretdb.io/) is up and running!
We've also added a local search plugin, set up pages on installation procedures, basic CRUD operations, and much more!

Please see our [Changelog](https://github.com/FerretDB/FerretDB/releases/) for more details on the new releases.

(image credit: FetzlePetzit / [furaffinity.net](https://www.furaffinity.net/view/1920045/))
