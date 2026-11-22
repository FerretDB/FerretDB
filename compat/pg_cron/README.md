# Avoid unused pg_cron run-ID writes

pg_cron 1.6.7 calls `NextRunId()` for every scheduled execution before checking `cron.log_run`. With history disabled, those IDs are unused, but advancing `cron.runid_seq` still produces WAL and commit records. Frequent empty index-maintenance jobs therefore keep generating disk activity.

The patch changes that allocation to:

```c
task->runId = CronLogRun ? NextRunId() : 0;
```

When history is enabled, allocation and audit records retain their existing behavior. Scheduling, job IDs, job execution, TTL, index builds and durability settings are unchanged.

## Build

```sh
podman build -t postgres-documentdb:cron-idle-fix compat/pg_cron
```

The Containerfile uses the pinned PostgreSQL 17 / DocumentDB 0.107 image and verifies the pg_cron 1.6.7 source archive checksum. Only the rebuilt `pg_cron.so` is copied into the final image. PostgreSQL, DocumentDB and their settings are unchanged.

The patch takes effect when PostgreSQL starts with that library. Use the existing `cron.log_run=off` and `cron.log_statement=off` settings for quiet operation. Turning history on intentionally retains its database writes.

## Verification

The accompanying results cover settled idle syscall/block-write attribution, active maintenance scans, a queued background index build, restart persistence, scheduled TTL deletion and enabled run-history success/failure records. Source and runtime results are recorded in the idle-write investigation report.

[pg_cron source](https://github.com/citusdata/pg_cron/blob/v1.6.7/src/pg_cron.c)

## Deployment verification

The image was deployed with a verified whole-cluster backup and rollback retained. MODERN, queued index creation, scheduled expiry and existing device informs passed. Scheduler writes and unused run-ID advancement remained zero during a 90-second observation under normal application traffic. [Results](live-verification.json).
