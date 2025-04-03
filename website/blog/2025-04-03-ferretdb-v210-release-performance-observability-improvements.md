---
slug: ferretdb-v210-release-performance-observability-improvements
title: 'FerretDB releases v2.1.0 with performance and observability improvements'
authors: [alex]
description: >
  We are happy to announce the release of FerretDB v2.1.0 with significant performance and observability improvements.
image: /img/blog/ferretdb-v2.1.0.jpg
tags: [release]
---

![FerretDB releases v2.1.0 with performance and observability improvements](/img/blog/ferretdb-v2.1.0.jpg)

We just released FerretDB v2.1.0, which comes on the back of the [successful v2.0.0 GA release](https://blog.ferretdb.io/ferretdb-v2-ga-open-source-mongodb-alternative-ready-for-production/).

<!--truncate-->

The latest release provides improvements in performance and observability, and addresses key issues that were present in the previous version.
Other important changes include a fix for the embeddable Go package and a resolution to indexing issues that affected the previous version.

## Important update information

Due to the nature of the changes in this release, a direct update is not supported.
Users must perform a backup of their current databases and conduct a clean installation of FerretDB v2.1.0 and the DocumentDB extension before restoring their data.

Please note that is a one-time requirement due to the significant changes in this release.
Future updates will not require such a process.

To update to FerretDB v2.1.0, please follow these steps:

1. Backup your data with `mongodump` or `mongoexport`.

2. Remove any existing installations FerretDB and DocumentDB images, containers, and volumes to avoid conflicts.
   For Debian users, ensure to uninstall all packages and delete related data directories.

3. Once that's done, install FerretDB v2.1.0 along with the appropriate DocumentDB extension — either by pulling the correct Docker images or downloading and installing the `.deb` packages.

4. Restore your data using `mongorestore` or `mongoimport`.

5. After restoring your data, verify that everything is functioning as expected.
   Check the logs for any errors or warnings that may need attention.

If you encounter any performance issues or have questions during the update process, please reach out to us through [our community channels](https://docs.ferretdb.io/#community).
Your feedback is invaluable and helps us continually improve FerretDB.

## Other changes

### Performance and observability enhancements

In this release, we've made some improvements in performance and observability.
Notably, our logging mechanisms now provide more actionable information which should help in enabling better troubleshooting.

### Indexing issue resolved

We fixed an issue related to indexing that was present in previous versions.
With this fix, you can expect more reliable and efficient index operations.

### Fixed embeddable package

FerretDB v2.1.0 resolves an issue with our embeddable package, which was previously broken in version v2.0.0.
Developers can now seamlessly use FerretDB as a library within their Go programs, facilitating greater flexibility and integration capabilities.

## Looking ahead

Since the release of FerretDB v2.0.0, we have been working hard to address the feedback and issues reported by our users.
We are committed to providing a truly open-source alternative to MongoDB that's highly performant, compatible, and reliable for all your database needs.

Visit [our GitHub](https://github.com/FerretDB) and [our website](https://www.ferretdb.com) to download,
contribute, or explore enterprise solutions.
