---
slug: new-ferretdb-v122-user-setup-feature
title: FerretDB releases v1.22 with an initial FerretDB user setup feature
authors: [alex]
description: >
  We have released FerretDB v1.22 with an initial FerretDB user setup feature, configurable document size limit, and several bug fixes and enhancements.
image: /img/blog/ferretdb-v1.22.0.jpg
tags: [release]
---

![FerretDB v1.22](/img/blog/ferretdb-v1.22.0.jpg)

We are thrilled to announce the release of FerretDB v1.22.0, which now includes an initial user setup feature, a configurable document size limit, and several bug fixes and enhancements.

<!--truncate-->

This release, one of the last in the FerretDB v1.x series, adds new features and improvements to make FerretDB even better.
Users can now set up an initial user for authentication, configure the document size limit, and benefit from several enhancements in this release.

As we continue to build and improve [FerretDB](https://www.ferretdb.com/), we are committed to providing a truly open-source document database that satisfies many MongoDB use cases.
In the background, we are working on FerretDB v2.0, which will drastically improve performance and compatibility.
It will also be a complete departure from our current architecture, and we can't wait to share it with you soon.

Read on to learn more about what's new in FerretDB v1.22.0.

## Enable initial FerretDB user setup

A standout feature in FerretDB v1.22 is the ability to set up an initial user for authentication.
This makes it easier for users to configure their FerretDB instance securely right from the start using dedicated flags or environment variables.
Here's how you can do it:

```sh
ferretdb --test-enable-new-auth=true --setup-username=user --setup-password=pass --setup-database=ferretdb
```

- `--setup-username`/`FERRETDB_SETUP_USERNAME`: Specifies the username to be created.
- `--setup-password`/`FERRETDB_SETUP_PASSWORD`: Specifies the password for the user (can be empty).
- `--setup-database`/`FERRETDB_SETUP_DATABASE`: Specifies the initial database that will be created.
- `--test-enable-new-auth`/`FERRETDB_TEST_ENABLE_NEW_AUTH`: Must be set to `true` to enable the new authentication setup.

Once the flags/environment variables are passed, FerretDB will create the specified user with the given password and the given database.

## Make maximum document size configurable

FerretDB previously had a 16MiB document size limit, which is not always practical today, especially when dealing with large documents or migrations.
This release introduces a configurable document size limit to address this issue.
You can now configure it using `--test-batch-size`/`FERRETDB_TEST_BATCH_SIZE` and `--test-max-bson-object-size-mi-b`/`FERRETDB_TEST_MAX_BSON_OBJECT_SIZE_MI_B` flags or environment variables.

## Bug fixes and enhancements

FerretDB now runs as a non-root user in production Docker images.
This change ensures the FerretDB process does not have root privileges within the container.
The `Dockerfile` has been updated to explicitly set ownership of the state directory to the `ferretdb` user, establishing the correct permissions even with anonymous volumes.

We have also improved the error message for invalid ownership/permissions of `state.json`.
This will provide clearer error messages when there are issues with the ownership or permissions of the `state.json` file and will also log the user and group.

We also made some adjustments to how document fields are sorted.
Document fields are now sorted in lexicographic order during updates.
For example, "bar" comes before "foo", "7" comes before "42", and "bar.7" comes before "bar.42".
This makes document field ordering consistent and predictable.

The new release also fixes the TCP port handler issue in Docker images.
It also resolved the batch-size error issue when using the embedded FerretDB package.

## Other changes

There are several other changes in this release, please [see the FerretDB v1.22.0 release changelog for more](https://github.com/FerretDB/FerretDB/releases/tag/v1.22.0).

There were four new contributors to FerretDB in this release: [@deferdeter](https://github.com/deferdeter), [@dockercui](https://github.com/dockercui), [@nullniverse](https://github.com/nullniverse), and [@pravi](https://github.com/pravi).
Thanks for contributing to FerretDB!

Our appreciation also goes out to all our contributors, the entire open-source community, and our users.

If you have any questions about FerretDB, please feel free to [reach out on any of our channels here](https://docs.ferretdb.io/#community).
