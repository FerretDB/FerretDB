---
slug: ferretdb-v2-ga-open-source-mongodb-alternative-ready-for-production
title: 'FerretDB 2.0 GA: Open Source MongoDB alternative, ready for production'
authors: [peter, aleksi]
description: >
  We are thrilled to announce the general availability (GA) of FerretDB v2.0,
  a groundbreaking release that delivers a high-performance,
  fully open-source alternative to MongoDB, ready for production workloads. 
image: /img/blog/ferretdb-v2-ga.png
tags: [release]
---

![We are thrilled to announce the general availability (GA) of FerretDB v2.0, a groundbreaking release that delivers a high-performance, fully open-source alternative to MongoDB, ready for production workloads.](/img/blog/ferretdb-v2-ga.png)

We are thrilled to announce the
[general availability (GA) of FerretDB v2.0](https://github.com/FerretDB/FerretDB/releases/tag/v2.0.0),
a groundbreaking release that delivers a high-performance,
fully open-source alternative to MongoDB, ready for production workloads.

<!--truncate-->

This release is a result of a year-long collaboration with our friends at Microsoft.
We worked together to combine the engine of
[Azure Cosmos DB for MongoDB (vCore)](https://learn.microsoft.com/en-us/azure/cosmos-db/mongodb/vcore/)
and FerretDB, releasing both under a fully open-source license.

Version 2.0 introduces transformative improvements in performance, compatibility, scalability, and flexibility.
Key highlights include:

- Over 20x faster performance powered by [Microsoft DocumentDB](https://github.com/microsoft/documentdb)
- Enhanced MongoDB compatibility for seamless application migration
- [Vector search](https://docs.ferretdb.io/guides/vector-search/) support for AI-driven use cases
- [Replication support](https://docs.ferretdb.io/guides/replication/) for high-availability
- [Enterprise-grade support and services](https://www.ferretdb.com/services)

With FerretDB v2.0 GA, organizations can now run MongoDB workloads at scale – free from proprietary licensing,
vendor lock-in, or platform restrictions.
Unlike SSPL-licensed MongoDB, it can also be provided as a service without limitations.

## Powered by Microsoft's DocumentDB

FerretDB 2.x utilizes Microsoft's newly released open-source DocumentDB PostgreSQL extension,
which significantly increases database performance.
Azure Cosmos DB for MongoDB (vCore) is built on the same extension,
enabling FerretDB and Cosmos DB users to move their workloads seamlessly between the two.

Among other improvements, DocumentDB introduces the BSON data type and operations to PostgreSQL,
giving us the tools to store and query data much more efficiently than before.
Ensuring ongoing compatibility between DocumentDB and FerretDB enables users to run document database workloads
on Postgres with increased performance and better support for their existing applications.

## Our Mission: Freedom Through Open Source

FerretDB was born to liberate developers and businesses from the constraints of closed-source databases.
As a true open-source project under the Apache 2.0 license, FerretDB empowers users to deploy anywhere –
on-premises, in the cloud, or in hybrid environments – without sacrificing portability or control.

While FerretDB 1.x remains available for legacy use cases,
the 2.0 GA release marks our full focus on PostgreSQL with DocumentDB as the backend.
This shift enables us to deliver enterprise-grade performance while staying true to open-source principles.
For users on earlier versions, FerretDB 1.x will remain accessible on GitHub,
and we encourage community contributions to its ongoing development.

## Unmatched Flexibility and Control

FerretDB 2.0 is built for organizations that refuse to compromise on flexibility.
Whether migrating MongoDB workloads, scaling globally, or experimenting with vector search,
FerretDB adapts to your infrastructure – not the other way around.
Run it on-premises, in any cloud, or across hybrid environments, all while avoiding ecosystem lock-in.

## Enterprise Support and Managed Services

To support large-scale deployments, we now offer enterprise-grade subscriptions and consulting services,
including dedicated support, performance tuning, and migration assistance.
If you are modernizing legacy systems or optimizing new projects, our team ensures a seamless experience.

Additionally, we're excited to introduce [FerretDB Cloud](https://cloud.ferretdb.com/),
a fully managed database-as-a-service offering.
Deploy and scale FerretDB effortlessly on AWS, Google Cloud, and (soon) Microsoft Azure,
with operational burdens handled by our experts.

Join the FerretDB Cloud waitlist [here](https://cloud.ferretdb.com/signup) for early access,
and stay tuned for expanded availability!

## The Beginnings of a New Open Standard for Document Databases

The collaboration of Microsoft and FerretDB marks the first step towards standardization
in the document database field: two powerful MongoDB alternatives join forces to provide a unified experience
and feature set for developers.
We believe that this is just the first step.
Visit [OpenDocDB.org](https://opendocdb.org) for more info about the initiative backed by FerretDB
and leading database companies like [Yugabyte](https://www.yugabyte.com) and [Percona](https://www.percona.com).

## What's Next

FerretDB 2.1 is scheduled to be released in the second half of March,
bringing more performance improvements for queries and aggregation pipelines that use indexes.
We also plan to improve observability, making diagnosing and understanding compatibility
and performance problems easier.
Over the following several versions, we plan to work on better session support and transactions.
But your feedback _will_ affect our [roadmap](https://github.com/orgs/FerretDB/projects/2/views/1).
The best time to share it is _now_!

## Get Started Today

FerretDB 2.0 GA is ready to serve your document database needs.
Visit [our GitHub](https://github.com/FerretDB) and [our website](https://www.ferretdb.com) to download,
contribute, or explore enterprise solutions.
Embrace the power of open source.
