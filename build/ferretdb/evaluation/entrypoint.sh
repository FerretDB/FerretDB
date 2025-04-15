#!/bin/bash

set -e

test -n "$POSTGRES_USER" || (echo "POSTGRES_USER must be set" && false)
test -n "$POSTGRES_PASSWORD" || (echo "POSTGRES_PASSWORD must be set" && false)

export FERRETDB_POSTGRESQL_URL="postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@127.0.0.1:5432/postgres"

exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
