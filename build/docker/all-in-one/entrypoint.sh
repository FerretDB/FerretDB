#!/bin/bash

set -ex

# Don't use `exec` so ctrl+c / SIGINT stops bash and not runsvdir (which ignores SIGINT).
# That's a hack, but good enough for all-in-one container.
/usr/bin/runsvdir /etc/service
