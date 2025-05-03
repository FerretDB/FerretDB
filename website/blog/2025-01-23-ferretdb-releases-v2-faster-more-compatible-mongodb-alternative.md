---
slug: ferretdb-releases-v2-faster-more-compatible-mongodb-alternative
title: 'FerretDB Releases 2.0: A Faster, More Compatible MongoDB Alternative'
authors: [peter]
description: >
  We are pleased to announce the first release candidate of FerretDB v2.0, a significant milestone in our objective to provide a high-performance, truly Open Source alternative to MongoDB.
image: /img/blog/ferretdb-v2.png
tags: [release]
---

![We just launched FerretDB 2.0 RC. The Truly Open Source MongoDB Alternative, built on Postgres. Fast, free, and more compatible. Now powered by DocumentDB.](/img/blog/ferretdb-v2.png)

We are pleased to announce [the first release candidate of FerretDB v2.0](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0-rc.1),
a significant milestone in our objective to provide a high-performance, truly Open Source alternative to MongoDB.

<!--truncate-->

Building on the strong foundations of 1.x, FerretDB 2.0 introduces major improvements in performance, compatibility, support, and flexibility to enable more complex use cases.
Some of these highlights include:

- More than 20x faster performance powered by DocumentDB
- Higher feature compatibility
- [Vector search](http://docs.ferretdb.io/guides/vector-search/) support
- [Replication](http://docs.ferretdb.io/guides/replication/) support
- Extensive support and services

With FerretDB v2.0, users can now run their MongoDB workloads without license limitations or risk of vendor lock-in.

## Now powered by Microsoft's release of DocumentDB

FerretDB 2.x utilizes Microsoft's newly released open source [DocumentDB](https://github.com/microsoft/documentdb) PostgreSQL extension, which significantly increases database performance.
Among other improvements, DocumentDB introduces the BSON data type and operations to PostgreSQL, giving us the tools to store and query data much more efficiently than before.
Ensuring ongoing compatibility between DocumentDB and FerretDB enables users to run document database workloads on Postgres with increased performance and better support for their existing applications.

## Our mission and why we are releasing 2.x

FerretDB is an open-source, MongoDB-compatible database designed to free users from the constraints of proprietary databases.
It's perfect for users who want the flexibility of MongoDB without being tied to a specific vendor.

FerretDB remains committed to true open-source principles.
You can **run it anywhere** – on-premises, in the cloud, or in hybrid environments – making it a great choice for businesses that prioritize **portability** and **freedom from lock-in**.

FerretDB v1.x was instrumental in helping us lay the foundation for a truly open-source alternative for MongoDB.
It has served as a reliable solution for applications with lighter workloads or specific use cases, and it continues to be well-suited for those environments.
With the release of FerretDB 2.0, we are now focusing exclusively on supporting PostgreSQL databases utilizing DocumentDB, a Postgres extension moving forward.
However, for those who rely on earlier versions and backends, **FerretDB 1.x** remains available on our GitHub repository, and we encourage the community to continue contributing to its development or fork and extend it on their own.

The open source nature of FerretDB – **under the Apache 2.0 license** – ensures you can still build, improve, and extend its capabilities.
It also means you can deploy it anywhere – cloud, on-prem, or hybrid environments – without restrictions, making it a great choice for businesses that prioritize **portability** and **freedom from lock-in**.
This user first approach stands in stark contrast to many alternatives, which are tied to proprietary licenses or specific cloud platforms.

## What's new in FerretDB 2.0?

FerretDB 2.0 represents a leap forward in terms of **performance and compatibility**.
Thanks to changes under the hood, FerretDB is now up to **20x faster** for certain workloads, making it as performant as leading alternatives on the market.
Users who may have encountered compatibility issues in previous versions will be pleased to find that FerretDB now supports a **wider range of applications**, allowing more apps to work seamlessly.

Aside from performance gains, FerretDB continues to be fully **open source under the Apache 2.0 license**, which means you can deploy it anywhere – cloud, on-prem, or hybrid environments – without restrictions.

### Built for flexibility and openness

One of the most compelling advantages of FerretDB 2.0 is its **unmatched flexibility**.
Whether you're looking to migrate existing MongoDB workloads, extend your database capabilities, or explore new use cases like **vector search**, FerretDB gives you the tools you need without locking you into a single ecosystem.
You can run FerretDB wherever your infrastructure lives, _be it on-prem or any cloud_.

### FerretDB support services

With the release of v2.0, we are also launching a suite of **subscription and consulting services** to support enterprises looking to adopt FerretDB at scale.
These services offer expert guidance, dedicated support, and performance tuning to ensure you get the most out of your database setup.

If you're migrating from MongoDB or seeking ways to extend your current workloads, our team is ready to assist you.
For more information, check out our services page.

### FerretDB Cloud: Managed FerretDB as a service

As part of the FerretDB 2.0 launch, we're excited to introduce [**FerretDB Cloud**](https://cloud.ferretdb.com),
our fully managed database-as-a-service offering.
With FerretDB Cloud, you can effortlessly deploy and scale FerretDB on **AWS** and **GCP**, with support for Microsoft **Azure** and additional cloud providers coming soon.
It doesn't matter if you are running production workloads or testing new applications, FerretDB Cloud will take care of the operational overhead, so you can focus on building your apps.

We invite you to join the **waitlist** for early access to FerretDB Cloud by signing up [here](https://cloud.ferretdb.com/signup).
Stay tuned as we expand availability across more platforms in the near future!

Managed FerretDB will also be available on other providers as well, including Percona Ivee and Civo.
