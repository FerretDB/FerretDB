#!/bin/sh

set -e

git describe --tags --dirty > version.txt
git rev-parse HEAD > commit.txt
git branch --show-current > branch.txt
