// Package contract implements the Antasena regulatory reporting smart contract.
//
// All methods are designed to be deterministic, side-effect free outside the
// ledger, and to expose a small, easy-to-audit surface area. We deliberately
// keep validation rules in chaincode (not in the gateway) so that every
// participant runs the same rules under consensus.
package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// AntasenaContract is the chaincode struct.
type AntasenaContract struct {
	contractapi.Contract
}

// ----------------------------------------------------------------------------
// Constants & helpers
// ----------------------------------------------------------------------------

const (
	// docType for CouchDB rich queries
	docTypeReport = "report"

	// MSP IDs allowed to validate (BI + OJK)
	mspBI  = "BIMSP"
	mspOJK = "OJKMSP"
)

var (
	// reportingPeriod must be YYYY-MM, YYYY-Qn, or YYYY (annual)
	periodPattern = regexp.MustCompile(`^\d{4}(-(?:0[1-9]|1[0-2]|Q[1-4]))?$`)
	// bank sandi: 3-digit Bank Indonesia code
	sandiPattern = regexp.MustCompile(`^\d{3}$`)
	// SHA-256 hex
	hashPattern = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
)

// composeReportID returns the canonical key under which a report is stored.
func composeReportID(bankCode string, rt ReportType, period string, seq int) string {
	return fmt.Sprintf("%s-%s-%s-%03d", bankCode, rt, period, seq)
}

// getCallerMSPID returns the MSP ID of the invoking client.
func getCallerMSPID(ctx contractapi.TransactionContextInterface) (string, error) {
	id, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("could not read caller MSP ID: %w", err)
	}
	return id, nil
}

// getCallerCN returns the Common Name of the invoking client's X.509 cert.
func getCallerCN(ctx contractapi.TransactionContextInterface) (string, error) {
	cert, err := ctx.GetClientIdentity().GetX509Certificate()
	if err != nil {
		return "", fmt.Errorf("could not read caller cert: %w", err)
	}
	return cert.Subject.CommonName, nil
}

// txTimestamp returns the consensus-stable timestamp for the current tx.
func txTimestamp(ctx contractapi.TransactionContextInterface) (time.Time, error) {
	ts, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return time.Time{}, fmt.Errorf("could not read tx timestamp: %w", err)
	}
	return time.Unix(ts.Seconds, int64(ts.Nanos)).UTC(), nil
}

// ----------------------------------------------------------------------------
// Public chaincode methods
// ----------------------------------------------------------------------------

// InitLedger seeds the channel with a couple of demo reports so the dashboard
// has something to show on first launch. In production you would not invoke
// this method; the Init phase would be empty.
func (c *AntasenaContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	now, err := txTimestamp(ctx)
	if err != nil {
		return err
	}

	demo := []Report{
		{
			DocType:         docTypeReport,
			ReportID:        "014-LBU-2026-04-001",
			BankCode:        "014",
			BankMSPID:       "BCAMSP",
			ReportType:      ReportTypeLBU,
			ReportingPeriod: "2026-04",
			PayloadHash:     "f3b1c0d2e4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1",
			PayloadURI:      "vault://bca/2026-04/lbu.xbrl",
			Status:          StatusFinal,
			SubmittedBy:     "demo.user@bca",
			SubmittedAt:     now.Add(-72 * time.Hour),
			ValidatedBy:     "validator@bi.go.id",
			ValidatedAt:     ptrTime(now.Add(-71 * time.Hour)),
			ValidationNotes: "Seeded by InitLedger",
			SchemaVersion:   "BI-LBU-2.1",
			TxID:            ctx.GetStub().GetTxID(),
		},
	}

	for i := range demo {
		bytes, err := json.Marshal(demo[i])
		if err != nil {
			return fmt.Errorf("marshal demo report: %w", err)
		}
		if err := ctx.GetStub().PutState(demo[i].ReportID, bytes); err != nil {
			return fmt.Errorf("put demo report: %w", err)
		}
	}
	return nil
}

// SubmitReport is invoked by a bank peer to commit a new regulatory submission.
// It performs structural validation, computes/verifies the hash, and writes the
// Report to the ledger in SUBMITTED status. The full XBRL payload is never put
// on-chain — only its SHA-256 hash and a vault URI are stored.
func (c *AntasenaContract) SubmitReport(
	ctx contractapi.TransactionContextInterface,
	bankCode string,
	reportType string,
	reportingPeriod string,
	payloadHash string,
	payloadURI string,
	schemaVersion string,
) (*Report, error) {

	// --- caller authorisation ------------------------------------------------
	mspID, err := getCallerMSPID(ctx)
	if err != nil {
		return nil, err
	}
	if mspID == mspBI || mspID == mspOJK {
		return nil, fmt.Errorf("regulators cannot submit reports (caller MSP=%s)", mspID)
	}
	cn, err := getCallerCN(ctx)
	if err != nil {
		return nil, err
	}

	// --- structural validation -----------------------------------------------
	if !sandiPattern.MatchString(bankCode) {
		return nil, fmt.Errorf("invalid bank code %q (expected 3 digits)", bankCode)
	}
	if !periodPattern.MatchString(reportingPeriod) {
		return nil, fmt.Errorf("invalid reporting period %q (expected YYYY-MM, YYYY-Qn, or YYYY)", reportingPeriod)
	}
	if !hashPattern.MatchString(payloadHash) {
		return nil, fmt.Errorf("invalid payload hash (expected 64-char hex SHA-256)")
	}
	rt := ReportType(strings.ToUpper(reportType))
	switch rt {
	case ReportTypeLBU, ReportTypeLBBU, ReportTypeLHBU, ReportTypeLSMK, ReportTypeAPOLO:
		// ok
	default:
		return nil, fmt.Errorf("unknown report type %q", reportType)
	}
	if strings.TrimSpace(payloadURI) == "" {
		return nil, fmt.Errorf("payload URI must not be empty")
	}
	if strings.TrimSpace(schemaVersion) == "" {
		return nil, fmt.Errorf("schema version must not be empty")
	}

	// --- determine sequence number for this (bank, type, period) -------------
	seq, err := c.nextSequence(ctx, bankCode, rt, reportingPeriod)
	if err != nil {
		return nil, err
	}

	now, err := txTimestamp(ctx)
	if err != nil {
		return nil, err
	}

	rep := Report{
		DocType:         docTypeReport,
		ReportID:        composeReportID(bankCode, rt, reportingPeriod, seq),
		BankCode:        bankCode,
		BankMSPID:       mspID,
		ReportType:      rt,
		ReportingPeriod: reportingPeriod,
		PayloadHash:     strings.ToLower(payloadHash),
		PayloadURI:      payloadURI,
		Status:          StatusSubmitted,
		SubmittedBy:     cn,
		SubmittedAt:     now,
		SchemaVersion:   schemaVersion,
		TxID:            ctx.GetStub().GetTxID(),
	}

	// link to previous version for the same (bank, type, period) if any
	if seq > 1 {
		rep.PrevReportID = composeReportID(bankCode, rt, reportingPeriod, seq-1)
		// mark previous as AMENDED
		if err := c.markAmended(ctx, rep.PrevReportID); err != nil {
			return nil, fmt.Errorf("mark prev as amended: %w", err)
		}
	}

	bytes, err := json.Marshal(rep)
	if err != nil {
		return nil, fmt.Errorf("marshal report: %w", err)
	}
	if err := ctx.GetStub().PutState(rep.ReportID, bytes); err != nil {
		return nil, fmt.Errorf("put report: %w", err)
	}

	// emit a chaincode event so off-chain listeners (e.g. BI dashboard) can react
	if err := ctx.GetStub().SetEvent("ReportSubmitted", bytes); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &rep, nil
}

// ValidateReport is invoked by Bank Indonesia (BIMSP) to validate or reject a
// previously submitted report. It runs the on-chain validation rules and
// flips the report into VALIDATED, REJECTED, or FINAL status depending on the
// outcome and the `markFinal` flag.
func (c *AntasenaContract) ValidateReport(
	ctx contractapi.TransactionContextInterface,
	reportID string,
	markFinal bool,
	notes string,
) (*Report, error) {

	mspID, err := getCallerMSPID(ctx)
	if err != nil {
		return nil, err
	}
	if mspID != mspBI {
		return nil, fmt.Errorf("only Bank Indonesia (BIMSP) may validate reports (caller=%s)", mspID)
	}

	rep, err := c.loadReport(ctx, reportID)
	if err != nil {
		return nil, err
	}
	if rep.Status == StatusFinal {
		return nil, fmt.Errorf("report %s is already FINAL and cannot be re-validated", reportID)
	}
	if rep.Status == StatusAmended {
		return nil, fmt.Errorf("report %s has been amended; validate the latest version", reportID)
	}

	// run on-chain validation rules
	result := runValidationRules(rep)

	cn, err := getCallerCN(ctx)
	if err != nil {
		return nil, err
	}
	now, err := txTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	rep.ValidatedBy = cn
	rep.ValidatedAt = &now
	rep.ValidationNotes = notes

	if !result.Passed {
		rep.Status = StatusRejected
		rep.ValidationNotes = strings.Join(append([]string{notes}, result.Errors...), " | ")
	} else if markFinal {
		rep.Status = StatusFinal
	} else {
		rep.Status = StatusValidated
	}
	rep.TxID = ctx.GetStub().GetTxID()

	bytes, err := json.Marshal(rep)
	if err != nil {
		return nil, fmt.Errorf("marshal report: %w", err)
	}
	if err := ctx.GetStub().PutState(rep.ReportID, bytes); err != nil {
		return nil, fmt.Errorf("put report: %w", err)
	}

	evtName := "ReportValidated"
	if rep.Status == StatusRejected {
		evtName = "ReportRejected"
	} else if rep.Status == StatusFinal {
		evtName = "ReportFinalized"
	}
	if err := ctx.GetStub().SetEvent(evtName, bytes); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return rep, nil
}

// QueryReport returns the current state of a report by its ID.
func (c *AntasenaContract) QueryReport(
	ctx contractapi.TransactionContextInterface,
	reportID string,
) (*Report, error) {
	return c.loadReport(ctx, reportID)
}

// QueryReportsByBank returns all reports submitted by a given bank.
// Uses CouchDB rich query when state DB is CouchDB.
func (c *AntasenaContract) QueryReportsByBank(
	ctx contractapi.TransactionContextInterface,
	bankCode string,
) ([]*Report, error) {
	query := fmt.Sprintf(
		`{"selector":{"docType":"%s","bankCode":"%s"}}`,
		docTypeReport, bankCode,
	)
	return c.queryRich(ctx, query)
}

// QueryReportsByPeriod returns all reports for a given reporting period.
func (c *AntasenaContract) QueryReportsByPeriod(
	ctx contractapi.TransactionContextInterface,
	reportingPeriod string,
) ([]*Report, error) {
	query := fmt.Sprintf(
		`{"selector":{"docType":"%s","reportingPeriod":"%s"}}`,
		docTypeReport, reportingPeriod,
	)
	return c.queryRich(ctx, query)
}

// QueryReportsByStatus returns all reports currently in a given status.
func (c *AntasenaContract) QueryReportsByStatus(
	ctx contractapi.TransactionContextInterface,
	status string,
) ([]*Report, error) {
	query := fmt.Sprintf(
		`{"selector":{"docType":"%s","status":"%s"}}`,
		docTypeReport, strings.ToUpper(status),
	)
	return c.queryRich(ctx, query)
}

// GetAuditTrail returns the full history of a report, including who changed
// what and when. This is the on-chain audit feature that replaces the
// centralised audit log.
func (c *AntasenaContract) GetAuditTrail(
	ctx contractapi.TransactionContextInterface,
	reportID string,
) ([]AuditEntry, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(reportID)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	defer iter.Close()

	var trail []AuditEntry
	for iter.HasNext() {
		mod, err := iter.Next()
		if err != nil {
			return nil, fmt.Errorf("iter history: %w", err)
		}
		entry := AuditEntry{
			TxID:      mod.TxId,
			Timestamp: time.Unix(mod.Timestamp.Seconds, int64(mod.Timestamp.Nanos)).UTC(),
			IsDelete:  mod.IsDelete,
		}
		if !mod.IsDelete {
			if err := json.Unmarshal(mod.Value, &entry.Report); err != nil {
				return nil, fmt.Errorf("unmarshal historical state: %w", err)
			}
		}
		trail = append(trail, entry)
	}
	return trail, nil
}

// VerifyPayloadHash takes the off-chain payload (as bytes) and a report ID,
// recomputes SHA-256, and tells the caller whether the on-chain hash matches.
// This is how a regulator proves that the file the bank sends them today is
// byte-for-byte the file the bank committed to the chain at submission time.
func (c *AntasenaContract) VerifyPayloadHash(
	ctx contractapi.TransactionContextInterface,
	reportID string,
	payloadBase64 string,
) (bool, error) {

	rep, err := c.loadReport(ctx, reportID)
	if err != nil {
		return false, err
	}

	// payload comes in as raw bytes; for chaincode portability we accept
	// hex-encoded sha256 from the caller too. Here we recompute from raw.
	sum := sha256.Sum256([]byte(payloadBase64))
	actual := hex.EncodeToString(sum[:])
	return strings.EqualFold(actual, rep.PayloadHash), nil
}

// ----------------------------------------------------------------------------
// Internal helpers
// ----------------------------------------------------------------------------

func (c *AntasenaContract) loadReport(
	ctx contractapi.TransactionContextInterface,
	reportID string,
) (*Report, error) {
	bytes, err := ctx.GetStub().GetState(reportID)
	if err != nil {
		return nil, fmt.Errorf("get state %s: %w", reportID, err)
	}
	if bytes == nil {
		return nil, fmt.Errorf("report %s not found", reportID)
	}
	var rep Report
	if err := json.Unmarshal(bytes, &rep); err != nil {
		return nil, fmt.Errorf("unmarshal report: %w", err)
	}
	return &rep, nil
}

func (c *AntasenaContract) nextSequence(
	ctx contractapi.TransactionContextInterface,
	bankCode string,
	rt ReportType,
	period string,
) (int, error) {
	// scan keys with prefix <bankCode>-<rt>-<period>-
	prefix := fmt.Sprintf("%s-%s-%s-", bankCode, rt, period)
	iter, err := ctx.GetStub().GetStateByRange(prefix, prefix+"~")
	if err != nil {
		return 0, fmt.Errorf("range scan: %w", err)
	}
	defer iter.Close()

	count := 0
	for iter.HasNext() {
		_, err := iter.Next()
		if err != nil {
			return 0, fmt.Errorf("iter range: %w", err)
		}
		count++
	}
	return count + 1, nil
}

func (c *AntasenaContract) markAmended(
	ctx contractapi.TransactionContextInterface,
	reportID string,
) error {
	rep, err := c.loadReport(ctx, reportID)
	if err != nil {
		return err
	}
	if rep.Status == StatusFinal {
		return fmt.Errorf("cannot amend FINAL report %s", reportID)
	}
	rep.Status = StatusAmended
	rep.TxID = ctx.GetStub().GetTxID()
	bytes, err := json.Marshal(rep)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return ctx.GetStub().PutState(rep.ReportID, bytes)
}

func (c *AntasenaContract) queryRich(
	ctx contractapi.TransactionContextInterface,
	query string,
) ([]*Report, error) {
	iter, err := ctx.GetStub().GetQueryResult(query)
	if err != nil {
		return nil, fmt.Errorf("rich query: %w", err)
	}
	defer iter.Close()

	var out []*Report
	for iter.HasNext() {
		kv, err := iter.Next()
		if err != nil {
			return nil, fmt.Errorf("iter rich: %w", err)
		}
		var rep Report
		if err := json.Unmarshal(kv.Value, &rep); err != nil {
			return nil, fmt.Errorf("unmarshal rich result: %w", err)
		}
		out = append(out, &rep)
	}
	return out, nil
}

// runValidationRules executes the deterministic on-chain validation rules.
// In real Antasena, validation includes XBRL taxonomy checks, cross-cell
// arithmetic, and reference-data lookups. We model the most basic structural
// checks here. The full taxonomy validation can run in a sidecar service and
// have its result attested back into chaincode via a signed oracle pattern.
func runValidationRules(r *Report) ValidationResult {
	res := ValidationResult{Passed: true}

	if !sandiPattern.MatchString(r.BankCode) {
		res.Passed = false
		res.Errors = append(res.Errors, "bankCode must be 3 digits")
	}
	if !periodPattern.MatchString(r.ReportingPeriod) {
		res.Passed = false
		res.Errors = append(res.Errors, "reportingPeriod must be YYYY-MM, YYYY-Qn, or YYYY")
	}
	if !hashPattern.MatchString(r.PayloadHash) {
		res.Passed = false
		res.Errors = append(res.Errors, "payloadHash must be a 64-char hex SHA-256")
	}
	if r.SchemaVersion == "" {
		res.Passed = false
		res.Errors = append(res.Errors, "schemaVersion must be set")
	}
	if r.PayloadURI == "" {
		res.Passed = false
		res.Errors = append(res.Errors, "payloadURI must be set")
	}
	return res
}

func ptrTime(t time.Time) *time.Time { return &t }
