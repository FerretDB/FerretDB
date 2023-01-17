---
slug: how-to-keep-control-of-your-infra-using-ferretdb-and-tigris
title: "How to Keep Control of Your Data and Infrastructure Using FerretDB and Tigris"
author: Alexander Fashakin
image: ../static/img/blog/masaaki-komori-_we0BQQewBo-unsplash-1024x684.jpg
date: 2022-09-21
---

![Keep control of your data](../static/img/blog/masaaki-komori-_we0BQQewBo-unsplash-1024x684.jpg)

<!--truncate-->

Photo by[Masaaki Komori](https://unsplash.com/@gaspanik?utm_source=unsplash&amp;utm_medium=referral&amp;utm_content=creditCopyText) on [Unsplash](https://unsplash.com/s/photos/gate?utm_source=unsplash&amp;utm_medium=referral&amp;utm_content=creditCopyText)

No matter what stage a company is at, there’s always a palpable fear of vendor lock-in when adopting new software.
This fear becomes heightened when it comes to database systems.

Vendor lock-in is prevalent with proprietary software, where vendors make users sweat over leaving the platform to adopt a new one.
This essentially defeats the purpose of having a future-proof tech stack.

Besides, you still have to factor in the potential increase in usage or license cost for these proprietary solutions.
In a way, it's hardly surprising that cost savings and vendor lock-in are two main factors that drive the adoption of open-source database software, according to a [2020 study by Percona](https://www.percona.com/open-source-data-management-software-survey).

This blog post will explore the meaning and implications of vendor lock-in when adopting database software.

## What is vendor lock-in?

According to [Wikipedia](https://en.wikipedia.org/wiki/Vendor_lock-in), “vendor lock-in, also known as proprietary lock-in or customer lock-in, makes a customer dependent on a vendor for products, unable to use another vendor without substantial switching costs.”

Vendor lock-in makes you dependent on a single vendor or technology without an easy way to migrate or switch to new software in the future without incurring significant financial implications, legal constraints, integration, or compatibility issues.

In its place, companies can enjoy all the benefits that their license covers, including maintenance, regular updates, hosting options, etc.

### The drawbacks of vendor lock-in

Let’s take a look at some of the drawbacks of getting locked to a particular technology:

* **Fear of deprecation or service decline**: If software suddenly shuts down or fails to meet your business requirements, you'd want to seek alternatives and migrate your data.
But this may be sometimes impossible or difficult, meaning you’ll likely be stuck.

* **Sudden price surge**: A vendor may decide to raise its license costs or levy fees for particular services, knowing that you’re locked and cannot avoid paying.

* **Lack of control**: Using proprietary software means you’ll have no control over new feature additions, upgrades, hosting, maintenance, and integrations.
In the same vein, you won’t be able to customize or tailor the product to suit your specific business needs.

### The curious case of MongoDB Atlas

In recent times, the widespread adoption of cloud-based software solutions has made them prime candidates for vendor lock-in.
While some offer a way to host and manage on your own through specific cloud providers, there’s still some way to go before becoming vendor-independent.

Take MongoDB Atlas, for example.
MongoDB Atlas is a Database as a Service (DBaaS) solution used by many across the globe.
It is the cloud-based MongoDB platform that offers a full range of features including analytics, load balancing, database automation, search and others.

The platform promises to liberate you from vendor lock-in and enable you to host your backend in any of the three popular cloud providers – Google Cloud, AWS, and Azure.
But can you deploy MongoDB Atlas on your own, either on-prem or in the cloud?
Are you in total control of your data?
Putting it simply, are you truly vendor-independent?

With MongoDB Atlas, you don’t get access to the entire source code and you’re hardly in control of your data, which defeats the purpose of having fully independent software.

Another consideration is migration.
Moving away from MongoDB Atlas can be incredibly challenging and costly for your business; you have to pay additional cost when moving your data to a new vendor.
Besides, a migration process will incur costs no matter how long it runs and how many attempts you go through – successful or not.

While it may be technically possible to avoid lock-in in MongoDB Atlas, it doesn’t necessarily mean it’s feasible for most businesses.
In that case, how can you prevent vendor lock-in and future-proof your infrastructure and workload?

## How to avoid vendor lock-in in your database with FerretDB and Tigris Data

One of the best solutions to get out of vendor lock-in is to adopt open-source software.
By building your application database on OSS, you can safely manage, build, maintain control of the source code, and deploy anywhere you want - either on-prem or in the cloud.

[FerretDB](https://www.ferretdb.io/) and [Tigris Data](https://www.tigrisdata.com/) offer a potential alternative to MongoDB Atlas, where users can safely take control of their own data without the needless stress of a lock-in.
But how does it work?

FerretDB is an [open-source proxy](https://blog.ferretdb.io/mongodb-compatibility-whats-really-important/) that converts MongoDB  wire protocol queries to an underlying backend.
When it comes to Tigris, FerretDB provides a MongoDB translation layer that converts MongoDB queries to Tigris requests.
Adopting both solutions will ensure you have a developer-friendly, cloud-based, open-source database platform that can be deployed in any environment under your control – either on-prem or in the cloud under your account.

Building your application on FerretDB and Tigris Data gives you access to MongoDB syntax, operators, and methods, all without learning a new language or system, having your entire application backend on Tigris Data.

With the entire application data in your control, you won’t need to worry about getting trapped.
Besides, your entire infrastructure, including deployment, configuration, monitoring, and security, will be handled for you.

## Get started with FerretDB and Tigris

Vendor lock-in is a big concern for most companies, making it difficult to migrate, scale, or upgrade effectively without incurring substantial costs.
And just because a proprietary software like MongoDB promises a way out of the lock-in problem doesn’t mean it won’t be problematic.
Essentially, the possibility of avoiding a vendor lock-in situation doesn’t mean it's feasible.

With FerretDB and Tigris, you’ll be able to build your application backend and have control of all your data without the fear of getting locked in.
Click [here](https://www.tigrisdata.com/beta "") to sign up for the Tigris beta and get early access.

Read this article to learn more on [why you need FerretDB as the ideal replacement for MongoDB](https://blog.ferretdb.io/developers-need-ferretdb-stackoverflow-developer-survey-2022/ "").
