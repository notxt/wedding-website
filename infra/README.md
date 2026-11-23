# infra

> **Torn down 2026-11-23.** None of these stacks currently exist. See "Tearing it down" for what was done and "Bringing it back" for how to redeploy.

CloudFormation templates for the wedding website. Deploy scripts live in `bin/`.

## Stacks

| Stack         | Template               | Purpose                                              |
|---------------|------------------------|------------------------------------------------------|
| `wedding-bootstrap` | `infra/bootstrap.yaml` | S3 bucket used to stage CFN deployment artifacts for later stacks. |
| `wedding-network`   | `infra/network.yaml`   | VPC + 4 subnets + IGW + route tables. Imported by `wedding-data` and `wedding-app`. |
| `wedding-data`      | `infra/data.yaml`      | RDS Postgres in the private subnets, plus the `ACCESS_CODE` + `SESSION_SECRET` secrets the app reads at boot. |
| `wedding-app`       | `infra/app.yaml`       | ACM + ALB + Route53 + EC2 ASG (1/1/1) + CodeDeploy. Brings the site live at `https://michaelandadrien4ever.com`. |

## Prerequisites

- AWS credentials available to the shell. Run `aws login` on the host once; the dev container mounts the host `~/.aws/config` read-only and `~/.aws/login` read-write (sharing the live credential cache) and resolves the profile from the cached login session (see `CLAUDE.md` and `compose.yml`).
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

This was done on 2026-11-23, after the wedding. Deleting the stacks alone is **not** enough: several resources are retained on purpose (`DeletionPolicy: Retain`) or live outside CloudFormation, and the order matters because `wedding-app` imports from the other three stacks and `wedding-data` imports from `wedding-network`.

Export the data first (`bin/db-tunnel.sh` stops working once `wedding-app` is gone, because it tunnels through the app instance).

```bash
export AWS_DEFAULT_REGION=us-west-2
ACCOUNT=$(aws sts get-caller-identity --query Account --output text)

# 1. App stack: ALB, ASG/EC2, ACM cert, A records, CodeDeploy, alarms. Site goes offline.
aws cloudformation delete-stack --stack-name wedding-app
aws cloudformation wait stack-delete-complete --stack-name wedding-app

# 2. Database. It is Retain + deletion-protected, and wedding-data cannot delete its
#    subnet group / security group while the instance exists, so remove it by hand first.
aws rds modify-db-instance --db-instance-identifier wedding --no-deletion-protection --apply-immediately
aws rds delete-db-instance --db-instance-identifier wedding \
  --final-db-snapshot-identifier "wedding-final-$(date +%F)" --delete-automated-backups
aws rds wait db-instance-deleted --db-instance-identifier wedding

# 3. Data, then network.
aws cloudformation delete-stack --stack-name wedding-data
aws cloudformation wait stack-delete-complete --stack-name wedding-data
aws cloudformation delete-stack --stack-name wedding-network
aws cloudformation wait stack-delete-complete --stack-name wedding-network

# 4. Retained buckets. The ALB log bucket is unversioned; the CFN bucket is VERSIONED,
#    so `s3 rb --force` is not enough for it: every object version must be deleted.
aws s3 rb "s3://${ACCOUNT}-us-west-2-wedding-alb-logs" --force
BUCKET="${ACCOUNT}-us-west-2-wedding-cfn"
aws s3api delete-objects --bucket "$BUCKET" --delete "$(aws s3api list-object-versions \
  --bucket "$BUCKET" --query '{Objects: [Versions, DeleteMarkers][][].{Key: Key, VersionId: VersionId}}' --output json)"
aws s3api delete-bucket --bucket "$BUCKET"
aws cloudformation delete-stack --stack-name wedding-bootstrap
aws cloudformation wait stack-delete-complete --stack-name wedding-bootstrap

# 5. Secrets: one retained by wedding-data, two created by deploy-data.sh outside CFN,
#    and (in 2026) an orphaned wedding/db-master from before the self-managed secret.
for s in wedding/db/master wedding/access-code wedding/session-secret wedding/db-master; do
  aws secretsmanager delete-secret --secret-id "$s" --force-delete-without-recovery
done

# 6. Old manual RDS snapshots: list them, delete all but the final one.
aws rds describe-db-snapshots --snapshot-type manual --query 'DBSnapshots[].DBSnapshotIdentifier'
aws rds delete-db-snapshot --db-snapshot-identifier <old-snapshot-id>

# 7. Domain and DNS. The hosted zone lives outside CFN; after the app stack is gone it
#    still holds the two ACM validation CNAMEs, which must be deleted before the zone
#    (Route53 console, or change-resource-record-sets with a DELETE batch).
aws route53domains disable-domain-auto-renew --region us-east-1 --domain-name michaelandadrien4ever.com
aws route53 delete-hosted-zone --id <zone-id>
```

### What was left behind in 2026

- RDS snapshot **`wedding-final-2026-11-23`** in us-west-2: the complete final `guests` and `rsvps` tables. Costs a few cents a month.
- The domain registration for `michaelandadrien4ever.com`, auto-renew off, expiring 2026-12-29. Its hosted zone was deleted.
- Nothing else. Before teardown the account ran at about $53/month (ALB ~$17, RDS ~$15, public IPv4 ~$11, EC2 ~$7, secrets and DNS ~$3).

## Bringing it back

The templates and scripts are unchanged, so the "Deploy" sequence above still works on an empty account. What it does not cover:

1. **Domain and DNS.** Register a domain (the original may have been taken after it lapsed), create a public Route53 hosted zone for it, and point the registration's name servers at that zone. `deploy-app-stack.sh` looks the zone up by name; the domain is the `DomainName` parameter in `infra/app.yaml`.
2. **Secrets.** `deploy-data.sh` recreates `wedding/access-code` (needs `ACCESS_CODE` in the environment) and `wedding/session-secret`.
3. **Data.** `infra/data.yaml` always creates an **empty** database; it has no snapshot parameter. To get the 2026 data back, either add `DBSnapshotIdentifier: wedding-final-2026-11-23` to `DbInstance` for the first deploy, or deploy empty and reload the guest list from `untracked/seed_guests.sql`. Migrations run on server start either way.
4. **Code.** `main` includes task 030 (dress-code page, copy refresh), which was never deployed. To reproduce exactly what guests saw, deploy the `last-deployed` tag instead.
