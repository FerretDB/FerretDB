#!/bin/sh

set -ex

cd /

test -n "${TARGETOS}"
test -n "${TARGETARCH}"

tarball_name=gocache_${TARGETOS}_${TARGETARCH}.tar
tar cf ${tarball_name} gocache
find / -not -name ${tarball_name} \( -type f -o -type d \) -delete
