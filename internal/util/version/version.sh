#!/bin/sh

set -ex

git describe --tags --dirty > version.txt
