---
slug: why-open-source-important-proprietary-uncertainties
title: Why Open Source Software is Important in the Face of Proprietary Uncertainties
authors: [alex]
description: MongoDB's recent deprecation of Data API and Device Sync highlights the risks of proprietary and SSPL-licensed solutions. Open source solutions like FerretDB offer control, community support, and long-term stability.
image: /img/blog/open-source-proprietary-risks.jpg
tags: [open source, sspl, document databases]
---

![Why open source software is important vs proprietary uncertainties](/img/blog/open-source-proprietary-risks.jpg)

A primary downside of non-open source solutions (proprietary, SSPL, and other open source pretenders) is that you're always at risk, at all times.
At any given moment, that software you depend on may be deprecated – and you're left with months or years of wasted effort.

<!--truncate-->

That was [the case with MongoDB over the past week with the deprecation of Data API and Device Sync](https://www.mongodb.com/docs/atlas/app-services/deprecation/#std-label-app-services-deprecation), among other features.

As of September 2024, the deprecated services include:

- Atlas Device Sync and Edge Server
- Atlas Data API and HTTPS Endpoints
- Atlas Device SDKs

By September 30, 2025, these services will reach end-of-life and be removed completely.
For those actively using these features, the cost of switching to an alternative may be substantial.
At [FerretDB](https://www.ferretdb.com/), this is why we advocate for open-source solutions.

## The risks of proprietary and SSPL-licensed solutions

Sadly, proprietary solutions can change their business models or services overnight without considering the community impact.

As you'll see from the comments in [this Reddit thread](https://www.reddit.com/r/mongodb/comments/1fct01v/fuck_you_mongodb/), Many developers and businesses now find themselves scrambling for alternatives, or they lose businesses or subscribers.
For others building apps that depend on those services, it's "back to the drawing board."

The Reddit comments highlight a recurring theme: developers feel blindsided, frustrated, and abandoned by services they once trusted.
Services, pricing, terms, or features change without notice and users are left with no choice but to quickly adapt or migrate to other services.

Besides, migrations from these solutions are never easy!
When a feature like the Data API is removed, alternative options are quite limited.

## Why open source solutions are important

Community is at the forefront of everything in open source!
Every feature, change, and development is done to meet the needs of the user.

With an open source solution, users continuously contribute to the development of the product and also extend and improve on it as much as they can.
There are transparent roadmaps, community feedback, and public development processes.
The eventual fate of certain features is not decided by a single vendor.

In contrast to proprietary solutions where you're always in danger of lock-in, open source software like FerretDB does not lock you into a specific ecosystem.

FerretDB offers a truly open-source MongoDB alternative database built on top of PostgreSQL.
You're free to host FerretDB on your own infrastructure, either on-premise, cloud, or hybrid – total control over your data and operations.

## FerretDB replacement for Data API

In light of the recent deprecation announcement, **[we have started working on an alternative for Data API](https://github.com/FerretDB/FerretDB/discussions/4578)**.
The first iteration of that HTTP API should be out in a few weeks, along with the first public version of FerretDB v2.
This next major version, built around the PostgreSQL extension, will bring significant compatibility and performance improvements suitable for most workloads.

Moving forward, developers and businesses should consider building their tech stacks with resilience in mind and prioritize open-source solutions that offer control, community support, and long-term stability.

Don't wait until the next deprecation or change affects your business!
Follow our [migration guide to start migrating from MongoDB to FerretDB](https://docs.ferretdb.io/migration/migrating-from-mongodb/).
