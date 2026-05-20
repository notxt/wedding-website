#!/usr/bin/env bash
set -euo pipefail

: "${AWS_REGION:=us-west-2}"
: "${PROJECT_NAME:=wedding}"
: "${APP_STACK_NAME:=${PROJECT_NAME}-app}"
: "${BOOTSTRAP_STACK_NAME:=${PROJECT_NAME}-bootstrap}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# --- look up dependencies from existing stacks ---
echo "Looking up bootstrap CFN bucket..."
BUCKET=$(aws cloudformation describe-stacks \
  --region "$AWS_REGION" \
  --stack-name "$BOOTSTRAP_STACK_NAME" \
  --query "Stacks[0].Outputs[?OutputKey=='CfnBucketName'].OutputValue" \
  --output text)
[ -n "$BUCKET" ] && [ "$BUCKET" != "None" ] || { echo "ERROR: CfnBucketName output not found on ${BOOTSTRAP_STACK_NAME}." >&2; exit 1; }
echo "  bucket: s3://${BUCKET}"

echo "Looking up CodeDeploy app + deployment group..."
APP_NAME=$(aws cloudformation describe-stacks \
  --region "$AWS_REGION" \
  --stack-name "$APP_STACK_NAME" \
  --query "Stacks[0].Outputs[?OutputKey=='CodeDeployApplicationName'].OutputValue" \
  --output text)
DG_NAME=$(aws cloudformation describe-stacks \
  --region "$AWS_REGION" \
  --stack-name "$APP_STACK_NAME" \
  --query "Stacks[0].Outputs[?OutputKey=='DeploymentGroupName'].OutputValue" \
  --output text)
[ -n "$APP_NAME" ] && [ "$APP_NAME" != "None" ] || { echo "ERROR: app stack not deployed yet — run bin/deploy-app-stack.sh first." >&2; exit 1; }
echo "  CodeDeploy app: $APP_NAME / deployment group: $DG_NAME"

# --- build the binary (linux arm64, static, stripped) ---
BUILD_DIR="$REPO_ROOT/build"
BUNDLE_DIR="$BUILD_DIR/bundle"
rm -rf "$BUILD_DIR"
mkdir -p "$BUNDLE_DIR"

echo "Building Go binary (GOOS=linux GOARCH=arm64)..."
( cd "$REPO_ROOT" && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
    go build -trimpath -ldflags='-s -w' -o "$BUNDLE_DIR/server" ./cmd/server )

# --- bundle ---
echo "Assembling deploy bundle..."
cp -r "$REPO_ROOT/deploy/." "$BUNDLE_DIR/"
chmod +x "$BUNDLE_DIR"/scripts/*.sh

SHA=$(cd "$REPO_ROOT" && git rev-parse --short HEAD)
ZIP="$BUILD_DIR/wedding-website-${SHA}.zip"

# Use python3 zipfile (always present in the dev container) and preserve Unix exec bits.
python3 - "$BUNDLE_DIR" "$ZIP" <<'PY'
import os, sys, zipfile
src, out = sys.argv[1], sys.argv[2]
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as zf:
    for root, _, files in os.walk(src):
        for f in files:
            full = os.path.join(root, f)
            rel = os.path.relpath(full, src)
            st = os.stat(full)
            info = zipfile.ZipInfo(rel)
            info.external_attr = (st.st_mode & 0xFFFF) << 16
            info.compress_type = zipfile.ZIP_DEFLATED
            with open(full, "rb") as fh:
                zf.writestr(info, fh.read())
PY
echo "  bundle: $ZIP ($(du -h "$ZIP" | cut -f1))"

# --- upload to s3 ---
S3_KEY="codedeploy-revisions/wedding-website-${SHA}.zip"
echo "Uploading to s3://${BUCKET}/${S3_KEY}..."
aws s3 cp --quiet --region "$AWS_REGION" "$ZIP" "s3://${BUCKET}/${S3_KEY}"

# --- trigger CodeDeploy ---
echo "Creating CodeDeploy deployment..."
DEPLOYMENT_ID=$(aws deploy create-deployment \
  --region "$AWS_REGION" \
  --application-name "$APP_NAME" \
  --deployment-group-name "$DG_NAME" \
  --description "git-sha=${SHA}" \
  --revision "revisionType=S3,s3Location={bucket=${BUCKET},key=${S3_KEY},bundleType=zip}" \
  --query 'deploymentId' --output text)
echo "  deployment: $DEPLOYMENT_ID"

echo "Waiting for deployment to complete..."
aws deploy wait deployment-successful \
  --region "$AWS_REGION" \
  --deployment-id "$DEPLOYMENT_ID"

echo "Deployment successful."
