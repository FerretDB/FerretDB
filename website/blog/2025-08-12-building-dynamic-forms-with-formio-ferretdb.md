---
slug: building-dynamic-forms-with-formio-ferretdb
title: 'Build Dynamic Forms with Form.io and FerretDB'
authors: [alex]
description: >
  Learn how to combine Form.io's powerful form-building capabilities with FerretDB, leveraging a robust PostgreSQL-backed database for your form data and submissions.
image: /img/blog/ferretdb-formio.jpg
tags: [compatible applications, open source, community]
---

![Build Dynamic Forms with Form.io and FerretDB](/img/blog/ferretdb-formio.jpg)

Forms are fundamental to almost every application, serving as the primary interface for collecting data, managing user inputs, and automating workflows.

<!--truncate-->

[Form.io](https://form.io/) provides a unique, full-stack form management platform, offering both a powerful form builder and a JSON-powered API backend.
The open-source Form.io Community Edition allows developers to self-host this powerful solution.

At [FerretDB](https://www.ferretdb.com/), we're dedicated to providing a truly open-source alternative to MongoDB, leveraging the reliability and power of PostgreSQL as its backend.

In this blog post, we're excited to explore how Form.io seamlessly integrates with FerretDB, offering a robust and self-hostable solution for all your dynamic form and data management needs.

## What is Form.io?

Form.io is a platform that offers a drag-and-drop form builder that simultaneously generates a powerful REST API and manages data submissions.
Key features of the Community Edition include:

- **Form builder & renderer:** Visually design complex forms that are automatically rendered in your web applications.
- **Auto-generated API:** Every form you build automatically gets a corresponding REST API endpoint for submission, retrieval, and management.
- **Data management:** Built-in capabilities for storing, viewing, and managing form submissions.
- **Open source & self-hostable:** Full control over your forms, data, and deployment environment.

## Why use Form.io with FerretDB?

Form.io relies on MongoDB as its primary database backend for storing all form definitions, submissions, user data, and application state.
Since FerretDB is designed to be an open-source alternative to MongoDB, it can serve as a drop-in replacement for Form.io's database.

- **Full open-source stack:** Create a complete, transparent, and controllable form management stack, from your form builder to your database, eliminating proprietary lock-in.
- **Simplified infrastructure:** If your existing data infrastructure is already based on PostgreSQL, integrating Form.io with FerretDB can streamline your database management and reduce operational overhead.
- **Scalability:** Leverage the scalability of both Form.io's API server and FerretDB's efficient handling of document data on a PostgreSQL backend.

## Connecting Form.io to FerretDB

Connecting a self-hosted Form.io Community Edition instance to your FerretDB instance is straightforward, as Form.io expects a MongoDB-compatible database.

Create a `docker-compose.yml` file with the following content to define the services for Form.io Community Edition and FerretDB:

```yaml
services:
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
      - ./ferretdb_data:/var/lib/postgresql/data
  formio-ce:
    image: formio/formio:rc
    container_name: formio-ce
    restart: always
    ports:
      - '3001:3001'
    environment:
      DEBUG: formio:*
      ROOT_EMAIL: <root-email>
      ROOT_PASSWORD: <root-password>
      NODE_CONFIG: |
        {
          "mongo": "mongodb://<username>:<password>@ferretdb:27017/formio-ce",
          "port": 3001,
          "jwt": {
            "secret": "<jwt-secret>"
          }
        }

networks:
  default:
    name: ferretdb
```

Replace `<username>`, `<password>`, `<root-email>`, and `<root-password>` with your desired credentials.
Also, replace `<jwt-secret>` with a secure secret for JWT authentication - this is used to sign JSON Web Tokens for secure user authentication in Form.io.

This setup defines two services: `ferretdb` for the FerretDB instance and `formio-ce` for the Form.io Community Edition application.
The `ferretdb` service uses the [FerretDB evaluation image](https://docs.ferretdb.io/installation/evaluation/), which is designed for quick testing and experiments.

Make sure you have Docker installed on your machine.
Then, run the following command in the directory where your `docker-compose.yml` file is located:

```sh
docker compose up -d
```

This command will start Form.io and FerretDB services in detached mode.
Then, open your browser and navigate to http://localhost:3001 to access the Form.io Community Edition interface.
Log in with `<root-email>` and the `<root-password>` you set in the `docker-compose.yml`.
You can now start building forms and managing data.
All data will be stored in FerretDB.

## Building a form with Form.io and FerretDB

Once Form.io Community Edition is running, you can access its intuitive form builder.
Let's create a simple contact form.

1. Log into your Form.io instance (http://localhost:3001).
2. Navigate to the "Forms" section and click "New Form".
3. Drag and drop some components like "Text Field", "Text Area", "Number", "Radio", "Checkbox", and "Button" to create a contact form.
4. Save the form.

This action will create a new form definition in Form.io, and behind the scenes, this definition is stored as a document in FerretDB.

Below is an example of how your form might look in the Form.io builder:

![Form.io Contact Form Builder](/img/blog/formio-contact-form.png)

Let's inspect how Form.io stores its data within FerretDB.
Connect to your FerretDB instance using `mongosh` or a GUI tool (like MongoDB Compass or DBeaver).

```sh
mongosh mongodb://<username>:<password>@localhost:27017/formio-ce
```

Connecting to the `formio-ce` database via Compass, let's query the `forms` collection to see the form we built earlier.

![Form.io data shown in FerretDB](/img/blog/formio-data.png)

Form.io writes and reads its data through FerretDB seamlessly, which in turn stores it efficiently in PostgreSQL, providing a reliable backend for your dynamic forms.

## Conclusion

The integration of Form.io and FerretDB provides a robust, scalable, and fully open-source solution for building and managing dynamic online forms.

- [Ready to get started? Try FerretDB today](https://github.com/FerretDB/FerretDB)
- [Have questions, suggestions, or requests? Join our community](https://docs.ferretdb.io/#community)
- [Discover more ways to integrate other compatible applications with FerretDB](https://docs.ferretdb.io/compatible-applications)
