---
slug: ferretdb-v240-is-available
title: 'FerretDB v2.4.0 is available!'
authors: [alex]
description: >
  FerretDB v2.4.0 introduces user management capabilities, logging improvements, and improved compatibility via the latest DocumentDB release.
image: /img/blog/ferretdb-v2.4.0.jpg
tags: [release]
---

![FerretDB v2.4.0 is available](/img/blog/ferretdb-v2.4.0.jpg)

We're pleased to announce the release of FerretDB v2.4.0.
This version continues our focus on improving compatibility, enhancing performance, and extending support across a wider range of tools and applications.

<!--truncate-->

This release works best with [DocumentDB v0.105.0-ferretdb-2.4.0](https://github.com/FerretDB/documentdb/releases/tag/v0.105.0-ferretdb-2.4.0).

Be sure to update to DocumentDB v0.105.0-ferretdb-2.4.0 before updating FerretDB to v2.4.0.
Refer to [our guide for instructions on updating to the latest version](https://docs.ferretdb.io/installation/ferretdb/docker/#updating-to-a-new-version).

## New features

We've expanded user management capabilities; `clusterAdmin` users can now perform user management commands, making it easier to handle administrative tasks across your deployment.

## Enhancements & bug fixes

Under the hood, we've made a series of refinements, including updates to the DocumentDB integration for smoother operation and optimized request and response handling.
This release improves logging by shifting continuation events to debug-level messages, reducing clutter in standard logs while keeping detailed insights available for troubleshooting.
These changes collectively enhance performance and maintainability.

## Documentation

Our documentation and guides for compatible apps continue to grow.
We are glad to help users integrate FerretDB into a broader range of tools, including LibreChat, Payload CMS, NodeBB, GrowthBook, Heyform, DBeaver, and many more on our page.

Explore these new guides on our [documentation page](https://docs.ferretdb.io/compatible-applications/).

## Other changes

In recent weeks, we've also been working on adding YugabyteDB support in local development setups.
To see a complete list of all the changes, [check out the release notes](https://github.com/FerretDB/FerretDB/releases/tag/v2.4.0).

Thanks to everyone who contributed to this release.
If you encounter any issues or have feedback, please reach out to us on any of [our community channels](https://docs.ferretdb.io/#community).
