#!/bin/bash

tmp="${GITHUB_TMP:=$(mktemp -d)}"

# Dump the logs of all engine containers
# (There should only be one, but just in case)
mapfile -t containers < <(docker ps --filter name="dagger-engine-*" -q)
if [[ "${#containers[@]}" -gt 0 ]]; then
  for container in "${containers[@]}"; do
    docker logs "$container" &> "$tmp/engine-$container.log"
  done
fi
