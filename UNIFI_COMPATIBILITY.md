# Application compatibility

Enable the server compatibility paths with `FERRETDB_UNIFI_COMPAT=true`.

| Setting | Behavior |
| --- | --- |
| `FERRETDB_UNIFI_COMPAT` | Enables the UniFi command, aggregation, capped-storage and namespace compatibility paths. |
| `FERRETDB_MAX_PERSISTENT_CURSORS` | Sets admission limits for cursors that retain PostgreSQL connections. |
| `FERRETDB_QUERY_SHAPES_PATH` | Optional output prefix for synthetic command-shape capture. |
| `FERRETDB_QUERY_SHAPES_SYNTHETIC_ONLY=true` | Required acknowledgement for shape capture. |
| `FERRETDB_QUERY_SHAPES_DURATION` | Capture duration, default five minutes and maximum ten minutes. |
| `FERRETDB_QUERY_SHAPES_PEER` | Optional exact client-address filter for synthetic capture. |

| Operation | Implementation |
| --- | --- |
| Durability probe | `serverStatus.wiredTiger.log.ferretdb.postgresqlDurableWAL` reports the actual PostgreSQL durability settings. Disabled/unknown settings produce zero. |
| Restore namespace moves | A PostgreSQL transaction moves collection metadata within one physical database, preserves tables/indexes and invalidates affected cursors. The catalog layout is pinned to DocumentDB 0.107. |
| Aggregation | Composite object keys preserve missing versus null members. Literal data and single-member key behavior are preserved. |
| Capped storage | Enforces document-count and BSON-byte limits, including conversion, concurrent writes and rollback. |
| Cancellation | Disconnects cancel database requests and connection-pool waits; Linux uses a shared epoll waiter. |
| Unacknowledged writes | `moreToCome` writes execute without a wire response. |
| Native statistics | The separate Java bridge persists source selection and includes the four native tables in encrypted application archives. |

| Tested application | Version |
| --- | --- |
| UniFi Network | 10.6.109 |
| UniFi OS Server | 5.1.42, native Network 10.6.109 |
| LibreChat | v0.8.9 |
| Backend | FerretDB 2.7.0, PostgreSQL 17, DocumentDB 0.107 |

[Application results](compat/TEST_RESULTS.md) · [Machine-readable results](compat/test-results.json) · [Java bridge](compat/unifi/README.md) · [DocumentDB recovery](compat/DOCUMENTDB_RECOVERY.md)
