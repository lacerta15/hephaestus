# Sample payloads

`sample-lbu-2026-04.xml` — a heavily-simplified stub of a Bank Indonesia LBU report. Use it with `make demo` or with the dashboard's "upload" button to test the submission flow end-to-end.

In production these payloads would be full XBRL documents conforming to the BI-LBU-2.1 taxonomy. The chain only stores the SHA-256 of the file, so the actual schema is opaque to chaincode.
