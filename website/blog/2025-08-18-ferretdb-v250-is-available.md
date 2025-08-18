---
slug: ferretdb-v250-is-available
title: 'FerretDB v2.5.0 is available!'
authors: [alex]
description: >
  FerretDB v2.5.0 now includes RPM packages for DocumentDB, changes to metric names, and new additions to our compatible apps documentation.
image: /img/blog/ferretdb-v2.5.0.jpg
tags: [release]
---

![FerretDB v2.5.0 is available](/img/blog/ferretdb-v2.5.0.jpg)

We are pleased to announce the release of FerretDB v2.5.0!

<!--truncate-->

This release introduces RPM packages for DocumentDB, changes to metrics, and some new additions to our compatible apps documentation.

This version works best with [DocumentDB v0.106.0-ferretdb-2.5.0](https://github.com/FerretDB/documentdb/releases/tag/v0.106.0-ferretdb-2.5.0).

## Breaking changes

### Metric names

We changed Prometheus metric names and updated their `HELP` text to make it clear that the current set of metrics is not yet stable.
FerretDB currently exposes Prometheus metrics on `/debug/metrics`.
If you scrape FerretDB metrics, please review and update your dashboards and alerts accordingly.

We plan to refine and promote some metrics to stable in the next release.

## What's Changed

### DocumentDB RPM packages

DocumentDB `.rpm` packages for Red Hat Enterprise Linux are now available!
Installation instructions are in the docs:
[https://docs.ferretdb.io/installation/documentdb/rpm/](https://docs.ferretdb.io/installation/documentdb/rpm/)

### Documentation

We continue to work on adding more compatible apps to our documentation.
This release includes new additions such as [WeKan](https://wekan.github.io/) and [Vault](https://www.vaultproject.io/).
You can find the complete list of compatible apps in our [documentation](https://docs.ferretdb.io/compatible-apps/).

In addition, images in our documentation can now be zoomed in with full resolution for easier viewing.

## Looking ahead

This release also includes various test improvements, dependency updates, and internal refactoring.
You can find the complete list in the [release notes](https://github.com/FerretDB/FerretDB/releases/tag/v2.5.0).

In the coming weeks, we are preparing several updates, including FerretDB Cloud and improvements to observability and metrics.
Stay tuned for more details.

Thanks to everyone who contributed to this release.
If you encounter any issues or have feedback, please reach out to us on any of [our community channels](https://docs.ferretdb.io/#community).
