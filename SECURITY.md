# Security Policy

## Supported versions

Hephaestus is currently a proof-of-concept. Only the `main` branch receives security review.

| Version | Status |
|---|---|
| `main` (PoC)  | ✅ Active review |
| Older tags    | ❌ No security support |

## Reporting a vulnerability

**Please do NOT open public GitHub issues for security vulnerabilities.**

Instead, email `security@hephaestus.example` (replace with project maintainer email) with:

1. A description of the vulnerability and its potential impact
2. Steps to reproduce (proof-of-concept welcome)
3. Suggested mitigation, if you have one

You should receive an acknowledgement within **3 business days**. We aim to triage critical issues within **7 days** and release a fix within **30 days** for high/critical severity.

## Scope

In scope:
- Chaincode logic flaws (state inconsistency, access-control bypass, replay)
- Fabric channel/MSP misconfiguration in the reference network
- Authentication/authorisation bypass in the API gateway
- Cryptographic weaknesses in our code (not in upstream Fabric or Go stdlib)
- Container or Ansible role misconfigurations that escalate privilege

Out of scope:
- Denial-of-service from a malicious organisation that already has channel access (BFT guarantees apply)
- Issues in upstream Hyperledger Fabric — please report to the Hyperledger project directly
- Issues in third-party dependencies — please report upstream

## Hall of fame

Researchers who responsibly disclose valid vulnerabilities will be credited (with permission) in this section.
