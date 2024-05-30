---
slug: postgresql-can-it-be-a-database-for-everything
title: 'PostgreSQL - can it be a database for everything?'
authors:
  - name: Marcin Gwóźdź
    title: Director of Strategic Alliances at FerretDB
    url: https://www.linkedin.com/in/marcin-gwóźdź-277abaa9
    image_url: /img/blog/marcin-gwozdz.jpeg
  - name: Samay Sharma
    title: Chief Technology Officer at Tembo
    image_url: /img/blog/samay-sharma.jpeg
description: >
  We recently had the opportunity to speak with the Tembo team and ask about their thoughts on the PostgreSQL ecosystem, how it can be a database for everything, and how FerretDB can be used with Tembo.
image: /img/blog/ferretdb-tembo-qa.jpg
tags:
  [
    open source,
    community,
    document databases,
    compatible applications,
    postgresql tools,
    cloud
  ]
---

![PostgreSQL - can it be a database for everything?](/img/blog/ferretdb-tembo-qa.jpg)

[PostgreSQL](https://www.postgresql.org/) is one of the most popular databases around the world.
A lot of companies have decided to build a business around that.

<!--truncate-->

There are many PostgreSQL experts around the world, and we can't even count the number of applications powered by this database.

Why is this possible?
PostgreSQL is open-source, so anyone can use it and learn easily and without limitations.
Every year more companies are founded and trying to find their niche.

We recently had the opportunity to speak with the [Tembo](https://tembo.io/) team and ask about their thoughts on the PostgreSQL ecosystem.

**What is Tembo? What is unique?**

_At Tembo, our goal is to productize the entire Postgres ecosystem into a developer-friendly platform, so that developers need to use less tools in their data stack. The modern data stack has become a sprawling landscape even though Postgres, with its extension ecosystem, can solve a lot of those problems. However, it's very hard to do in practice._

_We want it to be easy to use Postgres for non-typical use cases and, over time, for everything._

_To that end, we provide [Tembo Stacks](https://tembo.io/blog/tembo-stacks-intro), which are curated selections of extensions, apps, and Postgres configurations that are designed to address particular use cases. They are available as [open source](https://github.com/tembo-io/tembo/tree/main/tembo-operator/src/stacks), deployable with our Kubernetes operator, and are also available for a single-click deploy on [Tembo Cloud](https://cloud.tembo.io/)._

**Why did you decide to start such a company? What is the feedback from early adopters? Why PostgreSQL?**

_As we outline in the [Tembo Manifesto](https://tembo.io/blog/tembo-manifesto/), the modern data stack has too many moving parts. Each tool promises a solution to a data problem, yet collectively they contribute to ever-growing complexity and cost._

_PostgreSQL is a stable, community-supported, reliable open source project with a well-earned reputation. With extensions, it can support a large number of these use cases. Eg. you can do vector search with pgvector, ML with postgresml, OLAP with columnar, geospatial with PostGIS, and mongo with ferretdb._

_However, you can easily get stuck discovering which extensions to use, how to install and put them together, and how to configure Postgres appropriately for your use case. We aim to make it extremely easy for developers to discover and deploy extensions, so that they can use Postgres for everything._

_Overall feedback has been very encouraging. We launched our managed Cloud platform ~6 months ago, and developers from over 750 organizations have tried out Tembo Cloud. We've also received positive feedback for our community contributions, including pgmq, trunk, pg_later, and pg_vectorize._

**How do you believe Tembo is helping develop the PostgreSQL community?**

_More often than not, we hear that developers use a very small fraction of PostgreSQL capabilities, including extensions. It's hard for developers to discover, evaluate, trust, install, and successfully use extensions. Our goal is to enable every PostgreSQL user to use more extensions and to use Postgres for more use cases._

_We're now working to design a community solution to address the challenges related to extension discovery and distribution, and an initial version of the [proposal](https://gist.github.com/theory/898c8802937ad8361ccbcc313054c29d) is already under discussion._

_We've also contributed a number of extensions to the ecosystem._

**Why did you choose an open-source license for Tembo?**

_We're first and foremost, a Postgres company, and our philosophies are aligned with the community. While Trunk and Tembo Stacks are integrated into our platform, they are PostgreSQL licensed, and usable without our SaaS._

_We have also released all our extensions (namely pgmq, pg_later, pg_vectorize, prometheus_fdw, clerk_fdw) under the PostgreSQL license._

_Open sourcing software projects offer transparency and encourage community participation. We want to enable everybody to help us improve our products via issues or contributions. This allows us to build the best products and benefit from the wisdom of others._

**What is important for you in the open-source project?**

_We want to encourage collaboration on our open source projects. We want developers to feel a part of the community and make it easy for them to contribute to our projects._

_In fact, we now have a person who is a codeowner on pgmq who is not employed by Tembo. We consider this an important metric for the project's success._

**How do you see the future of open-source databases?**

_Open-source databases are here to stay. Both PostgreSQL and MySQL have been around for decades, and most new databases are also open source projects. Vendor lock-in is a genuine problem, especially for databases because of their critical-ness for a business._

_Being based on an open-source database gives us a lot of options and flexibility and the ability to report issues and contribute back as we see fit._

_Postgres is even more unique because it is genuinely a community-run project, and in spite of so many forks which have come and gone, Postgres has outlasted all of them. The future of open-source databases is bright._

**What is interesting for you in FerretDB? What use cases can you imagine for using Tembo and FerretDB together?**

_FerretDB is a perfect example of how one can use Postgres for more use cases than what it's typically been used for. Using FerretDB, developers can use Postgres to power their document store workloads without having to make application changes. That makes it much more attractive for them to migrate._

_Using FerretDB, we built our Mongo Alternative on Postgres stack, to allow users to benefit from getting a managed Postgres experience with the API being powered by Ferret. That way, developers who aren't experts in setting up and running Postgres operationally can also benefit from Ferret in a one-click manner._

**What was the biggest challenge during the integration of FerretDB with Tembo?**

_Integrating FerretDB with Tembo was an exciting process and a team effort. One of the key steps was to introduce new routing configurations to our Kubernetes operator. Before working on the FerretDB integration, the [tembo-operator](https://github.com/tembo-io/tembo/tree/main/tembo-operator) only supported TCP ingress for services communicating with Postgres, and HTTP ingress for any [application services](https://tembo.io/blog/tembo-operator-apps) running in pods next to Postgres. So in order for our users to successfully communicate with the FerretDB container, we need to build an API to allow appServices to request a TCP ingress from Kubernetes. All of this comes together to support a user experience which only requires downloading an SSL certificate and running a mongosh connection string to get things up and running._

**What are your expectations from FerretDB in the future?**

_We would love to partner with FerretDB even more deeply. We would like to hear about best practices and how we could optimize Postgres to give the best MongoDB alternative experience to our users. We are also excited to share any feedback we receive from our users about FerretDB with the team to improve the product._

**We understand Tembo believes in the philosophy of PostgreSQL being a "one database for everything". What are some key challenges you think need to be overcome to make that a reality?**

_First of all, Postgres has been designed for extendability, so the potential is clearly established._

_Next, Postgres can be very performant when configured well for a specific workload. That's why we combined the both of them to build stacks, which are optimized Postgres instances to tackle specific workloads._

_A key challenge is going to be comparing PostgreSQL to all other databases, so we can prove to our users that Postgres and its ecosystem of extensions is actually enough for most use cases. Once they see the potential of what these stacks can power, there's nothing which will stop them from picking "Postgres for everything"._

## Conclusion

As we can see, PostgreSQL is a powerful database with limitless capabilities.
The idea of open source allows us to create new ideas around it, and the whole ecosystem provides the necessary pieces to bring the ideas to life.

FerretDB is one of the pieces you can use to enhance your solution - by adding MongoDB compatibility to your PostgreSQL environment, you're adding flexibility to make life easier for your developers.
The key to success is to provide tools as simple as can be so developers will want to use them in different use cases.

The community may suggest the direction for evolving the project, which is another great advantage of open source philosophy and can speed up the process.
And who knows?
Maybe at one point, PostgreSQL may become one database for everything.

[Check out FerretDB on Github](https://github.com/FerretDB/FerretDB).

[Check out Tembo on GitHub](https://github.com/tembo-io/tembo).

### About speakers

- Samay Sharma - Chief Technology Officer - Tembo
- Marcin Gwozdz - Director of Strategic Alliances at FerretDB

### About Tembo

Tembo is the Postgres developer platform for building every data service.
We collapse the database sprawl and empower users with a high-performance, fully-extensible managed Postgres service.
With Tembo, developers can quickly create specialized data services using Stacks, pre-built Postgres configurations and deploy without complex builds or additional data teams.

### About FerretDB

FerretDB is a truly open-source alternative to MongoDB built on Postgres.
FerretDB allows you to use MongoDB drivers seamlessly with PostgreSQL as the database backend.
Use all tools, drivers, UIs, and the same query language and stay open-source.
Our mission is to enable the open-source community and developers to reap the benefits of an easy-to-use document database while avoiding vendor lock-in and faux pen licenses.

We are not affiliated, associated, authorized, endorsed by, or in any way officially connected with MongoDB Inc., or any of its subsidiaries or its affiliates.
