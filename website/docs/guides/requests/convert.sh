#!/usr/bin/env bash

for file in *.js; do
  sed -Ei '1s/^db\.runCommand\(\{/{/' "$file"
  sed -Ei '$s/\}\)/}/' "$file"
  sed -Ei 's/([[:space:]]*)([A-Za-z0-9_]+):/\1"\2":/g' "$file"
  sed -Ei "s/'/\"/g" "$file"
done
