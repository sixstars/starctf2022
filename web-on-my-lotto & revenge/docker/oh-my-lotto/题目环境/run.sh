#!/usr/bin/env bash
# docker-compose up
set -e
set -x

i=$1
COMPOSE_PROJECT_NAME="team${i}" \
PORT=$[53000+${i}] \
timeout 180 docker-compose up  --force-recreate

