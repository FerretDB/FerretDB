---
slug: streamlining-form-management-heyform-ferretdb
title: 'Streamlining Form Management with HeyForm and FerretDB'
authors: [alex]
description: >
  Discover how to combine HeyForm's intuitive open-source form builder with FerretDB's reliable, PostgreSQL-backed database for powerful and flexible data collection.
image: /img/blog/ferretdb-heyform.jpg
tags: [compatible applications, open source, community]
---

![Streamlining Form Management with HeyForm and FerretDB](/img/blog/ferretdb-heyform.jpg)

Forms are the backbone of data collection for almost every application, from simple contact forms to complex surveys and application processes.
HeyForm offers a modern, open-source solution for building and managing these essential forms with ease.

<!--truncate-->

At [FerretDB](https://www.ferretdb.com/), we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.

In this blog post, we're excited to explore how HeyForm seamlessly integrates with FerretDB, a drop-in replacement for MongoDB, ensuring a fully open-source stack for your form management and data collection needs.

## What is HeyForm?

[HeyForm](https://heyform.net/) is an open-source form builder designed to simplify the creation, deployment, and management of online forms.
It provides a user-friendly interface and powerful features, including:

- **Customizable templates:** Start from pre-built templates or design your own.
- **Workflow automation:** Integrate forms into your existing workflows.
- **Data management:** Collect, view, and export submission data.
- **Self-hostable:** Full control over your data and infrastructure.

HeyForm focuses on providing a flexible and developer-friendly platform that helps you gather information efficiently and integrate it into your applications.

## Why use HeyForm with FerretDB?

HeyForm uses MongoDB as its primary database backend for storing form definitions, submissions, and user data.
Given that FerretDB is designed to be a truly open source alternative to MongoDB, it can serve as a drop-in replacement for HeyForm's database.
This powerful combination offers several compelling advantages:

- **Open-source:** Both HeyForm and FerretDB are open-source projects, providing transparency, flexibility, and strong community backing, aligning perfectly with an open-source ethos.
- **Simplified infrastructure:** If your existing data infrastructure is already based on PostgreSQL, integrating HeyForm with FerretDB can streamline your database management and reduce operational overhead.
- **No vendor lock-in:** Enjoy the freedom of truly open-source solutions without concerns about proprietary licensing or vendor lock-in.

## Connecting HeyForm to FerretDB

Connecting HeyForm to your FerretDB instance is straightforward, as HeyForm uses MongoDB for primary data storage and KeyDB for caching and session management.
Refer to the [HeyForm self-hosting documentation](https://docs.heyform.net/open-source/self-hosting) for more details on its configuration options.

Here's a step-by-step guide to get you started:

1. **Setup Docker Compose file:**
   Create a `docker-compose.yml` file with the following content to define the services for HeyForm, FerretDB, and KeyDB (for caching).

   ```yaml
   services:
     heyform:
       image: heyform/community-edition:latest
       restart: always
       volumes:
         # Persist uploaded images
         - ./assets:/app/static/upload
       depends_on:
         - ferretdb
         - keydb
       ports:
         - '9513:8000'
       environment:
         APP_HOMEPAGE_URL: http://127.0.0.1:9513
         SESSION_KEY: key1
         FORM_ENCRYPTION_KEY: key2
         MONGO_URI: 'mongodb://<username>:<password>@ferretdb:27017/heyform'
         REDIS_HOST: keydb
         REDIS_PORT: 6379

     ferretdb:
       image: ghcr.io/ferretdb/ferretdb-eval:2
       restart: on-failure
       ports:
         - 27017:27017
       environment:
         - POSTGRES_USER=<username>
         - POSTGRES_PASSWORD=<password>
         - POSTGRES_DB=postgres
       volumes:
         # Persist FerretDB data
         - ferretdb_data:/var/lib/postgresql/data

     keydb:
       image: eqalpha/keydb
       restart: always
       command: keydb-server --appendonly yes --protected-mode no
       volumes:
         # Persist KeyDB data
         - keydb:/data

   volumes:
     ferretdb_data:
     keydb:
   ```

   Replace `<username>` and `<password>` with your FerretDB connection details.

   The above YAML file uses the latest HeyForm community edition image, FerretDB evaluation image, and KeyDB.
   The FerretDB evaluation image comes with FerretDB and PostgreSQL with DocumentDB extension, which is suitable for quick testing and experiments.

2. **Start the services:**
   Make sure you have [Docker](https://www.docker.com/) installed on your machine.

   Then, run the following command to start all services in the `docker-compose.yml` file.

   ```sh
   docker-compose up -d
   ```

   This command will pull the necessary images, create containers for HeyForm, FerretDB, and KeyDB, and start them in detached mode.

3. **Access HeyForm:** Once all the services are running, HeyForm should connect to FerretDB and KeyDB, initialize its database, and be accessible via its web interface.
   You can access it at http://127.0.0.1:9513 (or whatever port you specified in the Docker command).

   You can now log into HeyForm, start building forms, and collect submissions, with all your data seamlessly stored in FerretDB.

   ![HeyForm Form Dashboard](/img/blog/heyform-dashboard.png)

## Example of HeyForm with data in FerretDB

The below image shows a simple form created in HeyForm, demonstrating its intuitive interface and ease of use.

![HeyForm Form Example](/img/blog/heyform-example.png)

After setting up HeyForm and creating a new form, you can inspect how HeyForm stores its data within FerretDB.

Connect to your FerretDB instance using `mongosh` by running the following command in your terminal with the appropriate username and password:

```sh
mongosh mongodb://<username>:<password>@localhost:27017/heyform
```

Explore the `heyform` database to see the collections and documents created by HeyForm.

```text
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

The form you created in HeyForm is stored in the `formmodels` collection.
Query the `formmodels` collection by running `db.formmodels.findOne()` to see the details of the form created:

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
  name: 'ferretdb-form',
  createdAt: ISODate('2025-07-01T19:17:43.272Z'),
  updatedAt: ISODate('2025-07-01T19:50:09.958Z'),
  __v: 0,
  fieldUpdateAt: 1751399409
}
```

This output demonstrates that HeyForm successfully replaces MongoDB with FerretDB, which in turn stores its data efficiently in PostgreSQL.

## Conclusion

The integration of HeyForm and FerretDB provides a robust, scalable, and fully open-source solution for building and managing online forms.
By leveraging FerretDB, you can run your entire workloads in open source, without vendor lock-in or restrictive licenses.

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
