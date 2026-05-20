# 015b — Network stack: VPC + subnets

## Why
Foundation layer. VPC, subnets, IGW, and route tables — the shared network that both the data and app stacks build on. Owned by its own stack so it isn't tied to either's lifecycle.

## Acceptance criteria
- [ ] `infra/network.yaml` defines stack `wedding-network` with:
  - VPC `10.0.0.0/16`, DNS hostnames + DNS support enabled.
  - 2 AZs: us-west-2a, us-west-2b.
  - 2 public subnets, one per AZ (`10.0.0.0/20`, `10.0.16.0/20`), `MapPublicIpOnLaunch: true`.
  - 2 private subnets, one per AZ (`10.0.32.0/20`, `10.0.48.0/20`).
  - IGW attached to the VPC.
  - Public route table with default route to the IGW; both public subnets associated.
  - Private route table (no default route — RDS doesn't need outbound; we'd add a NAT only if we ever wanted private EC2); both private subnets associated.
- [ ] `bin/deploy-network.sh` deploys the stack (cfn-lint → validate-template → deploy → outputs).
- [ ] `cfn-lint infra/network.yaml` exit 0; `./bin/deploy-network.sh` succeeds end-to-end against AWS.
- [ ] Stack exports for downstream stacks: `wedding-network-VpcId`, `-VpcCidr`, `-PublicSubnetIds` (comma-list of both), `-PrivateSubnetIds` (comma-list of both).
- [ ] `infra/README.md` updated with the new stack + deploy command.
- [ ] `_tasks/015b-network-stack.md` deleted in the merging PR.

## Notes
- Region `us-west-2`, account `755621201894`.
- Public subnets will host ALB + EC2 (015d); private subnets will host RDS (015c).
- No NAT gateway — saves ~$32/mo. EC2 in public subnets gets outbound to AWS APIs via the IGW directly; ingress is SG-locked to the ALB.
- CIDR layout chosen for simple readability: `10.0.0.0/16` parent, `/20` blocks (4096 IPs each — more than we need, but easy to reason about).
- Both data and app stacks `Fn::ImportValue` from here.
