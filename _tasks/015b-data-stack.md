# 015b — Data stack: VPC + RDS + secrets

## Why
Durable layer for the wedding site: VPC, RDS Postgres, and Secrets Manager entries — split out from the app stack so guest data survives any teardown or rebuild of the app side.

## Acceptance criteria
- [ ] `infra/data.yaml` defines stack `wedding-data` with:
  - VPC `10.0.0.0/16`, 2 AZs (us-west-2a / us-west-2b), 2 public + 2 private subnets, IGW + route tables.
  - DB subnet group across the private subnets.
  - DB security group (no ingress yet — 015c adds the rule from the app SG cross-stack).
  - RDS Postgres: `db.t4g.micro`, single-AZ, 20 GB GP3, backup retention 7 days, `StorageEncrypted: true`, `DeletionProtection: true`, `DeletionPolicy: Retain`, `UpdateReplacePolicy: Retain`. Master password via `ManageMasterUserPassword: true`.
  - Secrets Manager entries (both `DeletionPolicy: Retain`):
    - `wedding/access-code` — operator-supplied; the access code shared with guests.
    - `wedding/session-secret` — 32 random bytes, base64-encoded; deploy script generates on first run.
- [ ] `bin/deploy-data.sh` creates/checks the two app secrets via `aws secretsmanager` **before** running `aws cloudformation deploy` — passes their ARNs as parameters so no secret values flow through CFN. Refuses to deploy if `wedding/access-code` doesn't yet exist and no `ACCESS_CODE` env var is supplied.
- [ ] `cfn-lint infra/data.yaml` exit 0; `./bin/deploy-data.sh` succeeds end-to-end against AWS.
- [ ] RDS shows `available`, deletion protection ON, encrypted, single-AZ, 7-day backups.
- [ ] Stack exports available for 015c to import: `wedding-data-VpcId`, `-PublicSubnetIds`, `-PrivateSubnetIds`, `-DbEndpoint`, `-DbPort`, `-DbName`, `-DbMasterSecretArn`, `-DbSecurityGroupId`, `-AccessCodeSecretArn`, `-SessionSecretArn`.
- [ ] `infra/README.md` updated with the new stack + deploy command.
- [ ] `_tasks/015b-data-stack.md` deleted in the merging PR.

## Notes
- Region `us-west-2`, account `755621201894`.
- The VPC lives in this stack (not a separate network stack) because the DB depends on it — if the data side survives a teardown, its network should too. Keeps the blast-radius boundary clean.
- Public subnets will host ALB + EC2 (015c); private subnets host RDS only. No NAT gateway — the EC2 instance gets AWS-API egress through the IGW from a public subnet.
- 015c adds DB SG ingress 5432 from the app SG via `AWS::EC2::SecurityGroupIngress` (cross-stack ref).
- `ManageMasterUserPassword: true` makes RDS own and rotate the master credential automatically; the app reads it from Secrets Manager at instance boot.
