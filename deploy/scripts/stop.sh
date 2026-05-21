#!/usr/bin/env bash
set -euo pipefail
# best-effort: succeed even if the unit isn't installed yet (first deploy)
systemctl stop wedding-website.service || true
