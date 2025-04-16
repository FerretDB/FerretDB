---
sidebar_position: 7
---

# Data API

FerretDB provides a built-in HTTP Data API for interacting with your database using standard REST calls.
This lets you perform MongoDB-style operations over HTTP without needing a MongoDB driver.

## Enable the Data API

To enable the Data API, set the `FERRETDB_LISTEN_DATA_API_ADDR` environment variable or the `--listen-data-api-addr` command-line flag to the desired address and port when starting FerretDB.
The default address is `:8080`.

```sh
--listen-data-api-addr=:8080
```

Or:

```sh
FERRETDB_LISTEN_DATA_API_ADDR=:8080
```

The Data API will be accessible at `http://localhost:8080`.

## Authentication

If you have auth enabled, set the basic Auth using your credentials:

```sh
-u <username>:<password>
```

## Supported endpoints

| Endpoint               | Description            |
|------------------------|------------------------|
| `/action/insertOne`    | Insert one document    |
| `/action/find`         | Query documents        |
| `/action/updateOne`    | Update one document    |
| `/action/deleteOne`    | Delete one document    |

All requests are sent using `POST` with `Content-Type: application/json`.

## Examples

### Insert

To insert a document, use the `/action/insertOne` endpoint.

```sh
curl -X POST http://localhost:8080/action/insertOne \
  -H "Content-Type: application/json" \
  -u username:password \
  -d '{
        "database": "test",
        "collection": "users",
        "document": {
          "name": "Andrew",
          "email": "andrew@example.com",
          "age": 25
        }
      }'
```

### Find

To find documents, use the `/action/find` endpoint.

```sh
curl -X POST http://localhost:8080/action/find \
  -H "Content-Type: application/json" \
  -u username:password \
  -d '{
        "database": "test",
        "collection": "users",
        "filter": { "name": "Andrew" }
      }'
```

### Update

To update a document, use the `/action/updateOne` endpoint.

```sh
curl -X POST http://localhost:8080/action/updateOne \
  -H "Content-Type: application/json" \
  -u username:password \
  -d '{
        "database": "test",
        "collection": "users",
        "filter": { "name": "Andrew" },
        "update": { "$set": { "email": "andrew.new@example.com" } }
      }'
```

### Delete

To delete a document, use the `/action/deleteOne` endpoint.

```sh
curl -X POST http://localhost:8080/action/deleteOne \
  -H "Content-Type: application/json" \
  -u username:password \
  -d '{
        "database": "test",
        "collection": "users",
        "filter": { "name": "Andrew" }
      }'
```

## Using with Postman

1. Open Postman and click **Import**.
2. Paste the URL `https://raw.githubusercontent.com/FerretDB/FerretDB/main/internal/dataapi/api/openapi.json` or upload it from file.
3. After import, go to the collection and configure Basic Auth under **Authorization**.
4. Send requests using the built-in endpoints.

### Sample `find` request

- Method: `POST`
- URL: `http://localhost:8080/action/find`
- Authorization: Basic Auth
- Body (raw, JSON):

```json
{
  "database": "test",
  "collection": "users",
  "filter": { "name": "Andrew" }
}
```

You now have a working setup for querying FerretDB over HTTP.
