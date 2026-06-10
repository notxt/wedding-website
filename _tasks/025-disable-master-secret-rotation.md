# 025 — Self-managed master DB secret, no rotation

## Why

The RDS-managed master password rotates every 7 days. The instance bakes the password into `/etc/wedding-website/env` at boot via userdata, so when rotation fires the env file goes stale and the app can no longer authenticate. This caused a real outage on 2026-06-09 (RSVP page returned "Something went wrong" until the next instance launch).

Pragmatic fix for a wedding website (private-subnet DB, IAM-gated SSM tunnel only path in): self-manage the master secret in CFN with no rotation. Threat model favors operational stability; rotation buys ~nothing given the access path.

## Acceptance criteria

- [ ] Pre-deploy snapshot of `wedding` RDS instance exists in `available` state
- [ ] `infra/data.yaml` defines a self-managed `AWS::SecretsManager::Secret` (`DbMasterSecret`) with `GenerateSecretString`, `DeletionPolicy: Retain`, no `RotationSchedule`
- [ ] `infra/data.yaml`'s `DbInstance` no longer sets `ManageMasterUserPassword`; uses `MasterUserPassword` dynamic reference to the self-managed secret
- [ ] `AWS::SecretsManager::SecretTargetAttachment` (`DbMasterSecretAttachment`) links the secret to the DB instance (canonical AWS pattern; augments the secret with host/port/dbname/engine)
- [ ] `DbMasterSecretArn` output points at the self-managed secret (ARN is stable forever)
- [ ] `bin/deploy-app-stack.sh` reads `DbMasterSecretArn` from the CFN export, not `rds describe-db-instances`
- [ ] `bin/db-tunnel.sh` reads `DbMasterSecretArn` from the CFN export, not `rds describe-db-instances`
- [ ] SPEC.md documents the no-rotation decision and threat model
- [ ] After deploy: `aws rds describe-db-instances` shows `MasterUserSecret: null`, the new secret value matches the DB's current password, site loads and accepts RSVPs end-to-end

## Notes

- Pre-deploy snapshot ID (recorded at start): `wedding-pre-self-managed-secret-20260610-040636`
- Old RDS-managed secret gets scheduled for deletion (7-day grace) when we opt out — recoverable during the window if anything weird happens.
- `DeletionPolicy: Retain` and `UpdateReplacePolicy: Retain` on `DbInstance` remain in place; password changes are in-place modifications and never trigger replacement.
- Phase 0 (snapshot) → Phase 2 (deploy data, then app) is documented in the PR description.
