#!/usr/bin/env bash
set -euo pipefail

: "${AWS_REGION:=us-west-2}"
: "${PROJECT_NAME:=wedding}"
: "${STACK_NAME:=${PROJECT_NAME}-data}"
: "${NETWORK_STACK_NAME:=${PROJECT_NAME}-network}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMPLATE="$REPO_ROOT/infra/data.yaml"

ACCESS_CODE_SECRET_NAME="${PROJECT_NAME}/access-code"
SESSION_SECRET_NAME="${PROJECT_NAME}/session-secret"

get_secret_arn() {
  aws secretsmanager describe-secret \
    --region "$AWS_REGION" \
    --secret-id "$1" \
    --query 'ARN' \
    --output text 2>/dev/null || true
}

# --- ACCESS_CODE secret ---
echo "Checking secret ${ACCESS_CODE_SECRET_NAME}..."
ACCESS_CODE_ARN="$(get_secret_arn "$ACCESS_CODE_SECRET_NAME")"
if [ -z "$ACCESS_CODE_ARN" ] || [ "$ACCESS_CODE_ARN" = "None" ]; then
  if [ -z "${ACCESS_CODE:-}" ]; then
    echo "ERROR: secret '${ACCESS_CODE_SECRET_NAME}' does not exist yet, and ACCESS_CODE env var is not set." >&2
    echo "First-time deploy requires:  ACCESS_CODE='<your shared phrase>' $0" >&2
    exit 1
  fi
  echo "Creating ${ACCESS_CODE_SECRET_NAME} from ACCESS_CODE env var..."
  ACCESS_CODE_ARN="$(aws secretsmanager create-secret \
    --region "$AWS_REGION" \
    --name "$ACCESS_CODE_SECRET_NAME" \
    --description "Shared RSVP access code for the wedding site (read by EC2 instance at boot)." \
    --secret-string "$ACCESS_CODE" \
    --query 'ARN' --output text)"
fi
echo "  ACCESS_CODE secret: $ACCESS_CODE_ARN"

# --- SESSION_SECRET secret ---
echo "Checking secret ${SESSION_SECRET_NAME}..."
SESSION_SECRET_ARN="$(get_secret_arn "$SESSION_SECRET_NAME")"
if [ -z "$SESSION_SECRET_ARN" ] || [ "$SESSION_SECRET_ARN" = "None" ]; then
  echo "Generating ${SESSION_SECRET_NAME} (32 random bytes, base64)..."
  SESSION_VALUE="$(openssl rand -base64 32)"
  SESSION_SECRET_ARN="$(aws secretsmanager create-secret \
    --region "$AWS_REGION" \
    --name "$SESSION_SECRET_NAME" \
    --description "HMAC key for the RSVP auth cookie." \
    --secret-string "$SESSION_VALUE" \
    --query 'ARN' --output text)"
fi
echo "  SESSION_SECRET secret: $SESSION_SECRET_ARN"

# --- Stack ---
echo "Linting $TEMPLATE with cfn-lint..."
cfn-lint "$TEMPLATE"

echo "Validating with aws cloudformation validate-template..."
aws cloudformation validate-template \
  --region "$AWS_REGION" \
  --template-body "file://$TEMPLATE" > /dev/null

echo "Deploying $STACK_NAME to $AWS_REGION (RDS creation typically takes 5-10 minutes)..."
aws cloudformation deploy \
  --region "$AWS_REGION" \
  --stack-name "$STACK_NAME" \
  --template-file "$TEMPLATE" \
  --no-fail-on-empty-changeset \
  --parameter-overrides \
    "ProjectName=$PROJECT_NAME" \
    "NetworkStackName=$NETWORK_STACK_NAME" \
    "AccessCodeSecretArn=$ACCESS_CODE_ARN" \
    "SessionSecretArn=$SESSION_SECRET_ARN"

echo
echo "Stack outputs:"
aws cloudformation describe-stacks \
  --region "$AWS_REGION" \
  --stack-name "$STACK_NAME" \
  --query 'Stacks[0].Outputs' \
  --output table
