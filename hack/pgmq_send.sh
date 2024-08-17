#!/usr/bin/env bash

PGPASSWORD=postgres psql -h localhost -p 5432 -U postgres -d metal -f hack/pgmq_send.sql
