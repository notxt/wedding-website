# infra

CloudFormation templates for the wedding website. Deploy scripts live in `bin/`.

## Stacks

| Stack         | Template               | Purpose                                              |
|---------------|------------------------|------------------------------------------------------|
| `wedding-bootstrap` | `infra/bootstrap.yaml` | S3 bucket used to stage CFN deployment artifacts for later stacks. |
| `wedding-network`   | `infra/network.yaml`   | VPC + 4 subnets + IGW + route tables. Imported by `wedding-data` and `wedding-app`. |
| `wedding-data`      | `infra/data.yaml`      | RDS Postgres in the private subnets, plus the `ACCESS_CODE` + `SESSION_SECRET` secrets the app reads at boot. |
| `wedding-app`       | `infra/app.yaml`       | ACM + ALB + Route53 + EC2 ASG (1/1/1) + CodeDeploy. Brings the site live at `https://michaelandadrien4ever.com`. |

## Prerequisites

- AWS credentials available to the shell. Run `aws login` on the host once; the dev container mounts `~/.aws` read-only and resolves the profile from the cached login session (see `CLAUDE.md` and `compose.yml`).
- Region defaults to `us-west-2`. Override with `AWS_REGION=...` if you ever need a different region.
- `cfn-lint` is installed in the dev container (`dev/Dockerfile`). Lint any template with `cfn-lint infra/<template>.yaml`.

## Deploy

Each stack has a deploy script in `bin/` that does the same thing: cfn-lint → `aws cloudformation validate-template` → `aws cloudformation deploy` → print outputs. Idempotent — re-running with no changes exits 0 via `--no-fail-on-empty-changeset`.

```bash
./bin/deploy-bootstrap.sh                                       # one-time, account-level bucket for CFN artifacts
./bin/deploy-network.sh                                         # VPC + subnets; foundation for the rest
ACCESS_CODE='<your shared phrase>' ./bin/deploy-data.sh         # RDS + secrets (first run only; re-runs no-op the secrets)
./bin/deploy-app-stack.sh                                       # ALB + ACM + Route53 + ASG + CodeDeploy
./bin/deploy-app.sh                                             # build + bundle + upload + trigger CodeDeploy (run any time to ship app changes)
```

`deploy-data.sh` provisions two Secrets Manager secrets (`wedding/access-code`, `wedding/session-secret`) **outside** CFN so their values never flow through stack parameters, then passes their ARNs into the stack. On first run it requires the `ACCESS_CODE` env var; subsequent runs reuse the existing secrets. The session secret is auto-generated (`openssl rand -base64 32`) on first run.

`deploy-app-stack.sh` looks up the Route53 hosted zone for `michaelandadrien4ever.com` and passes its ID into CFN. The stack creates an ACM cert DNS-validated against that zone, an ALB, an ASG of 1, and a CodeDeploy app. The instance comes up with no application binary — the first `deploy-app.sh` run installs it.

`deploy-app.sh` cross-compiles a static linux/arm64 binary, zips it with the contents of `deploy/`, uploads to `s3://${bootstrap.CfnBucket}/codedeploy-revisions/wedding-website-<git-sha>.zip`, then triggers and waits on a CodeDeploy deployment.

Stack outputs print at the end. You can also fetch them later:

```bash
aws cloudformation describe-stacks \
  --region us-west-2 \
  --stack-name wedding-bootstrap \
  --query 'Stacks[0].Outputs' --output table
```

## Tearing it down

The CFN bucket has `DeletionPolicy: Retain` and `UpdateReplacePolicy: Retain` — deleting the stack will **not** delete the bucket or its contents. To fully remove everything:

```bash
BUCKET=$(aws cloudformation describe-stacks \
  --region us-west-2 \
  --stack-name wedding-bootstrap \
  --query "Stacks[0].Outputs[?OutputKey=='CfnBucketName'].OutputValue" \
  --output text)

aws s3 rb "s3://$BUCKET" --force
aws cloudformation delete-stack --region us-west-2 --stack-name wedding-bootstrap
```
