# Antasena Chaincode

The smart contract that powers Hephaestus. Written in Go using `fabric-contract-api-go` v1.2.

## Public methods

| Method | Caller | Purpose |
|---|---|---|
| `InitLedger` | system | Seeds demo data on channel creation. |
| `SubmitReport` | bank peer | Commit a new report (hash + metadata). Auto-amends prior version. |
| `ValidateReport` | BIMSP only | Run on-chain validation; flip status to VALIDATED / REJECTED / FINAL. |
| `QueryReport` | any | Get the current state of a report by ID. |
| `QueryReportsByBank` | any | Rich query (CouchDB) — all reports for a bank. |
| `QueryReportsByPeriod` | any | Rich query — all reports for a reporting period. |
| `QueryReportsByStatus` | any | Rich query — all reports in a given status. |
| `GetAuditTrail` | any | Full immutable history (`GetHistoryForKey`) for a report. |
| `VerifyPayloadHash` | any | Re-hashes an off-chain payload and compares to the on-chain hash. |

## Lifecycle

```
DRAFT  ──submit──►  SUBMITTED  ──validate──►  VALIDATED  ──finalize──►  FINAL
                                  │                                       
                                  ├──reject──►  REJECTED  ──resubmit──►  SUBMITTED (n+1)
                                  │                                       (prev → AMENDED)
                                  └──amend──►   AMENDED
```

## Authorisation rules

- **Submitting a report** — caller MSP must be a bank MSP (anything except `BIMSP` and `OJKMSP`).
- **Validating a report** — caller MSP must be exactly `BIMSP`.
- **Reading reports** — any channel member can read; private collections can be added later for sensitive cells.

## On-chain validation

`runValidationRules` enforces:

- 3-digit bank sandi (`bankCode`)
- Reporting period in `YYYY-MM`, `YYYY-Qn`, or `YYYY` form
- 64-char hex SHA-256 payload hash
- Non-empty schema version and payload URI

Heavyweight XBRL taxonomy validation is intentionally **out of scope** for chaincode — that runs in a sidecar service per organisation and the result is attested back into chaincode via a signed event in a future iteration.

## Building

```bash
go mod tidy
go build ./...
go test ./test/... -v -cover
```

The chaincode is packaged and deployed by `network/scripts/deploy-chaincode.sh`.
