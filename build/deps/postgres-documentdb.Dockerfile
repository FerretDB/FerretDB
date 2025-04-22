# Use production image and full tag close to the release.
# FROM ghcr.io/ferretdb/postgres-documentdb:17-0.103.0-ferretdb-2.2.0

# Use moving development image during development.
# FROM ghcr.io/ferretdb/postgres-documentdb-dev:17-ferretdb

# FIXME
FROM ghcr.io/ferretdb/postgres-documentdb-dev:17-pr-toggles
