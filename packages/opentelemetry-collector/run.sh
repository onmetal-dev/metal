#!/usr/bin/env bash
set -ex
docker-compose down
exec docker-compose up > /dev/null 2>&1