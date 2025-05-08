---
slug: ferretdb-v220-arm-architecture-support
title: 'FerretDB v2.2.0: `arm64` architecture support and more'
authors: [alex]
description: >
  FerretDB v2.2.0 introduces ARM architecture support, new evaluation image, documentation updates, and bug fixes.
image: /img/blog/ferretdb-v2.2.0.jpg
tags: [release]
---

![FerretDB v2.2.0: ARM architecture support and more](/img/blog/ferretdb-v2.2.0.jpg)

We are pleased to announce the release of FerretDB v2.2.0.

<!--truncate-->

This release provides full support for `arm64` architecture for both FerretDB and DocumentDB, new evaluation image with production-ready settings, and multiple improvements across the codebase and documentation.

Below is a detailed breakdown of what's new in this version.

## New features

### Full `arm64` support

Support for `arm64` architecture is one of the most requested features in FerretDB v2.
We are happy to announce that FerretDB and DocumentDB now offer full `arm64` support, broadening compatibility for users on ARM-based systems.
This update ensures smoother deployments and better performance on platforms like AWS Graviton and Apple Silicon.

## New evaluation image

A new FerretDB evaluation (`ferretdb-eval`) Docker image has been introduced, built with production-ready settings for easier evaluation and testing.
Having a separate image allows users to quickly test FerretDB without needing to set up a production environment.
We still provide the `ferretdb-eval-dev` image for debugging purposes, which includes features that make it slower.

See the [installation guide](https://docs.ferretdb.io/installation/evaluation/) for more details on how to use the new evaluation image.

## Enhancements

In addition to the new evalution image, several updates have been made to our Docker images.
Docker tags have been refined to handle pre-release `git` tags more reliably.
Health checks and supervision have been enhanced to improve container stability and maintainability.
The `state` directory now uses Docker volumes by default, which should help with data persistence.

## Documentation updates

We've expanded and refined our documentation to assist with smoother installation and update.

A new [Kubernetes Installation guide has been added for FerretDB](https://docs.ferretdb.io/installation/ferretdb/kubernetes/) and [PostgreSQL with DocumentDB extension](https://docs.ferretdb.io/installation/documentdb/kubernetes/) to make it easier to deploy FerretDB in Kubernetes environments.

We also added new instructions to help users easily update to new versions of FerretDB – [see our documentation for more](https://docs.ferretdb.io/installation/ferretdb/docker/).

Our documentation now includes a guide on setting up the FerretDB Data API, which allows users to interact with FerretDB using a RESTful API.
Check out the [Data API documentation](https://docs.ferretdb.io/usages/data-api/) for more information.

In an effort to improve our documentation, some of our guides have been tested and verified against CTS (Compatibility Test Suite) to ensure we provide accurate and user-friendly documentation.

## New contributors

A range of smaller tweaks and dependency updates were added to improve overall stability.
See our release notes for the [complete list of all changes](https://github.com/FerretDB/FerretDB/releases/tag/v2.2.0).

We're excited to welcome a new contributor: [@vardbabayan](https://github.com/vardbabayan) made their first contribution, helping to rename binaries and packages.
Thank you to everyone who contributed to this release!

Ensure to check out our [our GitHub](https://github.com/FerretDB) and [website](https://www.ferretdb.com) for more information on how to download, contribute, or explore enterprise solutions.

Have any questions?
Reach out to us on [our community channels](https://docs.ferretdb.io/#community).
