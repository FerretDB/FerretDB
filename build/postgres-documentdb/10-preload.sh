#!/bin/bash

set -e

echo "shared_preload_libraries = 'pg_cron,pg_documentdb_core,pg_documentdb'" >> $PGDATA/postgresql.conf
echo "cron.database_name       = 'postgres'"                                 >> $PGDATA/postgresql.conf

source /usr/local/bin/docker-entrypoint.sh
docker_temp_server_stop
docker_temp_server_start "$@"
