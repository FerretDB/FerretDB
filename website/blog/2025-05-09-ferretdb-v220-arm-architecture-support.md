---
slug: ferretdb-v220-arm-architecture-support
title: 'FerretDB v2.2.0: ARM64 architecture support and more'
authors: [alex]
description: >
  FerretDB v2.2.0 introduces ARM64 architecture support, changes to the evaluation image, and several improvements across the codebase and documentation.
image: /img/blog/ferretdb-v2.2.0.jpg
tags: [release]
---

![FerretDB v2.2.0: ARM64 architecture support and more](/img/blog/ferretdb-v2.2.0.jpg)

FerretDB v2.2.0 is now available!

<!--truncate-->

This release provides full support for `arm64` architecture for both FerretDB and DocumentDB, changes to the evaluation image, and several improvements across the codebase and documentation.

Below is a summary of what's new in this version.

## Update guide

We added new update instructions to assist users moving to newer versions of FerretDB and DocumentDB.
For updating Docker images, refer to [the Docker update guide](https://docs.ferretdb.io/installation/ferretdb/docker/#updating-to-a-new-version).
For updating Debian packages, see [the Debian update guide](https://docs.ferretdb.io/installation/ferretdb/deb/#updating-to-a-new-version).

## Full `arm64` support

One of the most requested enhancements, full `arm64` support, is now available for both FerretDB and DocumentDB.
This broadens compatibility for users on ARM-based systems, ensuring smoother deployments and better performance on platforms like AWS Graviton and Apple Silicon.

## Evaluation images

We now provide two evaluation images:

- `ferretdb-eval-dev`: The existing evaluation image, previously named `ferretdb-eval`, which uses development builds of FerretDB and DocumentDB, has been renamed to `ferretdb-eval-dev`.
  It remains intended for debugging purposes.
- `ferretdb-eval`: A new image built with production builds of FerretDB and DocumentDB, recommended for evaluation and testing purposes.

For setup instructions and additional details, [see the evaluation installation guide](https://docs.ferretdb.io/installation/evaluation/).

## Documentation updates

We've expanded and updated our documentation to assist with smoother deployments and updates.

New guides are now available for [deploying both FerretDB](https://docs.ferretdb.io/installation/ferretdb/kubernetes/) and [PostgreSQL with DocumentDB extension](https://docs.ferretdb.io/installation/documentdb/kubernetes/) in Kubernetes environments.

Our documentation now includes a guide on setting up the FerretDB Data API, which allows users to interact with FerretDB using a RESTful API.
See the [Data API documentation](https://docs.ferretdb.io/usage/data-api/) for more information.

## Other changes

This release also includes a range of maintenance and stability improvements, such as dependency updates, minor codebase tweaks, and infrastructure changes.
See our release notes for the [full list of all changes](https://github.com/FerretDB/FerretDB/releases/tag/v2.2.0).

We also welcome a new contributor: [@vardbabayan](https://github.com/vardbabayan), who contributed by renaming binaries and packages.

Thank you to everyone who contributed to this release!

Be sure to check out [our GitHub](https://github.com/FerretDB) and [website](https://www.ferretdb.com) for more information on how to download, contribute, or explore enterprise solutions.

Curious about the story behind FerretDB and why we chose PostgreSQL as the database backend?
[Check out this episode of the Talking Postgres podcast](https://talkingpostgres.com/episodes/how-i-got-started-with-ferretdb-why-we-chose-postgres-with-peter-farkas), where FerretDB CEO, Peter Farkas, shares insights into the project's origins and technical choices.
The conversation covers everything from early inspiration (including a trek to K2 base camp) to open source principles and FerretDB's integration with the open source DocumentDB extension from Microsoft

Have any questions or feedback?
Join us on [our community channels](https://docs.ferretdb.io/#community).
