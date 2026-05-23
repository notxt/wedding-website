# 016 — Dev container SSM access + RDS tunnel helper

## Why
The two admins read RSVP/guest data by asking an agent (Claude Code) questions like "what's the
guest count?". The prod RDS is private (no public access, no Data API), so the only path in is an
SSM port-forward through the app EC2 instance. The dev container lacks `session-manager-plugin`, and
there's no convenient agent-friendly way to run a one-off query. No CFN change is needed — the
instance role already has `AmazonSSMManagedInstanceCore`.

## Acceptance criteria
- [ ] `dev/Dockerfile` installs `session-manager-plugin` (per-arch RPM, mirroring the arch-`case` style)
- [ ] `bin/db-tunnel.sh` opens an SSM port-forward to RDS and runs `psql`, with:
  - [ ] `-c "SQL"` / `-f file.sql` / stdin non-interactive modes that print to stdout and exit
  - [ ] interactive `psql` when run on a TTY with no SQL
  - [ ] read-only by default (`default_transaction_read_only=on`); `--write` lifts it
  - [ ] background tunnel with guaranteed teardown on exit; never prints the DB password
- [ ] `CLAUDE.md` documents the helper, the required host image rebuild, and caller IAM prerequisites
- [ ] `bash -n` (and `shellcheck` if available) pass; read-only AWS resolution (instance id + exports) confirmed

## Notes
- DB creds come from the RDS master secret (`wedding-data-DbMasterSecretArn`) for now; structure the
  credential fetch so a later swap to RDS IAM auth is easy. IAM auth is deferred pending a data-loss review.
- Caller prerequisites (not in this PR): `ssm:StartSession` on the instance + the
  `AWS-StartPortForwardingSessionToRemoteHost` document, and `secretsmanager:GetSecretValue` on the master secret.
- Defaults mirror `bin/deploy-*.sh`: `AWS_REGION=us-west-2`, `PROJECT_NAME=wedding`; `LOCAL_PORT=15432`.
