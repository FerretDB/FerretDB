---
slug: building-robust-content-management-with-payload-cms-and-ferretdb
title: 'Building Robust Content Management with Payload CMS and FerretDB'
authors: [alex]
description: >
  Learn how to combine Payload CMS's powerful content management capabilities with FerretDB's reliable, PostgreSQL-backed database for a robust and flexible content solution.
image: /img/blog/ferretdb-payloadcms.jpg
tags: [compatible applications, open source, community]
---

![Powering Your Notification Infrastructure with Novu and FerretDB](/img/blog/ferretdb-payloadcms.jpg)

In modern web development, a flexible and efficient Content Management System (CMS) is crucial for managing diverse content effectively.
Developers increasingly seek powerful tools that offer both control and ease of use.

<!--truncate-->

Open-source CMS solutions have gained popularity for their flexibility and community support.
At FerretDB, we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.

In this blog post, we're thrilled to explore how [Payload CMS](https://payloadcms.com/), the open-source, code-first content management system, seamlessly integrates with FerretDB to replace MongoDB, offering a truly open-source solution for your content management needs.

## What is Payload CMS?

Payload CMS is a full-stack, code-first, headless CMS, and application framework.
It stands out by giving developers full control over their backend, enabling them to build robust content management systems with ease.

Payload CMS is highly extensible and can be used not just as a headless CMS, but also as a full-stack application framework, making it a versatile choice for a wide range of projects.

## Why use Payload CMS with FerretDB?

Payload CMS officially supports MongoDB as a primary database option, which makes it an excellent candidate for integration with FerretDB.
As a true open source alternative to MongoDB, it can serve as a drop-in replacement for Payload CMS's MongoDB backend.

This combination offers several compelling advantages:

- **Open-source ecosystem:** Both Payload CMS and FerretDB are open-source, providing transparency, flexibility, and a strong community backing.
- **Simplified infrastructure:** If your existing data infrastructure is already based on PostgreSQL, integrating Payload CMS with FerretDB can simplify your database management and reduce operational overhead.
- **No vendor lock-in:** Leverage the freedom of open-source solutions without concerns about proprietary licensing or vendor lock-in.

## Connecting Payload CMS to FerretDB

Connecting Payload CMS to your FerretDB instance is straightforward, as Payload CMS expects a MongoDB-compatible database.
Here's a step-by-step guide to get you started:

1. **Ensure FerretDB is running:** Make sure your FerretDB instance is active and accessible.
   If you haven't set it up yet, refer to our[FerretDB Installation Guide](https://docs.ferretdb.io/installation/ferretdb/).

2. **Set up a new Payload CMS project (if you don't have one):** You can quickly scaffold a new Payload CMS project using their CLI:

   ```sh
   npx create-payload-app@latest
   ```

   As shown below, you can also specify the project name, template, and database connection string during the setup process:

   ```text
   npx create-payload-app
   Need to install the following packages:
   create-payload-app@3.42.0
   Ok to proceed? (y) y
   ┌   create-payload-app
   │
   ◇   ────────────────────────────────────────────╮
   │                                               │
   │  Welcome to Payload. Let's create a project!  │
   │                                               │
   ├───────────────────────────────────────────────╯
   │
   ◇  Project name?
   │  ferretapp
   │
   ◇  Choose project template
   │  website
   │
   ◇  Select a database
   │  MongoDB
   │
   ◇  Enter MongoDB connection string
   │  mongodb://username:password@localhost:27017/payload-db
   │
   │
   ◇  Using pnpm.
   │
   ◇  Successfully installed Payload and dependencies
   │
   ◇  Payload project successfully created!
   │
   ◇   Next Steps
   │
   │
   │  Launch Application:
   │
   │    - cd ./ferretapp
   │    - pnpm dev or follow directions in README.md
   │
   │  Documentation:
   │
   │    - Getting Started: https://payloadcms.com/docs/getting-started/what-is-payload
   │    - Configuration: https://payloadcms.com/docs/configuration/overview
   │
   │
   │
   └   Have feedback?  Visit us on GitHub: https://github.com/payloadcms/payload.
   ```

3. **Confirm the database connection string:** You can always confirm or update the database connection string of your project from the `.env` file, and ensure the `DATABASE_URI` environment variable points to your FerretDB instance.

   ```text
   DATABASE_URI= mongodb://username:password@localhost:27017/payload-db
   ```

4. **Launch Payload CMS and test:** Once configured, launch your Payload CMS application (e.g., `npm run dev` or `yarn dev`).
   It should connect to FerretDB, initialize the necessary collections, and allow you to access the admin panel.
   You can now start defining your content models and managing data through the Payload CMS admin interface, with all data seamlessly stored in FerretDB.

## Example of Payload CMS with FerretDB as the backend

Here's an example of a simple `Posts` collection defined in Payload CMS.
This collection allows you to create and manage posts with fields like title, slug, hero section, layout blocks, and metadata.

![Image of Payload CMS post definition with FerretDB as the backend](/img/blog/payloadcms-post.png)

The data will be stored in your FerretDB instance.

Using a Mongo Client (like Compass), connect to your FerretDB instance and switch to the database you specified in your `.env` file (e.g., `payload-db`).

Querying the `posts` collection in FerretDB shows the content created through Payload CMS:

![Image showing Mongo Compass querying the posts collection in FerretDB](/img/blog/payloadcms-ferretdb-compass.png)

That shows the recently created page in the `posts` collection, which includes the title, slug, hero section, layout blocks, and metadata.

## Conclusion

The integration of Payload CMS and FerretDB provides a powerful, scalable, and fully open-source solution for your content management needs.
By leveraging FerretDB, you can run Payload CMS completely in open source, without vendor lock-in or restrictive licenses.

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
