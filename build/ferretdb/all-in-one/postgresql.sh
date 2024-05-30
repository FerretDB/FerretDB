#!/bin/bash

set -ex

exec /usr/local/bin/docker-entrypoint.sh postgres
