---
title: "FerretDB new version release - 0.8.1"
slug: ferretdb-0-8-1-version-release
author: Alexander Fashakin
description: FerretDB 0.8.1 - the open-source MongoDB alternative - includes new features like version availability, `distinct` command & client TLS validation, and much more.
image: ../static/img/blog/FerretDB_v0.8.1.jpg
tags: [release]
---

![FerretDB version 0.8.1](../static/img/blog/FerretDB_v0.8.1.jpg)

<!--truncate-->

[FerretDB](https://www.ferretdb.io/) 0.8.1 is here, and it's better than ever!
On the back of our Beta release, we've gone even further to improve FerretDB.
And why shouldn’t we, especially when you consider the current economical landscape and the need for truly open source software?

This release and the continued support of our amazing community furthers strengthens our drive to bring you the open source alternative to MongoDB.
And now, as we approach our GA version later this quarter, we’re more emboldened to ensure FerretDB matches up with the behaviour you’d expect from MongoDB, and to create more compatibility with other applications and real world use cases.

Keep reading to learn more.

## New features

We'll start off with probably the most anticipated update!
It is surely a concern having to use an older version of FerretDB, without knowing if a new version release is available or not.

So first up on our list of updates for this release, we are delighted to announce that we now report the availability of newer versions of FerretDB in mongosh.
However, this is only available for users with telemetry enabled.
If telemetry is enabled and a newer version of FerretDB is available, you'll be notified and can stay up-to-date with the latest and best version of FerretDB.

If you are yet to enable telemetry and you’d love to access and enable this feature, please [see our documentation here](https://a9b5c3ea.ferretdb-docs-dev.pages.dev/telemetry/).

Next, we've implemented the `distinct` command in FerretDB.
With this command, you can easily find the unique values of a specific key in your data.
The `distinct` command takes three arguments: `distinct`, `key`, and `query`.
For instance, if you want to find the unique values of the "age" field in a specific "people" collection, run the command below:

```js
db.runCommand( { distinct: "people", key: "age", query: {} } )
```

The `distinct` argument essentially specifies the collection you want to query, the key argument specifies the field you want to find unique values for, and the query argument allows you to filter the results.

Even better, FerretDB now supports the `$rename` field update operator, enabling you to rename fields in a document without changing its contents.
You can use the operator this way:

```js
db.collection.update( { }, { $rename: { "oldField": "newField" } }, { multi: true } )
```

Additionally, we are continuously improving our authentication and security process.
And for that reason, we've also included a way to validate a client's TLS certificate when the root CA certificate is provided.
In essence, this makes it possible to configure FerretDB to validate a client's certificates against the given CA certificate and reject connections without valid certificates.

## Bug fix

With this release, we've also fixed a bug with the `distinct` command where filter wasn't applied to it.

## Documentation

Our documentation is also not left out from this round of improvements.
The biggest change is that addition of the FerretDB blog which was built on Docusaurus– an open source software, and which now resides in our centralized FerretDB repository.

We've added a section for CLI flags and environment variables.
On top of that, we've reformatted our documentation setup to ensure that the deployment URL is visible in logs and can be previewed, and we also added comments and warnings about Git LFS.

This is just the tip of a iceberg.
We have more changes in the works!

To learn about other changes on FerretDB 0.8.1, please read [our release notes](https://github.com/FerretDB/FerretDB/releases/tag/v0.8.1).

As always, we appreciate all our users, supporters, and the entire community that has been a part of the journey all to this moment.
You've all played a role in the growth of FerretDB, and we're excited to continue growing FerretDB because of your unwavering support.

Remember, if you have any questions or feedback, please let us know!
We're always here to help you get the most out of FerretDB.

FerretDB 0.8.1 is here, and it's better than ever!
On the back of our Beta release (please see more here), we've gone even further to make FerretDB better.
This release furthers our aim to bring you the truly open source alternative to MongoDB by using a proxy to convert MongoDB wire protocols to SQL, with the backend on PostgreSQL.
