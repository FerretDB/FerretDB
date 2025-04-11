#!/bin/bash

set -ex

echo "Waiting for Postgresql to finish..."

/usr/bin/sv term /etc/service/postgresql
