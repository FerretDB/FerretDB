---
slug: announcing-ferretdb-cloud-mongodb-compatible-documentdb
title: 'Announcing FerretDB Cloud: MongoDB-compatible database as a service, built on open-source DocumentDB'
authors: [peter, aleksi]
description: >
  We're excited to announce the launch of FerretDB Cloud, a fully managed,
  MongoDB-compatible database service built on an open-source DocumentDB.
image: /img/blog/ferretdb-cloud.png
tags:
  [cloud, document databases, mongodb compatible, open source, product, release]
---

![FerretDB Cloud is here! Easy to use MongoDB alternative as a service, built on Postgres. For those who don't like to be smacked with an atlas.](/img/blog/ferretdb-cloud.png)

We're excited to announce the launch of [FerretDB Cloud](https://cloud.ferretdb.com/), a fully managed,
MongoDB-compatible database service built on an [open-source DocumentDB](https://documentdb.io).
The DocumentDB project is
[backed by the Linux Foundation](https://www.linuxfoundation.org/press/linux-foundation-welcomes-documentdb-to-advance-open-developer-first-nosql-innovation),
as well as [AWS](https://aws.amazon.com/blogs/opensource/aws-joins-the-documentdb-project-to-build-interoperable-open-source-document-database-technology/),
[Microsoft](https://opensource.microsoft.com/blog/2025/08/25/documentdb-joins-the-linux-foundation/),
providing a solid foundation for a MongoDB-compatible document database alternative.

If you've ever wanted the convenience of MongoDB Atlas without the vendor lock-in,
licensing complexity when migrating, or a closed ecosystem, FerretDB Cloud was built for you.

<!--truncate-->

## Why FerretDB Cloud?

FerretDB was created with a mission: to make MongoDB workloads open, portable, and free of vendor lock-in.
Developers love MongoDB's query language and ecosystem, but not everyone wants to be tied to a single vendor.
FerretDB allows you to use existing MongoDB drivers and tools while running on top of PostgreSQL,
one of the most trusted open-source databases in the world.
While FerretDB is not 100% MongoDB compatible, it offers enough features to run MongoDB workloads,
often with no code changes.

With FerretDB Cloud, this experience becomes even simpler.
Instead of managing infrastructure, patching, or scaling PostgreSQL yourself, you can now provision
and run FerretDB in minutes.
Your applications still connect as if they were talking to MongoDB, but your data is stored in PostgreSQL
with an open-source DocumentDB extension, giving you reliability and long-term freedom.

In MongoDB Atlas, there are features that are only available as a service.
Using these features will make you vulnerable to vendor lock-in,
since there are no off-the-shelf alternatives on the market.
In contrast, FerretDB Cloud provides the same exact feature set as a local FerretDB instance.
For example, vector search in MongoDB is only available on Atlas,
while both FerretDB Cloud and on-prem FerretDB offer this right out of the box.

## Features

FerretDB Cloud offers multiple tiers, based on your needs.
Our packages are suitable for personal use, all the way to the various needs of large Enterprise workloads.
Our Free Tier also offers the opportunity to try FerretDB easily.

| Feature                                       | Free Tier    | Pro Tier                                   | Enterprise Tier                            | BYOA (Enterprise in Customer's Account)    |
| --------------------------------------------- | ------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| Tenancy Type                                  | Multi-tenant | Dedicated Tenancy                          | Dedicated Tenancy                          | Dedicated Tenancy                          |
| Default storage                               | 10 Gi        | Up to 64 Ti                                | Up to 64 Ti                                | Up to 64 Ti                                |
| Access control / RBAC                         | ✅️          | ✅️                                        | ✅️                                        | ✅️                                        |
| Cluster management REST APIs                  | ✅️          | ✅️                                        | ✅️                                        | ✅️                                        |
| Audit logs                                    | ✅️          | ✅️                                        | ✅️                                        | ✅️                                        |
| Metrics and logs dashboard                    | ✅️          | ✅️                                        | ✅️                                        | ✅️                                        |
| Encryption in transit (TLS)                   |              | ✅️                                        | ✅️                                        | ✅️                                        |
| Encryption at rest                            |              | ✅️                                        | ✅️                                        | ✅️                                        |
| Multi-cloud                                   |              | Coming soon                                | Coming soon                                | ✅️                                        |
| E2E billing                                   |              | ✅️                                        | ✅️                                        | ✅️                                        |
| PMM - Percona Monitoring and Management Stack |              |                                            | ✅️                                        | ✅️                                        |
| Custom networks + VPC peering                 |              |                                            | ✅️                                        | N/A                                        |
| SOC2-ready                                    |              |                                            | ✅️                                        | ✅️                                        |
| Backups                                       |              | 24h RTO<br />7d retention<br />sub-min RPO | 1h RTO<br />30d retention<br />sub-min RPO | 1h RTO<br />30d retention<br />sub-min RPO |
| SLA                                           |              | 99.90%                                     | 99.99%                                     | 99.99%                                     |
| Support                                       | Basic        | Priority                                   | Enterprise                                 | Enterprise                                 |

Currently, FerretDB Cloud is available only on AWS
(including [AWS Marketplace](https://aws.amazon.com/marketplace/seller-profile?id=seller-ttfkkekh4dm5g)).
We plan to add support for Microsoft Azure and Google Cloud shortly.
We are really excited about building the first cross-cloud DocumentDB-based solution.

## Getting Started

You can try FerretDB Cloud today at https://cloud.ferretdb.com/.
Sign up, create your first instance, and connect using your existing MongoDB drivers.
It's a similar developer experience you already know.

(Note: until further notice, new FerretDB Cloud subscriptions are subject to approval from a waitlist.)
