# Settled PostgreSQL idle writes

The remaining idle writes came from pg_cron's unused run-ID allocation. With `cron.log_run=off`, its launcher still advanced `cron.runid_seq` on every execution. Two DocumentDB index jobs run every two seconds. The sequence updates generated WAL and commits, which also caused later checkpoint writes.

## Attribution

The original library wrote **192,512 block bytes in 330 seconds** after the backend settled. Payload-free syscall tracing identified:

- `pg_wal/<segment>`: sequence WAL, commits, standby snapshots and checkpoint records.
- The `cron.runid_seq` relation: the unused run counter.
- `pg_xact/<segment>`: transaction commit status.
- `global/pg_control`: checkpoint state.
- `pg_logical/replorigin_checkpoint.tmp`: the checkpoint's small replication-origin file.

The WAL decode contained only sequence, transaction, standby and checkpoint records during this idle observation. Syscall bytes describe write requests; cgroup block bytes describe storage-accounted writes.

## Fix

The pg_cron 1.6.7 patch is one line:

```c
task->runId = CronLogRun ? NextRunId() : 0;
```

Run-history logging retains the original allocation when enabled. Job IDs, schedules, execution, TTL, index maintenance and PostgreSQL durability settings are unchanged. PostgreSQL/DocumentDB and FerretDB source are unchanged by this fix.

## Results

- Settled confirmation: **zero block writes, zero database-file write syscalls and zero new WAL over 90 seconds**.
- The preceding patched 330-second trace also recorded zero database-file writes and zero WAL. Its initial block-counter burst is retained separately in the JSON receipt.
- Index-queue scans increased by **656** in that preceding trace and **184** in the settled confirmation, while the run-ID sequence stayed unchanged.
- A real queued background index was built by the unchanged scheduled worker. Committed rows and the index survived a PostgreSQL restart.
- Scheduled TTL deleted three expired token/session/file fixtures and preserved six future/missing-expiry controls.
- With history enabled, successful and failed jobs retained distinct positive run IDs. This was checked in a separate background-worker instance.
- fsync, full-page writes and synchronous commit remained enabled. 

## Build

The pinned Containerfile, one-line patch and instructions are in [pg_cron/](pg_cron/README.md).

[Aggregate results](idle-write-results.json)
