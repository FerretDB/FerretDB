#!/bin/bash

set -e

if [ -z "$POSTGRES_PASSWORD" ]; then
    echo "Error: POSTGRES_PASSWORD must be set. See https://docs.ferretdb.io/installation/evaluation/"
    exit 1
fi

if [ "${POSTGRES_DB:-postgres}" != "postgres" ]; then
    echo "Error: POSTGRES_DB must be set to 'postgres' or unset. See https://docs.ferretdb.io/installation/evaluation/"
    exit 1
fi

# prevent unset POSTGRES_DB using the value of POSTGRES_USER, see https://hub.docker.com/_/postgres
export POSTGRES_DB="postgres"

export FERRETDB_POSTGRESQL_URL="postgresql://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD}@127.0.0.1:5432/${POSTGRES_DB}"

exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
