---
sidebar_position: 9
---

# Data API

FerretDB Data API is an open source alternative to MongoDB Atlas Data API.
It lets you perform MongoDB operations via HTTP requests without needing a MongoDB driver.
However, it is built directly into FerretDB and not a standalone service.

## Enable the Data API

To access Data API on FerretDB, set the environment variable or flag (`FERRETDB_LISTEN_DATA_API_ADDR`/`--listen-data-api-addr`) to the desired address and port when starting FerretDB.
The default address is `:8080`.

For example, to run FerretDB locally with the Data API on port `8080`, you can use:

```text
--listen-data-api-addr=:8080
```

Or:

```text
FERRETDB_LISTEN_DATA_API_ADDR=:8080
```

The Data API will be accessible at `http://localhost:8080`.
Please note that you need to provide your authentication credential in the request headers or as part of the URL if authentication is enabled.

## Using the Data API

The Data API supports standard MongoDB operations like `insert`, `find`, `update`, and `delete`.
It follows the [Data API OpenAPI documentation defined here](https://github.com/FerretDB/FerretDB/blob/main/internal/dataapi/api/openapi.json).

### Insert a document

To insert a single document, use the `/action/insertOne` endpoint.

```sh
curl -X POST http://localhost:8080/action/insertOne \
  -H "Content-Type: application/json" \
  -u <username>:<password> \
  -d '{
        "database": "db",
        "collection": "books",
        "document": {
          "_id": "pride_prejudice_1813",
          "name": "Pride and Prejudice",
          "authors": [{ "name": "Jane Austen", "nationality": "British" }],
          "publication": {
            "date": "1813-01-28T00:00:00Z",
            "publisher-name": "T. Egerton"
          }
        }
      }'
```

### Find a document

To find documents, use the `/action/find` endpoint.

```sh
curl -X POST http://localhost:8080/action/find \
  -H "Content-Type: application/json" \
  -u <username>:<password> \
  -d '{
        "database": "db",
        "collection": "books",
        "filter": { "_id": "pride_prejudice_1813" }
      }'
```

### Update a document

To update a document, use the `/action/updateOne` endpoint.

```sh
curl -X POST http://localhost:8080/action/updateOne \
  -H "Content-Type: application/json" \
  -u <username>:<password> \
  -d '{
        "database": "db",
        "collection": "books",
        "filter": { "_id": "pride_prejudice_1813" },
        "update": { "$set": { "isbn": "9780141439518" } }
      }'
```

### Delete a document

To delete a document, use the `/action/deleteOne` endpoint.

```sh
curl -X POST http://localhost:8080/action/deleteOne \
  -H "Content-Type: application/json" \
  -u <username>:<password> \
  -d '{
        "database": "db",
        "collection": "books",
        "filter": { "_id": "pride_prejudice_1813" }
      }'
```

## Import Data API OpenAPI Spec into API clients

The FerretDB Data API is compatible with OpenAPI 3.0, allowing you to import the API specification into various API clients like Postman, Insomnia, or Swagger UI.
Below is an example using Postman.

Import the FerretDB Data API spec directly into Postman via the following URL:

- `https://raw.githubusercontent.com/FerretDB/FerretDB/main/internal/dataapi/api/openapi.json`

This will load available endpoints into Postman, as shown below:

![Import Data API Specification into Postman](/img/docs/import-data-api.jpg)

Ensure to set the `{{baseURL}}` as `http://localhost:8080` or the listen address you configured for the Data API.

Using the same examples from earlier, you can now test the endpoints directly in Postman.
You can also set up environment variables in Postman for the `username` and `password` to avoid hardcoding them in the requests.

Below are examples of the `insertOne` and `find` endpoints in Postman after importing the OpenAPI spec.

![Insert a single document](/img/docs/insert-one.jpg)
![Find a single document](/img/docs/find-one.jpg)
