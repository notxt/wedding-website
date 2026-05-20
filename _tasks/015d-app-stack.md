# 015d — App stack: ALB + DNS + TLS + EC2 ASG + CodeDeploy + first live deploy

## Why
Bring the site live at `https://michaelandadrien4ever.com`. ALB-terminated TLS, single-instance self-healing ASG, CodeDeploy (IN_PLACE) for app revisions. Imports the network from 015b (`wedding-network`) and the data layer from 015c (`wedding-data`).

## Acceptance criteria
- [ ] `infra/app.yaml` defines stack `wedding-app` with:
  - ACM cert for `michaelandadrien4ever.com` + `www.michaelandadrien4ever.com` (SAN), DNS-validated against the existing Route53 hosted zone.
  - ALB security group (ingress 80 + 443 from world); App SG (ingress 8080 from ALB SG; egress 443 anywhere, 5432 to DB SG); cross-stack ingress on the imported DB SG from the App SG.
  - ALB across the imported public subnets. HTTPS:443 listener (ACM cert attached) forwarding to the target group. HTTP:80 listener: 301 → HTTPS.
  - Target group HTTP/8080, health check `GET /healthz`, threshold 2/2, interval 15s.
  - Listener rule on :443 matching `Host: michaelandadrien4ever.com` **OR** `www.michaelandadrien4ever.com` → target group (both hosts serve the same content).
  - Route53 A-record aliases for **both** apex and `www.` pointing at the ALB.
  - IAM instance role: `secretsmanager:GetSecretValue` on the three imported secret ARNs, S3 read on the bootstrap CFN bucket, `AmazonSSMManagedInstanceCore` (so Session Manager replaces SSH), CloudWatch Logs write.
  - CloudWatch log group `/wedding/app`, 30-day retention.
  - Launch template: AL2023 arm64 via SSM parameter `/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64`; instance type `t4g.micro`. Userdata installs the CodeDeploy agent + CloudWatch agent (tailing `journalctl -u wedding-website -o cat` into `/wedding/app`), fetches secrets, writes `/etc/wedding-website/env` (`DATABASE_URL=postgres://...?sslmode=require`, `ACCESS_CODE`, `SESSION_SECRET`, `PORT=8080`), and enables the systemd unit (unit itself is installed by CodeDeploy, not userdata).
  - ASG `MinSize=MaxSize=DesiredCapacity=1`, public subnets, target-group attached, `UpdatePolicy: AutoScalingRollingUpdate` so launch-template changes roll the instance.
  - CodeDeploy `Application` + `DeploymentGroup` (IN_PLACE, `CodeDeployDefault.AllAtOnce`, ASG-targeted, service role with the standard CodeDeploy managed policy).
- [ ] `deploy/` bundle committed:
  - `appspec.yml` — CodeDeploy hooks → scripts.
  - `wedding-website.service` — systemd unit. `ExecStart=/opt/wedding-website/server`, `EnvironmentFile=/etc/wedding-website/env`, `User=wedding`, `Restart=always`, `RestartSec=5`, output to journal.
  - `scripts/install.sh`, `start.sh`, `stop.sh` (no-op if not installed), `validate.sh` (polls `http://127.0.0.1:8080/healthz`).
- [ ] `bin/deploy-app-stack.sh` deploys the CFN stack (cfn-lint → validate-template → deploy → outputs).
- [ ] `bin/deploy-app.sh` (app revision script): `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o build/server ./cmd/server`; zip with `deploy/*`; `aws s3 cp` to `s3://${CfnBucket}/codedeploy-revisions/wedding-website-<git-sha>.zip`; `aws deploy create-deployment`; `aws deploy wait deployment-successful`.
- [ ] `cfn-lint infra/app.yaml` exit 0; both deploy scripts succeed end-to-end against AWS.
- [ ] `https://michaelandadrien4ever.com` serves the actual site (every route renders), valid TLS chain, `http://...` 301s, `www.` resolves and serves identically.
- [ ] Test RSVP submitted; row appears in `rsvps` (verified via Session Manager port-forward + `psql` against the RDS endpoint).
- [ ] Idempotency: re-run `bin/deploy-app.sh` with no code change → CodeDeploy reports a successful no-op deploy.
- [ ] Self-healing: `aws ec2 stop-instances` on the running instance → ASG replaces it; site recovers without manual intervention.
- [ ] `infra/README.md` updated with the app stack + deploy command + revision-deploy command.
- [ ] `SPEC.md` updated to note the deployment shape (ALB + ASG + CodeDeploy + RDS).
- [ ] `_tasks/015d-app-stack.md` deleted in the merging PR.

## Notes
- Region `us-west-2`, account `755621201894`.
- Imports from `wedding-network`: `VpcId`, `PublicSubnetIds`.
- Imports from `wedding-data`: `DbEndpoint`, `DbPort`, `DbName`, `DbMasterSecretArn`, `AccessCodeSecretArn`, `SessionSecretArn`, `DbSecurityGroupId`.
- Server is fully self-contained (templates + migrations + static via `embed.FS`). Deploy bundle = `server` binary + `appspec.yml` + `wedding-website.service` + `scripts/*.sh`.
- Server runtime is already correct for this topology: binds `0.0.0.0:$PORT`, `/healthz` pings DB (200/503), respects `X-Forwarded-For` / `X-Forwarded-Proto`, handles SIGTERM with 10s drain. **No code changes needed.**
- ALB terminates TLS; server speaks plain HTTP on 8080 internally.
- Route53 hosted zone for `michaelandadrien4ever.com` already exists; the deploy script looks up the zone ID via `aws route53 list-hosted-zones-by-name --dns-name michaelandadrien4ever.com`.
- CodeDeploy revisions live under `s3://${bootstrap.CfnBucket}/codedeploy-revisions/` — the bootstrap stack's bucket (`wedding-bootstrap-CfnBucketName`).
