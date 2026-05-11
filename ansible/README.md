# Production deployment with Ansible

This playbook deploys Hephaestus to real RHEL 8/9 (or Rocky 8/9) hosts. It is the production counterpart to the `make up` Docker Compose flow.

## Recommended topology

| Group | Hosts | Purpose |
|---|---|---|
| `orderers` | 3 | Bank Indonesia RAFT cluster |
| `bank_peers` | 5 (one per bank) | Peer + CouchDB + chaincode runtime |
| `observers` | 1 | OJK read-only peer |
| `ca` | 7 | One Fabric CA per organisation |

In production each organisation runs its **own** infrastructure — the playbook is structured so each bank operates only the hosts it owns. The `inventory/hosts.yml` example is just a single-operator view for testing.

## Hardware spec per node

See `docs/en/VALIDATOR_SPEC.md` for the recommended specs. Quick summary:

| Role | vCPU | RAM | Disk | Network |
|---|---|---|---|---|
| Orderer | 8 | 16 GB | 500 GB SSD | 1 Gbps, low latency to other orderers |
| Peer | 16 | 32 GB | 1 TB SSD | 1 Gbps |
| CA | 4 | 8 GB | 100 GB SSD | 100 Mbps |

## Run

```bash
ansible-galaxy collection install ansible.posix community.general
ansible-playbook -i inventory/hosts.yml site.yml --check    # dry-run
ansible-playbook -i inventory/hosts.yml site.yml
```

Tag-targeted runs:

```bash
ansible-playbook -i inventory/hosts.yml site.yml --tags peer
ansible-playbook -i inventory/hosts.yml site.yml --tags gateway
```

## Hardening checklist

The roles take care of:
- ✅ SELinux enforcing
- ✅ firewalld with explicit allow-list
- ✅ chronyd (NTP) — Fabric requires <500ms clock drift
- ✅ auditd active
- ✅ systemd `NoNewPrivileges`, `ProtectSystem=strict`, `PrivateTmp`
- ✅ Service runs as `hephaestus` (no root)

You should additionally:
- 🔐 Replace the default CA bootstrap password
- 🔐 Move TLS keys to an HSM (PKCS#11 — see `docs/en/SECURITY.md`)
- 🔐 Run a reverse proxy (nginx/HAProxy) with mTLS in front of the API gateway
- 🔐 Ship `/var/log/hephaestus/*.log` to a SIEM
