#!/usr/bin/env bash
set -euo pipefail

# Configure git to fetch a fresh GitHub App installation token on every push.
# This avoids embedding tokens in the remote URL or worrying about hourly expiry.
if [ -r "${GH_APP_KEY_FILE:-/dev/null}" ]; then
  git config --global \
    credential."https://github.com".helper \
    '!f() { echo "username=x-access-token"; echo "password=$(/usr/local/bin/gh-app-token)"; }; f'
  git config --global url."https://github.com/".insteadOf "git@github.com:"
  git config --global user.name "claude-notxt[bot]"
  git config --global user.email "281474455+claude-notxt[bot]@users.noreply.github.com"
fi

# Convenience for `gh` invocations during this session. The token expires in 1h;
# long-running shells can refresh with: export GH_TOKEN=$(gh-app-token)
if command -v /usr/local/bin/gh-app-token >/dev/null 2>&1 \
   && [ -r "${GH_APP_KEY_FILE:-/dev/null}" ]; then
  export GH_TOKEN="$(/usr/local/bin/gh-app-token 2>/dev/null || true)"
fi

exec "$@"
