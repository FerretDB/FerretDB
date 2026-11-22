# DocumentDB logical recovery

| Item | Procedure |
| --- | --- |
| Tested versions | PostgreSQL 17 and DocumentDB 0.107. |
| Application writers | Quiesce writers during backup/recovery. |
| Snapshot | Hold an exported repeatable-read snapshot for the dump and catalog exports. |
| Physical data | Run `pg_dump -Fc` against the exported snapshot. |
| Extension metadata | Export ordinary tables in `documentdb_api_catalog` and `documentdb_api_internal` with binary COPY from that snapshot. |
| Maintenance jobs | Export `cron.job`; handle extension-created jobs separately during restore to avoid duplicate IDs. |
| Sequence state | Record `last_value` and `is_called` for catalog/internal/cron sequences. |
| Integrity | Record hashes, BSON-table fingerprints, collection/index catalog fingerprints and physical index definitions. |
| Restore | Create a separate backend with the pinned extension; restore pre-data, physical data and explicit catalog data, then indexes and sequence state. |
| Capped metadata | Include the `ferretdb_unifi` tables added by these changes and ordering state. |
| Native UniFi statistics | Include their separate PostgreSQL database or the bridge's application archive. |
| Roles | Preserve required roles/grants separately from the database dump. |
| Verification | Compare data/catalog/index fingerprints, then check namespace reads and creation/write of a new collection. |

| Recovery test | Result |
| --- | --- |
| Physical BSON tables | 119 verified. |
| Catalog metadata | Collection/index fingerprints matched. |
| Physical indexes | Definitions matched. |
| Namespace use | Reads and subsequent creation/writes passed. |

PostgreSQL and `pg_dump` binaries are unchanged. The procedure adds the extension-owned catalog data omitted by the plain logical dump.
