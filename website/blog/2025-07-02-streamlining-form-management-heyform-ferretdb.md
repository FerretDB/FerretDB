---
slug: streamlining-form-management-heyform-ferretdb
title: 'Streamlining Form Management with HeyForm and FerretDB'
authors: [alex]
description: >
  Discover how to combine HeyForm's intuitive open-source form builder with FerretDB's reliable, PostgreSQL-backed database for powerful and flexible data collection.
image: /img/blog/ferretdb-heyform.jpg
tags: [compatible applications, open source, community]
---

![Enabling Feature Flags and A/B Testing with GrowthBook and FerretDB](/img/blog/ferretdb-growthbook.jpg)

Forms are the backbone of data collection for almost every application, from simple contact forms to complex surveys and application processes.
[HeyForm](https://heyform.net/) offers a modern, open-source solution for building and managing these essential forms with ease.

<!--truncate-->

At [FerretDB](https://www.ferretdb.com/), we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.

In this blog post, we're excited to explore how HeyForm, the intuitive open-source form builder, seamlessly integrates with FerretDB, offering a robust and high-performing solution for all your data collection needs.

## What is HeyForm?

HeyForm is an open-source form builder designed to simplify the creation, deployment, and management of online forms.
It provides a user-friendly interface and powerful features, including:

- **Drag-and-drop builder:** Easily create forms with various field types.
- **Customizable templates:** Start from pre-built templates or design your own.
- **Workflow automation:** Integrate forms into your existing workflows.
- **Data management:** Collect, view, and export submission data.
- **Self-hostable:** Full control over your data and infrastructure.

HeyForm focuses on providing a flexible and developer-friendly platform that helps you gather information efficiently and integrate it into your applications.

## Why use HeyForm with FerretDB?

HeyForm uses MongoDB as its primary database backend for storing form definitions, submissions, and user data.
Given that FerretDB is designed to be a wire-protocol compatible, open-source alternative to MongoDB, it can serve as a seamless drop-in replacement for HeyForm's database.
This powerful combination offers several compelling advantages:

- **Open-Source Synergy:** Both HeyForm and FerretDB are open-source projects, providing transparency, flexibility, and strong community backing, aligning perfectly with an open-source ethos.
- **PostgreSQL Reliability for Form Data:** By using FerretDB as HeyForm's backend, your critical form definitions and submission data benefit from PostgreSQL's battle-tested reliability, ACID compliance, and advanced data management capabilities.
- **Simplified Infrastructure:** If your existing data infrastructure is already based on PostgreSQL, integrating HeyForm with FerretDB can streamline your database management and reduce operational overhead.
- **No Vendor Lock-in:** Enjoy the freedom of truly open-source solutions without concerns about proprietary licensing or vendor lock-in.

## Connecting HeyForm to FerretDB

Connecting HeyForm to your FerretDB instance is straightforward, as HeyForm expects a MongoDB-compatible database.
Here's a step-by-step guide to get you started with a self-hosted HeyForm instance using docker run:

1. **Ensure FerretDB and Redis/KeyDB are running:** HeyForm uses MongoDB for primary data storage and Redis (or KeyDB, a Redis-compatible alternative) for caching and session management.
   Make sure both your FerretDB instance and your Redis/KeyDB instance are active and accessible.
   If you haven't set up FerretDB yet, refer to our [FerretDB Installation Guide](https://docs.ferretdb.io/installation/ferretdb/).
   For KeyDB, you can run it via Docker:

   ```sh
   docker run -d --name keydb -p 6379:6379 eqalpha/keydb
   ```

2. **Run HeyForm using Docker:** You can run HeyForm as a Docker container, specifying the necessary environment variables to connect to FerretDB and KeyDB.

   ```sh
   docker run -d \
     --name heyform \
     -p 9513:8000 \
     -v heyform_assets:/app/static/upload \
     -e APP_HOMEPAGE_URL="http://127.0.0.1:9513" \
     -e SESSION_KEY="your_secure_session_key_here" \
     -e FORM_ENCRYPTION_KEY="your_secure_encryption_key_here" \
     -e MONGO_URI="mongodb://<username>:<password>@host.docker.internal:27017/heyform" \
     -e REDIS_HOST="host.docker.internal" \
     -e REDIS_PORT="6379" \
     heyform/community-edition:latest
   ```

   Replace `<username>` and `<password>` with your FerretDB connection details.

3. **Access HeyForm:** nce the Docker container is running, HeyForm should connect to FerretDB and KeyDB, initialize its database, and be accessible via its web interface.
   You can access it at http://127.0.0.1:9513 (or whatever port you specified in the Docker command).

   You can now log into HeyForm, start building forms, and collect submissions, with all your data seamlessly stored in FerretDB.

## Example of HeyForm with data in FerretDB

The below image shows a simple form created in HeyForm, demonstrating its intuitive interface and ease of use.

![HeyForm Form Example](/img/blog/heyform-form-example.png)

After setting up HeyForm and creating a new form, you can inspect how HeyForm stores its data within FerretDB.

Connect to your FerretDB instance using a MongoDB shell or GUI tool (like MongoDB Compass or Mongo Express) and switch to the HeyForm database (default heyform):

```text
> use heyform
switched to db heyform
> show collections
appmodels
formanalyticmodels
formmodels
formreportmodels
integrationmodels
projectgroupmodels
projectmembermodels
projectmodels
submissioniplimitmodels
submissionmodels
teamactivitymodels
teaminvitationmodels
teammembermodels
teammodels
templatemodels
usermodels
usersocialaccountmodels
```

Now, let's query the `formmodels` collection to see the forms created in HeyForm by running `db.formmodels.findOne()`:

```js
{
  _id: 'zlGCuEdD',
  status: 1,
  draft: true,
  suspended: false,
  retentionAt: -1,
  reversion: 0,
  variables: [],
  logics: [],
  translations: {},
  hiddenFields: [],
  fields: [
    {
      title: [ 'Name' ],
      description: [],
      kind: 'short_text',
      validations: { required: false },
      properties: null,
      id: 'uTI3y9TR6Gzp',
      layout: {
        mediaType: 'image',
        mediaUrl: 'https://images.unsplash.com/photo-1646013532943-d5b86e8689b8?ixlib=rb-1.2.1&ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&auto=format&fit=crop&w=1080&q=80',
        brightness: 0,
        align: 'split_right'
      }
    },
    {
      title: [ 'Email' ],
      description: [],
      kind: 'email',
      validations: {},
      properties: {},
      id: 'jXWRpnXpbHXV'
    },
    {
      title: [ 'Thank you!' ],
      description: [ 'Thanks for completing this form. Now create your own form.' ],
      kind: 'thank_you',
      validations: null,
      properties: null,
      id: 'xH3vEBh0nzWM',
      layout: null
    }
  ],
  kind: 1,
  interactiveMode: 1,
  teamId: '8f2XC8N3',
  memberId: '6863fe5cde5a75001256241d',
  settings: {
    captchaKind: 0,
    filterSpam: false,
    active: false,
    published: true,
    allowArchive: true,
    requirePassword: false,
    locale: 'en',
    enableQuestionList: true
  },
  projectId: 'HjOiBwoM',
  name: 'ferret-form',
  createdAt: ISODate('2025-07-01T19:17:43.272Z'),
  updatedAt: ISODate('2025-07-01T19:50:09.958Z'),
  __v: 0,
  fieldUpdateAt: 1751399409
}
```

This output demonstrates that HeyForm successfully writes and reads its document-based data into FerretDB, which in turn stores it efficiently in PostgreSQL, providing a reliable backend for your form management.

![HeyForm Form Example](/img/blog/heyform-example.png)

## Conclusion

The integration of HeyForm and FerretDB provides a robust, scalable, and fully open-source solution for building and managing online forms.
By leveraging FerretDB's MongoDB compatibility, you can seamlessly integrate HeyForm into your PostgreSQL-backed stack, gaining the best of both worlds: flexible document modeling with the reliability of a relational database.

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Explore HeyForm's features and documentation](https://www.google.com/search?q=https://heyform.net/docs&authuser=1)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
