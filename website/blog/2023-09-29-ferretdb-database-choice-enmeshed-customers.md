---
slug: ferretdb-database-choice-enmeshed-customers
title: 'Why is FerretDB a database of choice for enmeshed customers?'
authors:
  - name: Marcin Gwóźdź
    title: Director of Strategic Alliances at FerretDB
    url: https://www.linkedin.com/in/marcin-gwóźdź-277abaa9
    image_url: /img/blog/marcin-gwozdz.jpeg
  - name: Julian König
    title: Senior Developer at j&s-soft GmbH
    image_url: /img/blog/julian.jpeg
  - name: Michael Feygelman
    title: Senior Project and Community Manager for enmeshed at j&s-soft GmbH
    image_url: /img/blog/michael-feygelman.jpeg
description: >
  We had a discussion with Julian König and Michael Feygelman from the enmeshed team on why they chose FerretDB, open source software, and avoiding vendor lock-ins.
image: /img/blog/ferretdb-enmeshed.png
tags: [open source, community, document databases, compatible applications]
---

![Why is FerretDB a database of choice for enmeshed customers?](/img/blog/ferretdb-enmeshed.png)

Modern times require modern solutions.
Every area of our lives can be enhanced by applications supporting us each day.

<!--truncate-->

However, when building the application, you cannot predict all use cases, and you have to constantly improve your work based on customers' feedback.
When we're talking about application design, it's important to avoid vendor lock-ins that may cause issues for customers in the future.

We were discussing such an approach with FerretDB users Julian König and Michael Feygelman from the enmeshed team:

**What is enmeshed, and why did you decide to create such software?**

_When you apply to several universities, you have to collect and send all your application documents several times: personal data, your high school diploma, internship certificates, proof of health insurance, sometimes an aptitude test, or proof of foreign language skills._

_Digital wallets such as [enmeshed](https://enmeshed.eu/) are now intended to provide a remedy, for example, with once-only principles. All important data and documents are stored locally on the user's device. Using the self-sovereign identities (SSI) paradigm, end users are given control and sovereignty over the data and documents they provide to external institutions and companies._

_enmeshed takes a very different approach from other SSI products with its centralized architecture: it combines the data sovereignty, security, and privacy of decentralized platforms with the transactional, support, and scalability capabilities of centralized architectures. In addition, there are no legal issues associated with the transfer of data over P2P networks - even if the data is encrypted. enmeshed aims to provide a low-threshold yet secure and GDPR-compliant ecosystem that can be used without prior knowledge._

_On the other hand, enmeshed can be easily integrated into existing processes, technologies, products, and landscapes. It acts more as a digitization satellite for end-user processes, enhancing an existing process rather than replacing it._

**Why did you choose FerretDB as your favorite database?**

_Our Connector uses MongoDB as its database engine. We received a feature request to support alternatives and started researching. When we found out about FerretDB, we looked at it and absolutely loved the idea behind it. We especially liked the fact that we didn't have to change any of our code, as FerretDB can be used with the same library we use for MongoDB._

_We release enmeshed under the MIT open source license. Because of our strategic focus and our policy-driven customers, we believe open source applications are the most sustainable form of software development. Therefore, whenever possible, we work with partners who share a similar strategy._

**What benefits can you see by using enmeshed with FerretDB?**

_Since FerretDB has many similarities to commercial database solutions, its ease of use and intuitive integration are significant advantages. We use it as a drop-in replacement for translating to SQL. Since most of our customers use relational databases (mostly PostgreSQL), they can install FerretDB with the included connector and get started immediately. The welcoming and responsive open source community, as well as the well-maintained documentation, make it very easy to work with._

**What is your experience during work with FerretDB engineers?**

_The collaboration has always been extremely friendly and efficient. The technical competence of the entire team was particularly noteworthy. Even though FerretDB was still in an early stage of maturity when we started working together, bug reports and requests were incorporated quite quickly._

**What are your expectations from FerretDB in the future?**

_enmeshed is in a state of constant growth. In 2024, the first educational institutions and learners will be connected, so the traffic will increase rapidly. We hope to work with FerretDB to meet this challenge and stabilize our application performance. FerretDB is also working with SAP on HANA compatibility. We'd like to use FerretDB in our other j&s-soft projects within our SAP development and consulting business._

**What is your customer's feedback?**

_Simply put, FerretDB works! Our Kubernetes Sidecar solution lets our customers connect to PostgreSQL using FerretDB. This greatly reduces the integration effort on the customer side. Another significant benefit is obviously the elimination of license fees._

Using open source software can provide a lot of advantages: it's easy to report missing features or bugs, and full code transparency provides users reliability and deep insight.
Additionally, it gives the possibility of contributing to developers who are seeing what is important to them.
At FerretDB, we're committed to supporting and listening to our customers, partners, and users' needs, with full support provided by our engineering team.

You can influence FerretDB by checking the [list of open issues on our GitHub](https://github.com/FerretDB/FerretDB/issues) and voting, reporting bugs, or contributing to those that are most important for you.

### About speakers

Marcin Gwozdz - Director of Strategic Alliances at FerretDB
Julian König – Senior Developer at j&s-soft GmbH
Michael Feygelman – Senior Project- and Community Manager for enmeshed at j&s-soft GmbH

### About enmeshed

enmeshed is an open source wallet that enables encrypted digital data exchange between educational institutions and learners.
It is an official part of the MVP of the National Networking Infrastructure for Digital Education of the Federal Ministry of Education and Research in Germany.
As a local data repository for transcripts and enrolment certificates, it can be integrated into existing school and university management systems in a GDPR-compliant manner.

### About FerretDB

FerretDB is a truly open-source alternative to MongoDB built on Postgres.
FerretDB allows you to use MongoDB drivers seamlessly with PostgreSQL as the database backend.
Use all tools, drivers, UIs, and the same query language and stay open-source.
Our mission is to enable the open-source community and developers to reap the benefits of an easy-to-use document database while avoiding vendor lock-in and faux pen licenses.

**We are not affiliated, associated, authorized, endorsed by, or in any way officially connected with MongoDB Inc., or any of its subsidiaries or its affiliates.**
