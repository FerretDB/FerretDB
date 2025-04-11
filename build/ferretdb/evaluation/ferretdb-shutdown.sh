#!/bin/bash

set -ex

echo "Waiting for FerretDB to finish..."

/usr/bin/sv term /etc/service/ferretdb
