---
slug: ferretdb-v210-release-performance-observability-improvements
title: 'FerretDB releases v2.1.0 with performance and observability improvements'
authors: [peter, aleksi]
description: >
  We are happy to announce the release of FerretDB v2.1.0, which brings significant performance improvements
  and observability enhancements.
image: /img/blog/ferretdb-v2-ga.png
tags: [release]
---

![FerretDB releases v2.1.0 with performance and observability improvements](/img/blog/ferretdb-v2-ga.png)

We just released FerretDB v2.1.0, which comes on the back of the successful v2.0.0 GA release.

<!--truncate-->

The latest release brings significant improvements in performance, observability, and addresses key issues to enhance your experience with FerretDB.
This version is a continuation of our commitment to providing a high-performance, fully open-source alternative to MongoDB, built on the robust foundation of PostgreSQL with DocumentDB extension.

Other important changes changes include a fix for the embeddable package and a resolution to an indexing issue that affected the previous version.
Let's dive into the details of what you can expect from FerretDB v2.1.0.

## Important upgrade information

Due to the nature of the changes in this release, a direct upgrade is not supported.
Users must perform a backup of their current databases and conduct a clean installation of FerretDB v2.1.0, followed by a data restoration.

Please note that this backup and restore procedure is a one-time requirement due to the significant changes introduced in this release.
Future upgrades are expected to follow a more streamlined process without necessitating a full backup and restore.

To upgrade to FerretDB v2.1.0, please follow these steps:

1. **Backup your data**

   - Use `mongodump` to create a backup of your existing databases.
   - Alternatively, `mongoexport` can be utilized for exporting data to JSON or CSV files.

2. **Remove existing installations**

   - **Docker setup:**

     - Stop and remove all FerretDB and DocumentDB containers.
     - Remove associated volumes and images to prevent conflicts.

   - **Debian setup:**
     - Uninstall the existing FerretDB and DocumentDB package using your package manager.
     - Ensure that all related data directories are cleaned up.

3. **Install FerretDB v2.1.0 and DocumentDB extension**

   - **Docker installation:**
     - Pull the latest FerretDB v2.1.0 image from the repository.

   - **Debian package installation:**
     - Download the FerretDB v2.1.0 .deb package and the DocumentDB extension package.
     - Install the package using `dpkg` or your preferred package manager.

4. **Restore your data**

   - Use `mongorestore` to import the data from your backup into the new installation.
   - If you used `mongoexport`, use `mongoimport` to restore your data accordingly.

After restoring your data, verify that everything is functioning as expected.
Check the logs for any errors or warnings that may need attention.

If you encounter any performance issues or have questions during the upgrade process, please reach out to us through [our community channels](https://docs.ferretdb.io/#community).
Your feedback is invaluable and helps us continually improve FerretDB.

## Other changes

### Performance and observability enhancements

In this release, we've made substantial strides in performance and observability with FerretDB v2.1.0.
Notably, we've refined our logging mechanisms to provide more insightful and actionable information, aiding in better system monitoring and troubleshooting.

### Indexing issue resolved

We've addressed a critical issue related to indexing that was present in previous versions.
With this fix, you can expect more reliable and efficient index operations, contributing to overall improved database performance.

### Fixed embeddable package

FerretDB v2.1.0 resolves an issue with our embeddable package, which was previously broken in version v2.0.0.
Developers can now seamlessly use FerretDB as a library within their Go programs, facilitating greater flexibility and integration capabilities.

## Looking ahead

We are committed to providing a seamless experience with FerretDB and provide you with the best possible tools for your database needs.
Visit [our GitHub](https://github.com/FerretDB) and [our website](https://www.ferretdb.com) to download,
contribute, or explore enterprise solutions.
Embrace the power of open source.
