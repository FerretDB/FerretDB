# Application test results

| Application | Version | Checks | Result |
| --- | --- | --- | --- |
| UniFi Network | 10.6.109 | Configuration APIs; rule and alert APIs | 55/55 feature checks and 28/28 rule/alert checks passed. |
| UniFi Network | 10.6.109 | Populated reports and native statistics | 15/15 report checks and 66/66 statistics checks passed. |
| UniFi Network | 10.6.109 | Encrypted AP informs and client traffic | Normal-load windows completed 240 API reads and 20 informs each, with two synthetic APs and 20 clients. |
| UniFi Network | 10.6.109 | MODERN persistence; encrypted backup/restore | Selection survived restarts and container recreation; all four native tables restored exactly. |
| UniFi Network | 10.6.109 | Retention and TTL | 12/12 automatic-retention checks passed; scheduled TTL removed 10,002 expired rows per tested backend. |
| UniFi OS Server | 5.1.42, Network 10.6.109 | Console login, proxied APIs, configuration, encrypted informs and clients | 18/18 workload checks passed, including one synthetic adopted AP and one populated client. |
| UniFi OS Server | 5.1.42, Network 10.6.109 | Native statistics and restart persistence | 20 schemas available; populated MODERN summary returned; Network and full OS restart checks passed. |
| UniFi OS Server | 5.1.42, Network 10.6.109 | Encrypted backup/restore and failure handling | All four native tables restored exactly; deleted/changed/added rows and configuration canaries matched the backup. Failed exports withheld incomplete backups. |
| UniFi OS Server | 5.1.42, Network 10.6.109 | External database startup | Full OS restart completed with MODERN selected and zero MongoDB processes. |
| LibreChat | v0.8.9 | Login, streamed chat, follow-ups, edits, history, organization, import and CRUD | 22/22 application checks passed, including history after restart. |
| LibreChat | v0.8.9 | Index creation and scheduled expiry | 77 collections, 491 indexes and 26 TTL indexes; three expired fixtures deleted and six control rows preserved. |
| PostgreSQL/DocumentDB | 17 / 0.107 | Durability and final transaction checks | fsync, full-page writes and synchronous commit enabled; zero deadlocks and zero idle transactions. |
| PostgreSQL/DocumentDB | 17 / 0.107 | Fresh logical recovery | 119 BSON tables, catalog metadata and physical indexes verified. |
| FerretDB gateway | 2.7.0 with these changes | CPU under unattended / normal UniFi traffic | 1.883% → 0.071% unattended; 2.028% → 0.229% normal. |
| UniFi OS / LibreChat | Versions above | Final service logs | Zero errors. |
| UniFi Network | 10.6.109 | Operational use | Trouble-free alongside the automated checks. |

CPU percentages use one core as 100%. CPU samples cover the same synthetic workload, with 600-second before and 300-second after windows. The CPU comparison measures the polling and epoll implementations of disconnect detection.


## Settled idle writes

Original pg_cron: 188 KiB over 330 seconds. Patched settled confirmation: zero block writes, database-file write syscalls and new WAL over 90 seconds, with maintenance scans continuing. Queued index build, restart persistence, scheduled TTL and enabled history checks passed. [Details](IDLE_WRITES.md).

## Deployment verification

The pg_cron fix retained MODERN and passed scheduler-write, queued-index, scheduled-TTL, device-inform, native-statistics and database-safety checks. A fresh whole-cluster backup and its local copy were verified; database rollback is retained. [Aggregate results](pg_cron/live-verification.json).
