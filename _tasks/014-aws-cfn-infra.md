# 014 — AWS infrastructure (CloudFormation)

## Why
Move from local-only to a real deployment. **Phase 2 — do not start until the user signs off on the local site.**

## Acceptance criteria
- [ ] Decide compute target: ECS Fargate vs App Runner (App Runner is simpler; Fargate has more knobs). Document the choice in `SPEC.md`.
- [ ] CFN template under `infra/`:
  - VPC + subnets (or use default VPC for App Runner)
  - RDS Postgres (db.t4g.micro is fine; private subnet; auto backup retention 7 days)
  - Compute service running the Go server, env vars wired from Secrets Manager
  - Secrets Manager entries for `ACCESS_CODE` and `SESSION_SECRET`
  - ALB + ACM cert + Route53 record (or App Runner custom domain)
  - CloudWatch log group with sane retention
- [ ] Dockerfile that builds the server (multi-stage; final image distroless or `gcr.io/distroless/static`)
- [ ] CI workflow (or a documented manual step) that builds + pushes the image
- [ ] Deploy + smoke test: visit the public URL, walk every page, RSVP a test row, confirm row in RDS
- [ ] Runbook in `infra/README.md`: how to deploy, how to roll back, how to rotate `ACCESS_CODE`
- [ ] `_tasks/014-aws-cfn-infra.md` deleted in the merging PR

## Notes
- Before splitting into sub-tasks: re-spec this task once 013 is done. Scope here will likely be too big for a single PR — split into infra-foundations / app-deploy / dns-and-tls.
- Don't put the dev `ACCESS_CODE` (`letmein`) anywhere near prod. Generate a fresh one and put it directly in Secrets Manager.
