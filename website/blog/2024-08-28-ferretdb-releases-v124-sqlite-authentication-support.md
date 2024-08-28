---
slug: ferretdb-releases-v124-sqlite-authentication-support
title: FerretDB releases v1.24 with SQLite authentication support
authors: [alex]
description: >
  We are happy to announce the release of FerretDB v1.24 with support for SQLite authentication, which is now available via the new experimental authentication feature.
image: /img/blog/ferretdb-v1.24.0.jpg
tags: [release]
---

![FerretDB v1.24 with SQLite authentication support](/img/blog/ferretdb-v1.24.0.jpg)

We are happy to announce the release of FerretDB v1.24 with support for SQLite authentication – now available via the experimental authentication mode.

<!--truncate-->

In the last few releases, we've been improving authentication on FerretDB through the additions of the [experimental authentication mode with support for `SCRAM-SHA-1` and `SCRAM-SHA-256`](https://blog.ferretdb.io/new-ferretdb-v121-release/), as well as [initial FerretDB user setup for Postgres](https://blog.ferretdb.io/new-ferretdb-v122-user-setup-feature/).
This release provides authentication support for the SQLite backend.

Some changes have also been made to resolve previous bugs and improve our documentation.

Let's dive into what's new in FerretDB v1.24.

## SQLite authentication support

This release introduces support for SQLite authentication.
You can easily secure your connection from scratch with user authentication as part of the initial setup.

You can set this up by enabling the experimental authentication mode `--test-enable-new-auth`/`FERRETDB_TEST_ENABLE_NEW_AUTH` and configuring your setup with the dedicated flags or environment variables.

For example:

```sh
ferretdb --test-enable-new-auth=true --setup-username=user --setup-password=pass --setup-database=ferretdb
```

Read our documentation on the [experimental authentication mode to learn more](https://docs.ferretdb.io/security/authentication/#experimental-authentication-mode).

## Embeddable package

As communicated in the previous release, this version renames the `SLogger` field to `Logger`, finishing the migration from `[zap](https://github.com/uber-go/zap)` to `[slog](https://pkg.go.dev/log/slog)`.

## Bug fixes

This release addressed an issue with building FerretDB using the `go build -trimpath` command.

The new version also fixes an issue with Docker's `HEALTHCHECK` in the production image by making the `HEALTHCHECK CMD` instructions as `exec` arrays instead of shell commands.
This change prevents errors when `/bin/sh` is missing.

We also fixed authentication issues with a C# driver.

## Other changes

To improve user experience across our documentation and blog, we've enabled zoom functionality on images to enable users to explore images in more detail.

We also made some updates to provide clearer guidance on managing and interpreting logs.

For a complete list of changes, please see the [release note for FerretDB v1.24](https://github.com/FerretDB/FerretDB/releases/tag/v1.24.0).

As always, we appreciate all the contributions and support for FerretDB.
We had 4 new contributors in this release – [@nalgeon](https://github.com/nalgeon), [@Evengard](https://github.com/Evengard), [@dasjoe](https://github.com/dasjoe), and [@kaiwalyakoparkar](https://github.com/kaiwalyakoparkar) made their first contribution.

Ensure to try out the SQLite authentication on the new release and let us know what you think.

If you have any questions or feedback, [reach out to us on our community channels](https://docs.ferretdb.io/#community).
