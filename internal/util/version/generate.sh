#!/bin/sh

set -e

git describe --tags --dirty > version.txt
git branch --show-current > branch.txt
