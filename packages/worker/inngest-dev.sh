#!/usr/bin/env bash

set -ex

# As of 2024-06-26, `npx inngest-cli@latest dev` (1) only works with npx and 
# (2) `bun run <package.json script>` does weird things if the script references npx
# so put the npx command here in this shell script so we can have a package.json script like
# "dev:inngest": "./inngest-dev.sh"
# and use `bun run dev:inngest`
exec npx inngest-cli@latest dev