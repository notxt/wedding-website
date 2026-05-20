# 015c — Data stack: RDS + secrets

## Why
Durable layer for the wedding site: RDS Postgres + Secrets Manager. Split into its own stack so guest data survives any teardown or rebuild of the app side. Imports the VPC from `wedding-network` (015b).

## Acceptance criteria
- [ ] `infra/data.yaml` defines stack `wedding-data` with:
  - DB subnet group across the imported private subnets.
  - DB security group in the imported VPC (no ingress yet — 015d adds the rule from the app SG cross-stack).
  - RDS Postgres: `db.t4g.micro`, single-AZ, 20 GB GP3, backup retention 7 days, `StorageEncrypted: true`, `DeletionProtection: true`, `DeletionPolicy: Retain`, `UpdateReplacePolicy: Retain`. Master password via `ManageMasterUserPassword: true`.
  - Secrets Manager entries (both `DeletionPolicy: Retain`):
    - `wedding/access-code` — operator-supplied; the access code shared with guests.
    - `wedding/session-secret` — 32 random bytes, base64-encoded; deploy script generates on first run.
- [ ] `bin/deploy-data.sh` creates/checks the two app secrets via `aws secretsmanager` **before** running `aws cloudformation deploy` — passes their ARNs as parameters so no secret values flow through CFN. Refuses to deploy if `wedding/access-code` doesn't yet exist and no `ACCESS_CODE` env var is supplied.
- [ ] `cfn-lint infra/data.yaml` exit 0; `./bin/deploy-data.sh` succeeds end-to-end against AWS.
- [ ] RDS shows `available`, deletion protection ON, encrypted, single-AZ, 7-day backups.
- [ ] Stack exports available for 015d to import: `wedding-data-DbEndpoint`, `-DbPort`, `-DbName`, `-DbMasterSecretArn`, `-DbSecurityGroupId`, `-AccessCodeSecretArn`, `-SessionSecretArn`.
- [ ] `infra/README.md` updated with the new stack + deploy command.
- [ ] `_tasks/015c-data-stack.md` deleted in the merging PR.

## Notes
- Region `us-west-2`, account `755621201894`.
- Imports from `wedding-network`: `VpcId`, `PrivateSubnetIds`.
- 015d adds DB SG ingress 5432 from the app SG via `AWS::EC2::SecurityGroupIngress` (cross-stack ref to the export below).
- `ManageMasterUserPassword: true` makes RDS own and rotate the master credential automatically; the app reads it from Secrets Manager at instance boot.
- Data-loss protections layered: stack separation + `DeletionPolicy: Retain` + `UpdateReplacePolicy: Retain` + `DeletionProtection: true` + encryption + 7-day backups + final snapshot semantics (RDS refuses to delete without `--skip-final-snapshot`).
