#!/usr/bin/env bash
# Query the private prod RDS by port-forwarding through the app EC2 instance
# over SSM, then running psql locally. Agent-friendly: pass SQL with -c/-f (or
# on stdin) for a one-shot query that prints to stdout and exits; run with no
# args on a terminal for an interactive psql shell.
#
# Read-only by default (default_transaction_read_only=on) so an accidental
# write can't touch the guest list; pass --write to allow writes.
#
# Requires AWS creds with: ssm:StartSession on the instance + the
# AWS-StartPortForwardingSessionToRemoteHost document, and
# secretsmanager:GetSecretValue on the RDS master secret. The dev container
# ships session-manager-plugin and psql (rebuild the image to get the plugin).
set -euo pipefail

: "${AWS_REGION:=us-west-2}"
: "${PROJECT_NAME:=wedding}"
: "${LOCAL_PORT:=15432}"

usage() {
  cat >&2 <<EOF
Usage: bin/db-tunnel.sh [-c "SQL" | -f FILE.sql] [--write] [--port N]
       bin/db-tunnel.sh                # interactive psql (on a TTY)
       echo "SQL" | bin/db-tunnel.sh   # read SQL from stdin

Env: AWS_REGION (default us-west-2), PROJECT_NAME (wedding), LOCAL_PORT (15432)
EOF
}

SQL=""
SQL_FILE=""
WRITE=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    -c|--command) SQL="${2:-}"; shift 2 ;;
    -f|--file)    SQL_FILE="${2:-}"; shift 2 ;;
    --write)      WRITE=1; shift ;;
    --port)       LOCAL_PORT="${2:-}"; shift 2 ;;
    -h|--help)    usage; exit 0 ;;
    *) echo "ERROR: unknown argument: $1" >&2; usage; exit 2 ;;
  esac
done
if [[ -n "$SQL" && -n "$SQL_FILE" ]]; then
  echo "ERROR: pass only one of -c / -f" >&2
  exit 2
fi

for b in aws jq psql; do
  command -v "$b" >/dev/null 2>&1 || { echo "ERROR: '$b' not found on PATH" >&2; exit 1; }
done
command -v session-manager-plugin >/dev/null 2>&1 || {
  echo "ERROR: session-manager-plugin not found — rebuild the dev container (docker compose build dev) to pick it up." >&2
  exit 1
}

aws_q() { aws --region "$AWS_REGION" "$@"; }

# --- resolve the app instance and DB connection facts ------------------------
INSTANCE_ID="$(aws_q autoscaling describe-auto-scaling-groups \
  --auto-scaling-group-names "${PROJECT_NAME}-app" \
  --query "AutoScalingGroups[0].Instances[?LifecycleState=='InService']|[0].InstanceId" \
  --output text)"
[[ -n "$INSTANCE_ID" && "$INSTANCE_ID" != "None" ]] \
  || { echo "ERROR: no InService instance in ASG ${PROJECT_NAME}-app" >&2; exit 1; }

get_export() {
  aws_q cloudformation list-exports --query "Exports[?Name=='$1'].Value" --output text
}
DB_ENDPOINT="$(get_export "${PROJECT_NAME}-data-DbEndpoint")"
DB_PORT="$(get_export "${PROJECT_NAME}-data-DbPort")"
DB_NAME="$(get_export "${PROJECT_NAME}-data-DbName")"
# The secret ARN comes from RDS itself, not the stack export: RDS may recreate
# the managed secret on rotation events, in which case the export lags reality.
DB_SECRET_ARN="$(aws_q rds describe-db-instances \
  --db-instance-identifier "$PROJECT_NAME" \
  --query 'DBInstances[0].MasterUserSecret.SecretArn' --output text)"
[[ -n "$DB_ENDPOINT" && "$DB_ENDPOINT" != "None" ]] \
  || { echo "ERROR: could not resolve ${PROJECT_NAME}-data-DbEndpoint export" >&2; exit 1; }
[[ -n "$DB_SECRET_ARN" && "$DB_SECRET_ARN" != "None" ]] \
  || { echo "ERROR: RDS instance ${PROJECT_NAME} has no MasterUserSecret" >&2; exit 1; }
[[ -n "$DB_PORT" && "$DB_PORT" != "None" ]] || DB_PORT=5432
[[ -n "$DB_NAME" && "$DB_NAME" != "None" ]] || DB_NAME="$PROJECT_NAME"

# --- open the SSM port-forward (background) with guaranteed teardown ---------
SSM_LOG="$(mktemp)"
SSM_PID=""
cleanup() {
  if [[ -n "$SSM_PID" ]]; then
    kill "$SSM_PID" 2>/dev/null || true
    wait "$SSM_PID" 2>/dev/null || true
  fi
  rm -f "$SSM_LOG" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# stdin from /dev/null so the backgrounded session never eats our SQL on stdin.
aws_q ssm start-session \
  --target "$INSTANCE_ID" \
  --document-name AWS-StartPortForwardingSessionToRemoteHost \
  --parameters "host=$DB_ENDPOINT,portNumber=$DB_PORT,localPortNumber=$LOCAL_PORT" \
  >"$SSM_LOG" 2>&1 </dev/null &
SSM_PID=$!

# wait for the local end of the tunnel to accept connections (~20s budget)
tries=0
until (exec 3<>"/dev/tcp/127.0.0.1/$LOCAL_PORT") 2>/dev/null; do
  if ! kill -0 "$SSM_PID" 2>/dev/null; then
    echo "ERROR: SSM session exited before the tunnel opened:" >&2
    cat "$SSM_LOG" >&2 || true
    exit 1
  fi
  tries=$((tries + 1))
  if [[ "$tries" -ge 40 ]]; then
    echo "ERROR: tunnel did not open on 127.0.0.1:$LOCAL_PORT within timeout:" >&2
    cat "$SSM_LOG" >&2 || true
    exit 1
  fi
  sleep 0.5
done
echo "→ tunnel 127.0.0.1:$LOCAL_PORT → $DB_ENDPOINT:$DB_PORT (via $INSTANCE_ID)" >&2

# --- credentials -------------------------------------------------------------
# Sets DB_USER and exports PGPASSWORD from the RDS master secret. To move to
# RDS IAM auth later, swap this function body for `aws rds generate-db-auth-token`.
load_db_credentials() {
  local secret
  secret="$(aws_q secretsmanager get-secret-value --secret-id "$DB_SECRET_ARN" \
    --query SecretString --output text)"
  DB_USER="$(printf '%s' "$secret" | jq -r '.username')"
  PGPASSWORD="$(printf '%s' "$secret" | jq -r '.password')"
  export PGPASSWORD
}
load_db_credentials
[[ -n "$DB_USER" && "$DB_USER" != "null" ]] \
  || { echo "ERROR: could not read DB username from the master secret" >&2; exit 1; }

# --- run psql over the tunnel ------------------------------------------------
export PGSSLMODE=require
if [[ "$WRITE" -eq 0 ]]; then
  export PGOPTIONS="-c default_transaction_read_only=on"
fi

PSQL=(psql -h 127.0.0.1 -p "$LOCAL_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1)
if [[ -n "$SQL" ]]; then
  "${PSQL[@]}" -c "$SQL"
elif [[ -n "$SQL_FILE" ]]; then
  "${PSQL[@]}" -f "$SQL_FILE"
elif [[ ! -t 0 ]]; then
  "${PSQL[@]}" -f -
else
  "${PSQL[@]}"
fi
