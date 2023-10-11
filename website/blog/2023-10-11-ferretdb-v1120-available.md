---
slug: ferretdb-v1120-available
title: FerretDB v1.12.0 is available with new PostgreSQL backend ready for testing
authors: [alex]
description: >
  We’ve just released FerretDB v1.12.0 with many new interesting updates on our new PostgreSQL backend, Docker images, arm64 binaries and packages, and more.
image: /img/blog/ferretdb-v1.12.png
tags: [release]
---

![FerretDB v.1.12.0 - new release](/img/blog/ferretdb-v1.12.png)

We've just released FerretDB v1.12.0 with many new interesting updates on our new PostgreSQL backend, Docker images, arm64 binaries and packages, and more.

<!--truncate-->

Let's dive in!

## New PostgreSQL backend

In the last couple of weeks, we've been working on migrating to the new PostgreSQL backend, and we're happy to announce that it's now available for testing in the new release.
If you're curious to know more on the new backend, [see here](https://blog.ferretdb.io/ferretdb-v1-10-production-ready-sqlite/#the-new-architecture).

At the moment, it's not enabled by default; you can enable it by setting `--postgresql-new` flag or `FERRETDB_POSTGRESQL_NEW=true` environment variable.
You can expect it to be enabled by default in the next release, so please stay tuned!

We encourage you to try out the new PostgreSQL backend and let us know what you think – we can't wait to learn about all your discoveries!

## arm64 binaries now available

We're also happy to announce that we've added linux/arm64 binaries and .deb/.rpm packages.
This has been a requested feature by the FerretDB community, and we're thrilled that we can finally provide them for you. Check them out [here](https://github.com/FerretDB/FerretDB/releases/).

## Docker images changes

We've also made changes to our Docker images.
Production Docker images use `scratch` as a base Docker image, with the only file present in the image being a FerretDB binary (with root TLS certificates embedded).

## Exciting time for the community

As we celebrate Hacktoberfest this month, we've had a growing number of contributors from the open-source community, and we're really happy about this.
In this release alone, we had 5 new contributors – [@Mihai22125](https://github.com/Mihai22125) [@Akhaled19](https://github.com/Akhaled19), [@rohitkbc](https://github.com/rohitkbc), [@princejha95](https://github.com/princejha95), and [@jrmanes](https://github.com/jrmanes) – and this so exciting!

Fostering the spirit of open source is a core mission of ours, where anyone – developers, writers, designers, etc – can feel comfortable contributing to community-driven open source projects.
This is why we wrote this blog post to assist new contributors get started in open source – [see it here](https://blog.ferretdb.io/how-to-contribute-to-open-source-2022/).

## Other changes

We understand the importance of metrics and stats for your databases and in this release, we've enabled basic metrics collection regarding the number of databases and collections on the Prometheus collector within the Registry.
We've also enabled extra details about indexes for `dbStats` response.

In addition, bugs from the previous release were addressed in this release.
For example, we've relaxed restrictions when `_id` is not the first field in projection, allowing it to be at any index.

Please see our [release notes](https://github.com/FerretDB/FerretDB/releases/tag/v1.12.0)

In recent weeks, the support and enthusiasm from the community have been remarkable.
We appreciate you!

Please be sure to try the new PostgreSQL backend architecture and the arm64 binaries and packages.
[Contact us here](https://docs.ferretdb.io/#community), we want to hear your experience with them!
