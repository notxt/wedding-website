#!/usr/bin/env bash
# Launch a shell in the dev container with the compose port mappings honored
# (so `go run ./cmd/server` inside the container is reachable at
# http://localhost:8080 from the host browser).
#
# Plain `docker compose run` ignores the `ports:` block in compose.yml;
# only `up` or `run --service-ports` map host ports.
#
# Optional args after `dev` are forwarded to the container's entrypoint.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

exec docker compose run --rm --service-ports dev "$@"
