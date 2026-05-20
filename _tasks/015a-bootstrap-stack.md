# 015a — Bootstrap CFN stack + CFN dev tooling

## Why
First infra PR. Carved out of `015-aws-cfn-infra.md`. Two things need to land before anything else: (1) a way to lint/validate templates locally so iteration is fast, and (2) an S3 bucket future stacks can stage CFN deployment artifacts to (`aws cloudformation package` won't run without one).

## Acceptance criteria
- [ ] `dev/Dockerfile` installs `cfn-lint` (pinned).
- [ ] `docker compose build dev` succeeds; `cfn-lint --version` works inside the rebuilt container.
- [ ] `infra/bootstrap.yaml` defines a single retained S3 bucket for CFN deployment artifacts (versioning, encryption, public access blocked, lifecycle rules).
- [ ] `cfn-lint infra/bootstrap.yaml` exits 0 with no findings.
- [ ] `aws cloudformation validate-template --template-body file://infra/bootstrap.yaml --region us-west-2` succeeds.
- [ ] `bin/deploy-bootstrap.sh` deploys the stack (defaults region to `us-west-2`, runs cfn-lint + validate-template pre-flight, prints outputs).
- [ ] `infra/README.md` documents prerequisites, deploy command, and teardown.
- [ ] `.gitignore` no longer ignores `bin/`.
- [ ] Stack deployed successfully end-to-end against the real AWS account; bucket appears in `s3 ls`.
- [ ] `_tasks/015a-bootstrap-stack.md` deleted in the merging PR.

## Notes
- Bucket name pattern: `${AWS::AccountId}-${AWS::Region}-${ProjectName}-cfn`.
- Outputs exported as `${ProjectName}-bootstrap-CfnBucketName` / `${ProjectName}-bootstrap-CfnBucketArn` so later stacks can `Fn::ImportValue`.
- `DeletionPolicy: Retain` is intentional — accidental stack delete must not blow away staged artifacts.
- After this PR merges, rebuild the dev container (`docker compose build dev`) to pick up `cfn-lint`.
