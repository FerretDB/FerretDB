# Application migration instructions

This file governs work in this directory. To use it for a migration elsewhere, explicitly ask the agent to read and follow this file and README.md before operating on that deployment.

## Scope

Help the user move LibreChat, UniFi Network or UniFi OS Server from MongoDB to a compatible FerretDB/PostgreSQL/DocumentDB stack. Preserve application data, credentials, controller identity and a working rollback. Use the current contribution or a released upstream version only after checking that it includes the required fixes. Read ../TEST_RESULTS.md and ../../UNIFI_COMPATIBILITY.md.

## Discovery and preparation

- Read existing infrastructure documentation and applicable AGENTS.md files. Inspect versions, runtime, service definitions, database names, storage, networks and backup procedures with read-only checks. Keep credentials and identifying values out of output.
- Use an isolated PostgreSQL instance and new volumes. Do not initialize DocumentDB inside an unrelated PostgreSQL instance or alter its pg_cron settings.
- Use explicit image digests or recorded immutable image IDs. Check source capability and DocumentDB catalog compatibility, not just an image tag. The recorded test stack uses PostgreSQL 17 and DocumentDB 0.107; newer combinations need their own checks.
- Make a resource manifest of exact names, IDs, ownership labels, image IDs, mounts and ports. Stop on a name collision. No general prune or recursive cleanup outside the owned paths.
- Keep passwords, connection strings, backups, dumps, inventories and raw logs outside the repository with restrictive permissions. Do not publish data or contact external services with it.
- Rehearse using synthetic data first. A copied real-data rehearsal needs explicit user authorization and isolation from real devices, notifications, webhooks and external actions.

## Migration

- Preparing a migration does not authorize stopping production. Obtain explicit authorization for the concrete cutover and outage, unless already provided in the active task. Continue authorized preparation without repeated permission questions.
- Before cutover, take a fresh consistent MongoDB backup and application configuration backup. Verify restoration into an isolated target. Preserve database names, collection options, BSON types, indexes, TTL definitions, users/roles where applicable and all application databases. Do not silently skip failed indexes or records.
- Keep the application version unchanged during the database move. Stop all writers for the final copy. Use tools compatible with the source MongoDB version; modern Database Tools may not support old servers.
- Change only the application's database configuration and required network attachment. Preserve UniFi controller files, site/device identity, inform address, certificates, encryption keys and ports. Do not reset or re-adopt devices to compensate for a failed migration.
- Native UniFi statistics have a separate PostgreSQL database and optional bridge. Preserve it and the source selection. A MongoDB dump does not include those statistics. Check application build pins and third-party notices before installing/building the bridge.

## Acceptance and rollback

- Compare source and restored collection/document counts and index definitions, accounting explicitly for TTL expiry. Verify critical BSON types and representative reads/writes. Validate application login, persistence across restart and the feature checks in README.md.
- Check backend/controller logs, informs, native statistics if enabled, scheduled TTL, queued index creation, cancellation, durability, lock waits, deadlocks and idle transactions. Measure settled idle CPU/writes without changing maintenance schedules or durability to improve results.
- Keep MongoDB, its backup and the original configuration available. Leave the application on the new backend when checks pass and the user authorized that outcome. Do not roll back a healthy deployment merely because a test window ended.
- If rollback is needed, stop writers first. Identify writes made since cutover and agree how to preserve/reconcile them before pointing back at stale MongoDB. Report the data boundary plainly.
- Remove only manifest-listed temporary resources after their role is complete. Keep backups and rollback resources until the user authorizes removal. Report verification results and remaining resources without private identifiers.

## Publication

Prepare local drafts by default. Check all proposed content for credentials, personal data, local paths, device/network identities, run timestamps and private history. Publishing or editing a remote repository requires approval for the exact content and destination. Do not sign contributor agreements on the user's behalf.
