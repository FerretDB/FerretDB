#!/bin/bash

set -e

test -n "$POSTGRES_PASSWORD" || (echo "cannot start FerretDB: POSTGRES_PASSWORD must be set" && false)
test "${POSTGRES_DB:-postgres}" = "postgres" || (echo "cannot start FerretDB: POSTGRES_DB must be set to 'postgres' or unset" && false)

export FERRETDB_POSTGRESQL_URL="postgresql://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD}@127.0.0.1:5432/postgres"

exec /usr/local/bin/ferretdb
