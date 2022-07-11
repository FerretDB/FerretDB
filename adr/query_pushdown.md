Select queries of the form `{_id: <ObjectID>}`

## Postgres

Both work:

```sql
select * from test where (_jsonb->'_id')::jsonb->>'$o' = '507f1f77bcf86cd799439011';
```

```sql
select * from test where _jsonb->'_id' = '{"$o":"507f1f77bcf86cd799439011"}'::jsonb;
```

[PostgreSQL functions](https://www.postgresql.org/docs/14/functions-json.html)

## Tigris

```go
collection, err := db.DescribeCollection(ctx, param.collection)
if err != nil {
    return nil, lazyerrors.Error(err)
}

var schema tjson.Schema
if err = schema.Unmarshal(collection.Schema); err != nil {
    return nil, lazyerrors.Error(err)
}

objectID:= ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0x00, 0x00, 0x01}
primaryKey := schema.PrimaryKey // it is an array

it, err := db.Read(ctx, "coll1", driver.Filter(`{"<Primary Key>" : "<ObjectID>"}`), nil)
```