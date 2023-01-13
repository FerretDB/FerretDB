---
slug: new-release-ferretdb-0-4-0
title: "New release: FerretDB 0.4.0!"
author: Peter Farkas
image: ../static/img/blog/4ferrets.png
tags: [release]
date: 2022-06-28
---

![New FerretDB release 0.4.0](../static/img/blog/4ferrets.png)

<!--truncate-->

We are happy to announce that [FerretDB’s newest 0.4.0 release is now available on GitHub,](https://github.com/FerretDB/FerretDB/releases/tag/v0.4.0) with some exciting new features and fixes in it.
We wish to thank everyone in the community who contributed to this release, either with code or feedback.
Special thanks to [ribaraka](https://github.com/ribaraka) and [fenogentov](https://github.com/fenogentov) for their sustained efforts.

Since 0.3.0, we are getting more and more user reports from those early adopters, who were curious enough to go ahead and replace their MongoDB instances with FerretDB - with success!
The feedback we are getting is invaluable, though your mileage may vary based on the complexity of your application’s  database needs.

As mentioned before, we are already compatible with applications, like SAP’s [CLA assistant](https://github.com/cla-assistant), and there will be many more to come in the next couple of months.

We anticipate that with 0.4.0, we are getting several steps closer to providing an open source MongoDB database experience for those who care about avoiding vendor lock-in.

If you decide to try FerretDB, please [contact us](http://www.ferretdb.io/contact) and let us know how it went!
We thrive on the feedback of the community and early adopters, and we are happy to re-prioritize a specific feature, if feasible.

## New features in 0.4.0

With this release, we are adding support for specific field operators such as `$setOnInsert`, `$unset` and `$currentDate`.

We are also adding support for array querying and the `$elemMatch` array query operator.

Stubs were also added for features related to MongoDB’s Free Monitoring in preparation to implement such a feature in the future.
This is a feature which exists since MongoDB 4.0, and provides support for getting data from your running instances, such as resource utilization and execution times.

## Added support for Tigris

While our primary focus is still PostgreSQL, our newest release marks the first time where we include functionality related to [Tigris](http://www.tigrisdata.com).
Tigris is one of the database backends we are going to support next to PostgreSQL and potential others in the future.

![support for tigris](../static/img/blog/cf73bb31-fa1b-4465-8277-e73da46127de-1650484034538-1-300x120.png)Tigris is a work in progress, fully open source database as a service platform, which aims to add MongoDB compatibility by using FerretDB as an interface.
[Check them out on GitHub](https://github.com/tigrisdata).

We see enormous potential in our partnership in terms of value: together, FerretDB and Tigris will be able to provide an alternative to MongoDB Atlas.
We think this is huge, especially knowing that the people behind Tigris are among the most seasoned infrastructure and database engineers in the industry.

While we added functionality which supports Tigris, this is very much preliminary, and instructions on how to take it for a test drive will be provided later.

Please see the [full changelog on GitHub](https://github.com/FerretDB/FerretDB/releases/tag/v0.4.0)!
