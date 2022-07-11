Select queries of the form `{_id: <ObjectID>}`

## Postgres

```sql
select * from test where (_jsonb->'_id')::jsonb->>'$o' = '507f1f77bcf86cd799439011';
```

```sql
select * from test where _jsonb->'_id' = '{"$o":"507f1f77bcf86cd799439011"}'::jsonb;
```

## Tigris

TODO: How will we ensure that Tigris compares values the same way as MongoDB?