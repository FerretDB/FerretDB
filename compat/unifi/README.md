# UniFi native statistics bridge

| Package | Implementation input | Native statistics backend |
| --- | --- | --- |
| Standalone Network 10.6.109 | `internal-dependencies.jar` | Configured PostgreSQL database |
| Native OS Network 10.6.109 | `ace.jar` with nested implementation/JDBC JARs | OS PostgreSQL 14 `unifi-network` database |

## Build

```sh
python build_bridge.py --packaging os --application-jar /path/to/ace.jar --asm /path/to/asm-9.9.jar --output ./build
```

For standalone packaging, select `--packaging standalone` and supply `internal-dependencies.jar`. Use `--javac` to select a compiler. The builder verifies the application and ASM hashes. The output contains our bridge and ASM. Vendor implementation classes are excluded.

## Install

| Package | JVM options | Files |
| --- | --- | --- |
| Standalone | `-Dunifi.new.stat.framework.enabled=true -javaagent:/config/native-statistics-bridge.jar` | Bridge at the agent path; configured PostgreSQL driver under `/usr/lib/unifi/lib/local`. |
| OS | `-Dunifi.new.stat.framework.enabled=true -Dferretdb.native.externalMongo=true -javaagent:/usr/lib/unifi/ferretdb-native-bridge/native-statistics-bridge.jar` | Bridge and emitted JDBC JAR under `/usr/lib/unifi/ferretdb-native-bridge`. |

OS configuration uses a final systemd environment file/drop-in to retain the JVM options and `UNIFI_MONGODB_SERVICE_ENABLED=false`. Disable/mask the unused `mongodb.service` and `unifi-mongodb.service` after configuring Network for the external backend. The lifecycle hook checks `db.mongo.local=false` and both application/statistics Mongo URIs before skipping the local MongoDB process start. Normal driver connections still verify the external database.

| Configuration | Source |
| --- | --- |
| PostgreSQL host/port/database/user/password | `UNIFI_POSTGRESQL_HOST`, `UNIFI_POSTGRESQL_PORT`, `UNIFI_POSTGRESQL_DB_NAME`, `UNIFI_POSTGRESQL_ADMIN_USER`, `UNIFI_POSTGRESQL_ADMIN_PASSWORD` |
| OS defaults | Localhost, port 5432, database `unifi-network`, role `unifi_network_admin`, native loopback authentication. |
| State directory | Standalone `/config/data/ferretdb-native-statistics`; OS `/var/lib/unifi/ferretdb-native-statistics`. |

| Check | Result |
| --- | --- |
| Selection persistence | MODERN and explicit LEGACY survived Network restarts; MODERN survived OS restart. |
| Archive recovery | All four native tables restored exactly, including microseconds, null/Unicode keys and payload bytes. |
| Export failure | Failed native connection prevented incomplete backup publication. |
| Preference write failure | Active selection remained unchanged. |

[Full results](../TEST_RESULTS.md)
