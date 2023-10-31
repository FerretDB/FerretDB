---
slug: building-mongodb-like-developer-experience-vultr
title: 'Building the MongoDB-like developer experience on Vultr'
authors:
  - name: Marcin Gwóźdź
    title: Director of Strategic Alliances at FerretDB
    url: https://www.linkedin.com/in/marcin-gwóźdź-277abaa9
    image_url: /img/blog/marcin-gwozdz.jpeg
  - name: Mike Biondi
    title: Engineering Manager of Cloud Native at Vultr
    image_url: /img/blog/mike-biondi.jpeg
description: >
  We enjoyed talking with Mike Biondi from Vultr about why they decided to build a MongoDB-compatible service using FerretDB.
image: /img/blog/ferretdb-vultr.jpg
tags: [open source, community, document databases, compatible applications]
---

![Building the MongoDB-like developer experience on Vultr](/img/blog/ferretdb-vultr.jpg)

Easy access to tools and freedom of choice are keys to developers' productivity.
On the market, we can find managed service providers (MSPs) with different offers; however, almost each of them appreciates the value that open-source solutions could provide to their users.

<!--truncate-->

By building a technology stack around open-source software, MSPs can offer developers precisely what they are looking for and be up to date with the latest trends.

We enjoyed talking with Mike Biondi from Vultr about why they decided to build a MongoDB-compatible service using [FerretDB](https://www.ferretdb.com).

## What is Vultr?

[Vultr](https://www.vultr.com/) offers industry-leading cloud GPU and cloud computing globally
without the complexity or cost of hyperscale cloud providers.
As a member of the [MACH Alliance](https://the.machalliance.org/book/vultr), Vultr aims to provide businesses and developers worldwide with unrivaled ease of use, price-to-performance, and global reach.

- **Vultr Cloud Compute:**
  Spin up and scale best-in-class VMs and Bare Metal in a few clicks in 32 cloud data center locations worldwide.
  Cloud Compute, powered by the latest-generation AMD and Intel CPUs, puts global developers and businesses in the driver's seat, providing on-demand enterprise-grade cloud computing at industry-leading price-to-performance.
  Easily deploy and scale shared or dedicated VMs and Bare Metal at any of Vultr's cloud data centers, paying only for what you use.

- **Vultr Cloud GPU:**
  Unleash the potential of flexible, large-scale, high-performance computing accelerated by NVIDIA data center GPUs.
  Cloud GPU, underpinned by the partnership of Vultr and NVIDIA, stands as a beacon for next-gen GPU-accelerated infrastructure.
  Sidestepping the usual complications of driver setups and licensing, it offers users a direct conduit to the raw power of NVIDIA GPUs for any computational endeavor.
  Supported cloud NVIDIA GPUs include the NVIDIA HGX H100, A100, L40S, A40, and A16.

- **Vultr Kubernetes Engine:**
  Harness the potential of containerized applications with Vultr Kubernetes Engine (VKE).
  A fully managed service, VKE mitigates Kubernetes' inherent complexities, ensuring operational confidence and scalability.
  With features like container orchestration, no management fee for the control plane, and resilient infrastructure that maximizes resource allocation, VKE is the definitive solution for modern deployment needs.

### What are the benefits for customers who decide to choose your platform?

_Vultr has global availability of full-stack cloud GPU and cloud computing from 32 cloud data centers across six continents. With Vultr, you get transparent, predictable pricing with bandwidth included, as well as access to a composable, best-of-breed ecosystem of cloud services through the [Vultr Cloud Alliance](https://www.vultr.com/cloudalliance/). Vultr is easy to use and API first, with no certification needed._

### What do you think about open source?

_We at Vultr believe open source is critical for unlocking innovation and enabling an entire ecosystem of infrastructure services to power enterprise solutions. Open source has enabled developers over the past 25+ years to experiment and build all of the technology we leverage today in the workplace, in our homes, and in our daily lives. Vultr is committed to providing developers with single-click access to any open source frameworks and technology to leverage our cloud infrastructure to power the next 25 years of innovation._

### How do your customers benefit from using open-source software?

_Open-source software has many benefits for small and large organizations, particularly cost, reliability, and security. The community-driven nature of these projects provides transparency and rapid iteration on updates, which enterprise software doesn't often match. Another key benefit is longevity, as open source solutions tend to be supported for a long time with new features and fixes added by developers and organizations who invest in seeing the project succeed because their very own projects often rely on it._

### Why is FerretDB an interesting solution for you?

_Document-based "NoSQL" databases can be helpful depending on the project, and this is an area our initial Managed Database offerings did not cover. With MongoDB's recent enterprise nature making it less accessible for independent developers, open-source drop-in replacements such as FerretDB felt like a natural fit to fill this gap. FerretDB's use of a PostgreSQL backend made it an exciting pair with our Managed PostgreSQL offering. It's also lovely to see how this product seamlessly covers the most common MongoDB functionality._

### What are your experiences while working with FerretDB?

_As noted in the official communication from FerretDB, the current state of the software as a drop-in replacement for MongoDB is that it consistently captures the bulk of the most common functionality. While there are certainly edge-case differences, some of which are part of the public roadmap, our experience with FerretDB has been smooth in that it works as a replacement precisely in the way it's being advertised. Most of the same commands and protocols work right out of the box, and it's been fast and smooth to work with while testing our managed product._

### What are you expecting from FerretDB in the future?

_FerretDB is off to a solid start as a drop-in replacement for MongoDB. We expect to see feature parity continue to grow going forward. The public roadmap posted on GitHub is very promising for this, and we would love to see the FerretDB team continue to deliver on it._

### What is the future of cloud services?

_The future of cloud services is compatible. Mix-and-match services from an open source and commercial software ecosystem are critical to giving developers the freedom, choice, and flexibility to build new things and move fast. We also believe the future is centered around a new cloud architecture where cloud-native engineering and ML engineering come together to deliver new MACH-compliant ML-powered applications leveraging CPUs and top-of-the-line NVIDIA GPUs at the edge._

## Conclusion

Using FerretDB, Vultr can offer customers a no-SQL database as a service alternative for those looking for a MongoDB alternative.
Vultr is an excellent proposition for companies who want to provide developers with familiar tools, no vendor lock-in, and reduce costs and complexity levels.

### About speakers

- Mike Biondi - Engineering Manager of Cloud Native at Vultr
- Marcin Gwóźdź - Director of Strategic Alliances at FerretDB

### About Vultr

Vultr aims to make high-performance cloud computing easy to use, affordable, and locally accessible for businesses and developers worldwide.
Vultr has served over 1.5 million customers across 185 countries with flexible, scalable, global Cloud Compute, Cloud GPU, Bare Metal, and Cloud Storage solutions.
Founded by David Aninowsky and completely bootstrapped, Vultr has become the world's largest privately-held cloud computing company without raising equity financing.
Learn more at [Vulr](https://www.vultr.com).

### About FerretDB

FerretDB is a truly open-source alternative to MongoDB built on Postgres.
FerretDB allows you to use MongoDB drivers seamlessly with PostgreSQL as the database backend.
Use all tools, drivers, UIs, and the same query language and stay open-source.
Our mission is to enable the open-source community and developers to reap the benefits of an easy-to-use document database while avoiding vendor lock-in and faux pen licenses.
We are not affiliated, associated, authorized, endorsed by, or in any way officially connected with MongoDB Inc. or any of its subsidiaries or its affiliates.
