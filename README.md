<div align="center">

# 🔨 Hephaestus

### Permissioned Blockchain for Indonesian Bank Regulatory Reporting
### *Blockchain Permissioned untuk Pelaporan Regulator Perbankan Indonesia*

A Hyperledger Fabric based proof-of-concept that re-imagines **Bank Indonesia's Antasena** reporting platform on a tamper-evident, distributed ledger.

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Hyperledger Fabric](https://img.shields.io/badge/Hyperledger-Fabric%202.5-2F3134)](https://www.hyperledger.org/use/fabric)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)](docker/)
[![Status](https://img.shields.io/badge/Status-PoC-orange)]()

[English](#english) · [Bahasa Indonesia](#bahasa-indonesia) · [Architecture](docs/en/ARCHITECTURE.md) · [Whitepaper](docs/en/WHITEPAPER.md) · [Deployment](docs/en/DEPLOYMENT.md)

</div>

---

## English

### What is Hephaestus?

**Hephaestus** is a permissioned blockchain platform that demonstrates how Bank Indonesia's *Antasena* — the integrated reporting system that consolidates regulatory submissions from every commercial bank in Indonesia — can be re-architected on a distributed ledger.

Today, Antasena receives report files (XBRL/XML) from 100+ banks. Reports are validated centrally, stored centrally, and audited centrally. Hephaestus replaces the centralised pipeline with a **multi-organisation Hyperledger Fabric network**, where:

- Every bank operates a **peer node** that endorses its own submissions.
- Bank Indonesia and OJK operate **orderer/observer nodes** for consensus and oversight.
- Reports are hashed on-chain (full payloads stay off-chain in the bank's vault).
- Validation rules run as **chaincode** — auditable, versioned, and identical for every participant.
- Every state change is **immutable, timestamped, and cryptographically signed**.

> **Why "Hephaestus"?** In Greek mythology, Hephaestus is the divine smith who forges the unbreakable. We forge tamper-evident regulatory truth.

### Key features

- 🏦 **Multi-org permissioned network** — 1 BI orderer cluster + N bank peers + 1 OJK observer
- 📜 **Antasena-style chaincode** — Submit, Validate, Query, Lifecycle, Audit (Go)
- 🔐 **Fabric CA + MSP per bank** — every signature traceable to a real-world identity
- 🐳 **Docker Compose PoC** — `make up` brings the full network online in <2 minutes
- 🛰️ **Production deploy** — Ansible playbook for multi-VM RHEL/Rocky installation
- 🌐 **REST API gateway** — Node.js + Express for bank back-office integration
- 📊 **Reference dashboard** — minimal web UI to submit, validate, and audit reports

### Quick start (Docker Compose)

```bash
git clone https://github.com/<your-org>/hephaestus.git
cd hephaestus
make up           # spins up 5 banks + BI + OJK + 3 orderers
make demo         # submits a sample LBU report and validates it
make logs         # tail peer & orderer logs
make down         # tear down + clean state
```

Open `http://localhost:8080` for the dashboard. API docs at `http://localhost:3000/api/docs`.

### Production multi-node deployment

```bash
cd ansible
# edit inventory/hosts.yml with your real VMs
ansible-playbook -i inventory/hosts.yml site.yml
```

See [docs/en/DEPLOYMENT.md](docs/en/DEPLOYMENT.md) for hardware specs, network ports, hardening, and HA topology.

### Project structure

```
hephaestus/
├── chaincode/antasena/       Go chaincode (smart contract)
├── network/                  Fabric crypto-config, configtx, channel artifacts
├── docker/                   Docker Compose PoC for 5-bank consortium
├── api-gateway/              Node.js REST API
├── web-dashboard/            Minimal HTML/JS reference dashboard
├── ansible/                  Multi-VM production deployment
├── docs/en/  docs/id/        Bilingual documentation
├── pptx/                     Presentation deck
├── linkedin/                 Launch post copy
└── samples/                  Sample LBU report payloads
```

### Documentation

| Document | Purpose |
|---|---|
| [Whitepaper](docs/en/WHITEPAPER.md) | Vision, problem, solution, governance |
| [Architecture](docs/en/ARCHITECTURE.md) | Network topology, data flow, threat model |
| [Validator Spec](docs/en/VALIDATOR_SPEC.md) | Hardware, OS, network requirements per node |
| [Deployment](docs/en/DEPLOYMENT.md) | Step-by-step install (Docker + Ansible) |
| [Pros & Cons](docs/en/PROS_CONS.md) | Honest assessment vs. status quo |
| [FAQ](docs/en/FAQ.md) | Common questions |

### Roadmap

- [x] Phase 1 — PoC chaincode, Docker Compose, REST API
- [ ] Phase 2 — Hardware Security Module (HSM) integration for key custody
- [ ] Phase 3 — Off-chain XBRL vault with IPFS + private collections
- [ ] Phase 4 — Cross-chain bridge to OJK SLIK / SIPESAT
- [ ] Phase 5 — Pilot with 3-5 commercial banks + BI sandbox

### License

Apache License 2.0 — see [LICENSE](LICENSE).

### Disclaimer

This is an **independent proof-of-concept** and is not affiliated with, endorsed by, or commissioned by Bank Indonesia, OJK, or any commercial bank. "Antasena" is referenced as a real-world target use case to ground the design in a recognisable problem domain.

---

## Bahasa Indonesia

### Apa itu Hephaestus?

**Hephaestus** adalah platform blockchain *permissioned* yang mendemonstrasikan bagaimana *Antasena* — sistem pelaporan terintegrasi milik Bank Indonesia yang mengkonsolidasikan laporan dari seluruh bank umum di Indonesia — dapat diarsitektur ulang di atas *distributed ledger*.

Saat ini Antasena menerima file laporan (XBRL/XML) dari 100+ bank. Laporan divalidasi terpusat, disimpan terpusat, dan diaudit terpusat. Hephaestus menggantikan pipeline terpusat tersebut dengan **jaringan Hyperledger Fabric multi-organisasi**, di mana:

- Setiap bank menjalankan **peer node** yang meng-endorse laporannya sendiri.
- Bank Indonesia dan OJK menjalankan **orderer/observer node** untuk konsensus dan pengawasan.
- Laporan di-hash on-chain (payload lengkap tetap di vault bank).
- Aturan validasi berjalan sebagai **chaincode** — bisa diaudit, di-versioning, dan identik untuk semua peserta.
- Setiap perubahan state bersifat **immutable, ber-timestamp, dan ditandatangani secara kriptografis**.

> **Kenapa nama "Hephaestus"?** Dalam mitologi Yunani, Hephaestus adalah dewa pandai besi yang menempa hal-hal yang tak terhancurkan. Kami menempa kebenaran regulatori yang tamper-evident.

### Fitur utama

- 🏦 **Jaringan permissioned multi-organisasi** — 1 cluster orderer BI + N peer bank + 1 observer OJK
- 📜 **Chaincode bergaya Antasena** — Submit, Validate, Query, Lifecycle, Audit (Go)
- 🔐 **Fabric CA + MSP per bank** — setiap signature bisa ditelusuri ke identitas dunia nyata
- 🐳 **PoC Docker Compose** — `make up` menjalankan full network dalam <2 menit
- 🛰️ **Deploy produksi** — Ansible playbook untuk instalasi multi-VM RHEL/Rocky
- 🌐 **REST API gateway** — Node.js + Express untuk integrasi back-office bank
- 📊 **Dashboard referensi** — UI web minimal untuk submit, validasi, dan audit

### Mulai cepat

```bash
git clone https://github.com/<your-org>/hephaestus.git
cd hephaestus
make up           # menjalankan 5 bank + BI + OJK + 3 orderer
make demo         # submit laporan LBU contoh dan validasi
make logs         # tail log peer & orderer
make down         # hentikan dan bersihkan state
```

Buka `http://localhost:8080` untuk dashboard. API docs di `http://localhost:3000/api/docs`.

### Deploy multi-node produksi

```bash
cd ansible
# edit inventory/hosts.yml dengan VM Anda
ansible-playbook -i inventory/hosts.yml site.yml
```

Lihat [docs/id/DEPLOYMENT.md](docs/id/DEPLOYMENT.md) untuk spesifikasi hardware, port jaringan, hardening, dan topologi HA.

### Dokumentasi

| Dokumen | Tujuan |
|---|---|
| [Whitepaper](docs/id/WHITEPAPER.md) | Visi, problem, solusi, tata kelola |
| [Arsitektur](docs/id/ARCHITECTURE.md) | Topologi jaringan, alur data, threat model |
| [Spesifikasi Validator](docs/id/VALIDATOR_SPEC.md) | Hardware, OS, kebutuhan jaringan per node |
| [Deployment](docs/id/DEPLOYMENT.md) | Instalasi step-by-step (Docker + Ansible) |
| [Kelebihan & Kekurangan](docs/id/PROS_CONS.md) | Penilaian jujur vs status quo |
| [FAQ](docs/id/FAQ.md) | Pertanyaan umum |

### Disclaimer

Ini adalah *proof-of-concept independen* dan tidak berafiliasi dengan, di-endorse oleh, atau ditugaskan oleh Bank Indonesia, OJK, maupun bank komersial manapun. "Antasena" disebut sebagai use case dunia nyata untuk membumikan desain pada problem domain yang dikenal.

---

<div align="center">

**Forging tamper-evident regulatory truth.**

Built with 🔨 for the Indonesian banking ecosystem.

</div>
