#!/usr/bin/env bash
set -euo pipefail
# Poll /healthz on the local instance until it returns 200, up to ~45 seconds.
for i in $(seq 1 45); do
  if curl -sf -o /dev/null --max-time 2 http://127.0.0.1:8080/healthz; then
    echo "healthz ok after ${i}s"
    exit 0
  fi
  sleep 1
done
echo "validate.sh: /healthz did not return 200 within 45s" >&2
journalctl -u wedding-website --no-pager --lines 50 >&2 || true
exit 1
