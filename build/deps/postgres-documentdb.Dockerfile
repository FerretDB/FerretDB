# Use production image and full tag close to the release.
# FROM ghcr.io/ferretdb/postgres-documentdb:17-0.104.0-ferretdb-2.3.0

# Use moving development image during development.
FROM ghcr.io/ferretdb/postgres-documentdb-dev:17-ferretdb
