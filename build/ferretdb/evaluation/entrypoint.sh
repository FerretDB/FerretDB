#!/bin/bash

set -e

if [ -z "$POSTGRES_PASSWORD" ]; then
    echo "Error: POSTGRES_PASSWORD must be set"
    exit 1
fi

if [ "${POSTGRES_DB:-postgres}" != "postgres" ]; then
    echo "Error: POSTGRES_DB must be set to 'postgres' or unset"
    exit 1
fi

# explicitly set POSTGRES_DB, because if POSTGRES_DB is unset it uses POSTGRES_USER value
# see https://hub.docker.com/_/postgres
export POSTGRES_DB="postgres"

export FERRETDB_POSTGRESQL_URL="postgresql://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD}@127.0.0.1:5432/postgres"

exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
