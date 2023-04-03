---
slug: ferretdb-0-8-1-version-release
title: "FerretDB new version release - 0.8.1"
author: Alexander Fashakin
description: FerretDB 0.8.1 - the open-source MongoDB alternative - includes new features like version availability, `distinct` command & client TLS validation, and much more.
image: /img/blog/FerretDB-v0.8.1.jpg
tags: [release]
date: 2023-01-19
---

FerretDB 0.8.1 - the open-source MongoDB alternative - includes new features like version availability, `distinct` command & client TLS validation, and much more.

![FerretDB version 0.8.1](/img/blog/FerretDB-v0.8.1.jpg)

<!--truncate-->

[FerretDB](https://www.ferretdb.io/) 0.8.1 is here, and it's better than ever!
On the back of our Beta release, we've gone even further to improve FerretDB with lots of new features and enhancements.
This FerretDB release and the continued support of our amazing community further strengthens our drive to bring you the open source alternative to MongoDB.

Keep reading to learn more.

## What's new

We'll start off with probably the most anticipated update!
We are delighted to announce that we now report the availability of newer versions of FerretDB in mongosh.
However, this is only available for users with telemetry enabled.
If telemetry is enabled and a newer version of FerretDB is available, you'll be notified and can stay up-to-date with the latest and best version of FerretDB.

If you are yet to enable telemetry and you’d love to access this feature, please [see our documentation here](https://docs.ferretdb.io/telemetry/).

Next, we've implemented the `distinct` command in FerretDB.
With this command, you can easily find the unique values of specific fields in your data.
The `distinct` command takes three arguments: `distinct`, `key`, and `query`.
See usage below:

```js
db.collection.distinct(
    <key>,
    {
        <query>
    }
)
```

Even better, FerretDB now supports the `$rename` field update operator, enabling you to rename fields in a document without changing its contents.
You can use the operator this way:

```js
db.collection.update(
    { },
    {
        $rename: {
            <oldField>: <newField>
        }
    },
    {
        multi: true
    }
)
```

Additionally, we are continuously improving our authentication and security process.
And for that reason, we've also included a way to validate a client's TLS certificate when the root CA certificate is provided, and reject connections without valid certificates.

## Documentation

Our documentation is also not left out from this round of improvements.
The biggest change is that addition of the FerretDB blog which was built on Docusaurus– an open source software - and which now resides in our centralized FerretDB repository.

We've added a section for CLI flags and environment variables.
On top of that, we've reformatted our documentation setup to ensure that the deployment URL is visible in logs and can be previewed, and we also added comments and warnings about Git LFS.

To learn about other changes on FerretDB 0.8.1, please read [our release notes](https://github.com/FerretDB/FerretDB/releases/tag/v0.8.1).

As always, we appreciate all our users, supporters, and the entire community that has been a part of the journey all to this moment.
You've all played a role in the growth of FerretDB, and we're excited to continue growing FerretDB because of your unwavering support.

Remember, if you have any questions or feedback, please [let us know](https://docs.ferretdb.io/#community)!
We're always here to help you get the most out of FerretDB.
