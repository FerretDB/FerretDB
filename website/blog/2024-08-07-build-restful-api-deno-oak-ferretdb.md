---
slug: build-restful-api-deno-oak-ferretdb
title: 'How to Build a RESTful API with Deno, Oak, and FerretDB'
authors: [alex]
description: >
  In this blog post, we'll guide you through setting up a RESTful API using Deno, Oak, and FerretDB.
image: /img/blog/ferretdb-deno.jpg
tags: [tutorial, community, open source]
---

![Run FerretDB ](/img/blog/ferretdb-deno.jpg)

What if you could overcome Node.js's limitations and build a RESTful API with enhanced security and simplicity?

<!--truncate-->

[Node.js](https://nodejs.org/en) has been the most popular server-side Javascript runtime environment for quite some time now.
For many developers, it has become a crucial part of the popularized web development stacks such as MEAN (MongoDB, Express, Angular, and Node.js) and MERN (MongoDB, Express, React, and Node.js).

However, Node.js is not without its limitations – a centralized module system, no built-in support for Typescript, complex dependency management, and poor security.

To address these issues, the creators of Node.js developed [Deno](https://deno.com/).
It comes with built-in TypeScript support, a more secure runtime that requires explicit permission to access files, networks, and environments, and a simplified module system that uses URL imports instead of centralized packages.

In this blog post, we'll guide you through setting up a RESTful API using Deno, Oak, and FerretDB for data storage.

## Deno, Oak, and FerretDB: A new alternative stack?

Unlike the typical Node.js-based stacks, Deno, Oak, and FerretDB together offer a modern alternative framework for web development.

Instead of Express, you can use [Oak](https://deno.land/x/oak) – a middleware framework built specifically for Deno – for building web applications and APIs and can take full advantage of its features and ecosystem without additional configuration.

[FerretDB](https://www.ferretdb.com/), a truly open source alternative to MongoDB built on Postgres, offers you flexibility and freedom without limiting you in any way – no need to worry about vendor lock-in associated with proprietary solutions[](https://blog.ferretdb.io/5-ways-to-avoid-database-vendor-lock-in/).

This combination of Deno, Oak, and FerretDB offers a powerful, secure, and modern framework for building scalable web applications and APIs, which makes it a strong alternative to traditional Node.js-based stacks.

## Prerequisites

- FerretDB is installed and running locally: If you don't currently have FerretDB running, [follow this quickstart guide to set up an instance](https://docs.ferretdb.io/quickstart-guide/).

## Install Deno locally

We will start by installing Deno on our local machine.
The following command installs Deno on Linux/MacOS:

```sh
curl -fsSL https://deno.land/install.sh | sh
```

After installation, follow the installation and manually add the Deno binary to your system path.
Run `deno --version` to confirm the installation.

```text
$ deno --version
deno 1.45.5 (release, aarch64-apple-darwin)
v8 12.7.224.13
typescript 5.5.2
```

### Building the application

Let's start by creating a directory for the application.
From your terminal, run the following command:

```sh
mkdir sample_app && cd sample_app
```

#### Set up the dependencies

Unlike Node.js, Deno does not require a `package.json` or a package manager like npm.
Instead, dependencies are imported directly via URLs.
We can start by creating a file to manage all the dependencies.

Create a `deps.ts` file that exports the `Application` and `Router` classes from the Oak framework, and the `MongoClient` class from the MongoDB driver for use in other parts of our application.

```typescript
export { Application, Router } from 'https://deno.land/x/oak@v10.5.1/mod.ts'
export { MongoClient, ObjectId } from 'https://deno.land/x/mongo@v0.32.0/mod.ts'
```

#### Set up database connection

Since we are setting up Deno with FerretDB, we should create a database connection for our application.

Create a `db.ts` file for database connection:

```typescript
import { MongoClient, ObjectId } from './deps.ts'

const client = new MongoClient()
await client.connect('<FerretDB_connection_URI>')

const db = client.database('library')
export const books = db.collection<BookSchema>('books')

interface BookSchema {
  _id?: ObjectId
  title: string
  author: string
  genre: string
}
```

In the code above, we set up a connection to the FerretDB database using the MongoDB client imported from the `deps.ts` file.
We also define a `BookSchema` interface to structure our data.

Ensure to replace `<FerretDB_connection_URI>` with the connection URI to your FerretDB instance.

#### Create the main server file

Create an `server.ts` file for your server:

```typescript
import { Application, Router } from './deps.ts'
import { books } from './db.ts'
import { ObjectId } from 'https://deno.land/x/mongo@v0.32.0/mod.ts'

const app = new Application()
const router = new Router()
const PORT = 3000

router
  .get('/', (context) => {
    context.response.body = { message: 'Hello from a Deno API!' }
  })
  .get('/api/books', async (context) => {
    const allBooks = await books.find().toArray()
    context.response.body = allBooks
  })
  .get('/api/books/:id', async (context) => {
    const id = context.params.id
    if (id) {
      const book = await books.findOne({ _id: new ObjectId(id) })
      if (book) {
        context.response.body = book
      } else {
        context.response.status = 404
        context.response.body = { message: 'Book not found' }
      }
    } else {
      context.response.status = 400
      context.response.body = { message: 'Invalid book ID' }
    }
  })
  .post('/api/books', async (context) => {
    const body = await context.request.body().value
    const insertId = await books.insertOne(body)
    context.response.body = { id: insertId }
  })
  .patch('/api/books/:id', async (context) => {
    const id = context.params.id
    if (id) {
      const body = await context.request.body().value
      await books.updateOne({ _id: new ObjectId(id) }, { $set: body })
      context.response.body = { message: 'Book updated' }
    } else {
      context.response.status = 400
      context.response.body = { message: 'Invalid book ID' }
    }
  })
  .delete('/api/books/:id', async (context) => {
    const id = context.params.id
    if (id) {
      await books.deleteOne({ _id: new ObjectId(id) })
      context.response.body = { message: 'Book deleted' }
    } else {
      context.response.status = 400
      context.response.body = { message: 'Invalid book ID' }
    }
  })

app.use(router.routes())
app.use(router.allowedMethods())

console.log(`Server running at http://localhost:${PORT}`)
await app.listen({ port: PORT })
```

Each route is set up with a path and a callback function to handle HTTP requests.
The first route handles `GET` requests at the root URL (`"/"`), and then sends a simple JSON message back to the client.

The next route handles `GET` requests at `"/api/books"`, and fetches all book records from the database and returns them to the client.

For `GET` requests at `"/api/books/:id"`, we retrieve a specific book by its `ObjectId`, returning the book if found, or a `404` error if not.
The `POST` route at `"/api/books"` reads the request body to insert a new book into the database and returns the new book's ID.

The `PATCH` route updates the specified book's details if the ID is valid and sends back a confirmation message.

The `DELETE` route removes the specified book from the database and provides a success response message or an error if the ID is invalid.

#### 6. Run the server

Run the server with the necessary permissions:

```sh
deno run --allow-net --allow-read --allow-write index.ts
```

Once the application is up and running, the endpoint `http://localhost:3000/` should be accessible.

We will setup a connection in Postman to test if the Deno server is running and accessible.
If it is, you should get this response from the API:

```json
{ "message": "Hello from a Deno API!" }
```

A screenshot of the endpoint's output in Postman can be seen below.

![Deno connection via Postman](/img/blog/ferretdb-deno/deno-connection.png)

#### Inserting a new book

Below is an example of a `POST` request made to the API endpoint `http://localhost:3000/api/books`.
Here, we are inserting one database record into our FerretDB database.

```json
{
  "title": "The Great Gatsby",
  "author": "F. Scott Fitzgerald",
  "genre": "Classic"
}
```

![Insert one record into database](/img/blog/ferretdb-deno/insert-record.png)

#### Get all the records

Next, we will make a `GET` request to retrieve all book records from the database.
Since there's only one record at this point, the expected response should be the data we just inserted.

![Get all records from the database](/img/blog/ferretdb-deno/get-records.png)

#### Update record by ObjectId

Since we are not working with a fixed schema or data type, we can update the genre for the data record to an array.
That way, we can have it cover more than a singular genre.

We will use a `PATCH` request to update the `genre` field of the record.

```json
{
  "genre": ["tragedy", "classic"]
}
```

![Update record by ObjectId](/img/blog/ferretdb-deno/update-record.png)

#### Delete record by ObjectId

Finally, we will delete the record by its `ObjectId` using a `DELETE` request.

![Delete record by ObjectId](/img/blog/ferretdb-deno/delete-record.png)

## Conclusion

As we step back from this interesting setup, we've not only built a functional API but also provided an alternative approach to building web applications.
Deno's built-in TypeScript support and security features offer a modern, efficient alternative for API development.

Deno, Oak and FerretDB (FORD or FOAD stack?) is a viable alternative stack for many developers to build full-stack web applications.

So if you're looking to try out new tools or enhance your existing workflow, we encourage you to experiment with Deno, Oak and FerretDB and let us know what you think.

Feel free to reach out to us on our [community channels](https://docs.ferretdb.io/#community) with your thoughts, feedback, and questions.
