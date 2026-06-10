#!/usr/bin/env bash
set -euo pipefail

: "${AWS_REGION:=us-west-2}"
: "${PROJECT_NAME:=wedding}"
: "${STACK_NAME:=${PROJECT_NAME}-app}"
: "${NETWORK_STACK_NAME:=${PROJECT_NAME}-network}"
: "${DATA_STACK_NAME:=${PROJECT_NAME}-data}"
: "${BOOTSTRAP_STACK_NAME:=${PROJECT_NAME}-bootstrap}"
: "${DOMAIN_NAME:=michaelandadrien4ever.com}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMPLATE="$REPO_ROOT/infra/app.yaml"

echo "Looking up Route53 hosted zone for ${DOMAIN_NAME}..."
HOSTED_ZONE_FULL_ID=$(aws route53 list-hosted-zones-by-name \
  --dns-name "${DOMAIN_NAME}." \
  --query "HostedZones[?Name==\`${DOMAIN_NAME}.\`].Id" \
  --output text | head -1)
if [ -z "$HOSTED_ZONE_FULL_ID" ] || [ "$HOSTED_ZONE_FULL_ID" = "None" ]; then
  echo "ERROR: no Route53 hosted zone found for ${DOMAIN_NAME}." >&2
  exit 1
fi
HOSTED_ZONE_ID="${HOSTED_ZONE_FULL_ID##*/}"
echo "  HostedZoneId: ${HOSTED_ZONE_ID}"

echo "Looking up DbMasterSecretArn from the data-stack export (self-managed, stable ARN)..."
DB_MASTER_SECRET_ARN=$(aws cloudformation list-exports \
  --region "$AWS_REGION" \
  --query "Exports[?Name=='${DATA_STACK_NAME}-DbMasterSecretArn'].Value" \
  --output text)
[ -n "$DB_MASTER_SECRET_ARN" ] && [ "$DB_MASTER_SECRET_ARN" != "None" ] || { echo "ERROR: export ${DATA_STACK_NAME}-DbMasterSecretArn not found. Deploy the data stack first." >&2; exit 1; }
echo "  DbMasterSecretArn: $DB_MASTER_SECRET_ARN"

echo "Linting $TEMPLATE with cfn-lint..."
cfn-lint "$TEMPLATE"

echo "Validating with aws cloudformation validate-template..."
aws cloudformation validate-template \
  --region "$AWS_REGION" \
  --template-body "file://$TEMPLATE" > /dev/null

echo "Deploying $STACK_NAME to $AWS_REGION (cert DNS validation + ALB + ASG; typically 5-10 minutes)..."
aws cloudformation deploy \
  --region "$AWS_REGION" \
  --stack-name "$STACK_NAME" \
  --template-file "$TEMPLATE" \
  --no-fail-on-empty-changeset \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides \
    "ProjectName=$PROJECT_NAME" \
    "NetworkStackName=$NETWORK_STACK_NAME" \
    "DataStackName=$DATA_STACK_NAME" \
    "BootstrapStackName=$BOOTSTRAP_STACK_NAME" \
    "DomainName=$DOMAIN_NAME" \
    "HostedZoneId=$HOSTED_ZONE_ID" \
    "DbMasterSecretArn=$DB_MASTER_SECRET_ARN"

echo
echo "Stack outputs:"
aws cloudformation describe-stacks \
  --region "$AWS_REGION" \
  --stack-name "$STACK_NAME" \
  --query 'Stacks[0].Outputs' \
  --output table
