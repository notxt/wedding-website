#!/usr/bin/env bash
set -euo pipefail

: "${AWS_REGION:=us-west-2}"
: "${PROJECT_NAME:=wedding}"
: "${STACK_NAME:=${PROJECT_NAME}-bootstrap}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMPLATE="$REPO_ROOT/infra/bootstrap.yaml"

echo "Linting $TEMPLATE with cfn-lint..."
cfn-lint "$TEMPLATE"

echo "Validating with aws cloudformation validate-template..."
aws cloudformation validate-template \
  --region "$AWS_REGION" \
  --template-body "file://$TEMPLATE" > /dev/null

echo "Deploying $STACK_NAME to $AWS_REGION..."
aws cloudformation deploy \
  --region "$AWS_REGION" \
  --stack-name "$STACK_NAME" \
  --template-file "$TEMPLATE" \
  --no-fail-on-empty-changeset \
  --parameter-overrides "ProjectName=$PROJECT_NAME"

echo
echo "Stack outputs:"
aws cloudformation describe-stacks \
  --region "$AWS_REGION" \
  --stack-name "$STACK_NAME" \
  --query 'Stacks[0].Outputs' \
  --output table
