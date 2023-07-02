#!/bin/sh

set -ex

cd /

# download Go modules into Docker cache directory that may or may not be empty
export GOMODCACHE=/gomodcache-docker
cp /host/go.mod /host/go.sum .
go mod verify
go mod download -x

# copy them to directory that is not a Docker cache and cannot be cleaned up on a whim by it
cp -R /gomodcache-docker /gomodcache

# import build cache from the host that may be absent
mkdir /gocache
tar xf /host/tmp/docker/gocache/gocache.tar || true
