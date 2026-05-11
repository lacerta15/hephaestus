package contract

import "time"

// ReportType enumerates the kinds of regulatory submissions Antasena handles.
// We model the most common Bank Indonesia / OJK report types.
type ReportType string

const (
	// LBU — Laporan Bulanan Bank Umum (monthly commercial bank report)
	ReportTypeLBU ReportType = "LBU"
	// LBBU — Laporan Berkala Bank Umum (periodic commercial bank report)
	ReportTypeLBBU ReportType = "LBBU"
	// LHBU — Laporan Harian Bank Umum (daily commercial bank report)
	ReportTypeLHBU ReportType = "LHBU"
	// LSMK — Laporan Stabilitas Moneter & Keuangan (monetary & financial stability)
	ReportTypeLSMK ReportType = "LSMK"
	// APOLO — Aplikasi Pelaporan Online (OJK)
	ReportTypeAPOLO ReportType = "APOLO"
)

// ReportStatus tracks where a report is in its regulatory lifecycle.
type ReportStatus string

const (
	StatusDraft     ReportStatus = "DRAFT"     // submitted by bank, not yet on-chain finalised
	StatusSubmitted ReportStatus = "SUBMITTED" // committed by bank peer
	StatusValidated ReportStatus = "VALIDATED" // BI validation chaincode passed
	StatusRejected  ReportStatus = "REJECTED"  // BI rejected — bank must resubmit
	StatusFinal     ReportStatus = "FINAL"     // accepted by BI, immutable
	StatusAmended   ReportStatus = "AMENDED"   // superseded by a later submission
)

// Report is the on-chain representation of a regulatory submission.
//
// Note: the heavyweight XBRL payload itself is NOT stored on-chain. We
// store only its SHA-256 hash plus the metadata needed for routing,
// auditing, and validation. The bank's vault retains the original payload.
type Report struct {
	DocType         string       `json:"docType"`         // discriminator for CouchDB rich queries
	ReportID        string       `json:"reportId"`        // composite: <bankCode>-<type>-<period>-<seq>
	BankCode        string       `json:"bankCode"`        // sandi bank, e.g. "014" for BCA
	BankMSPID       string       `json:"bankMSPID"`       // Fabric MSP ID, e.g. "BCAMSP"
	ReportType      ReportType   `json:"reportType"`      // LBU, LBBU, etc.
	ReportingPeriod string       `json:"reportingPeriod"` // e.g. "2026-04" (monthly), "2026-Q1"
	PayloadHash     string       `json:"payloadHash"`     // SHA-256 of the off-chain XBRL/CSV
	PayloadURI      string       `json:"payloadURI"`      // pointer to off-chain vault (e.g. ipfs://, s3://)
	Status          ReportStatus `json:"status"`
	SubmittedBy     string       `json:"submittedBy"`     // X.509 CN of the bank user
	SubmittedAt     time.Time    `json:"submittedAt"`
	ValidatedBy     string       `json:"validatedBy,omitempty"`     // BI user / validator MSP
	ValidatedAt     *time.Time   `json:"validatedAt,omitempty"`
	ValidationNotes string       `json:"validationNotes,omitempty"`
	SchemaVersion   string       `json:"schemaVersion"`   // taxonomy version, e.g. "BI-LBU-2.1"
	PrevReportID    string       `json:"prevReportID,omitempty"` // chains amendments
	TxID            string       `json:"txId"`            // Fabric tx that produced this state
}

// ValidationResult is what the chaincode-side validation routine returns.
type ValidationResult struct {
	Passed bool     `json:"passed"`
	Errors []string `json:"errors,omitempty"`
}

// AuditEntry represents one historical state of a Report, returned by GetAuditTrail.
// It mirrors the structure produced by Fabric's GetHistoryForKey API.
type AuditEntry struct {
	TxID      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
	Report    Report    `json:"report"`
}
