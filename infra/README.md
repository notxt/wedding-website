# infra

CloudFormation templates for the wedding website. Deploy scripts live in `bin/`.

## Stacks

| Stack         | Template               | Purpose                                              |
|---------------|------------------------|------------------------------------------------------|
| `wedding-bootstrap` | `infra/bootstrap.yaml` | S3 bucket used to stage CFN deployment artifacts for later stacks. |
| `wedding-network`   | `infra/network.yaml`   | VPC + 4 subnets + IGW + route tables. Imported by `wedding-data` and `wedding-app`. |

## Prerequisites

- AWS credentials available to the shell. Run `aws login` on the host once; the dev container mounts `~/.aws` read-only and resolves the profile from the cached login session (see `CLAUDE.md` and `compose.yml`).
- Region defaults to `us-west-2`. Override with `AWS_REGION=...` if you ever need a different region.
- `cfn-lint` is installed in the dev container (`dev/Dockerfile`). Lint any template with `cfn-lint infra/<template>.yaml`.

## Deploy

Each stack has a deploy script in `bin/` that does the same thing: cfn-lint → `aws cloudformation validate-template` → `aws cloudformation deploy` → print outputs. Idempotent — re-running with no changes exits 0 via `--no-fail-on-empty-changeset`.

```bash
./bin/deploy-bootstrap.sh    # one-time, account-level bucket for CFN artifacts
./bin/deploy-network.sh      # VPC + subnets; foundation for the rest
```

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
