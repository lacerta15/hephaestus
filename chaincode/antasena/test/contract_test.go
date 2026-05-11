// Package test contains unit tests for the Antasena chaincode.
//
// These tests use the lightweight mock harness in mock_stub.go. They are
// intentionally focused on the deterministic, business-logic surface of the
// contract (validation rules, lifecycle transitions, query shape). Integration
// tests that exercise the real Fabric network live under network/scripts/
// and are run by `make demo`.
package test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hephaestus-project/chaincode/antasena/contract"
)

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestSubmitReport_HappyPath(t *testing.T) {
	stub := NewMockStub("BCAMSP", "alice@bca")
	cc := &contract.AntasenaContract{}

	rep, err := cc.SubmitReport(stub.Ctx(), "014", "LBU", "2026-04",
		sha256Hex("hello-world"), "vault://bca/2026-04.xbrl", "BI-LBU-2.1")
	require.NoError(t, err)
	require.NotNil(t, rep)

	assert.Equal(t, "014", rep.BankCode)
	assert.Equal(t, contract.ReportTypeLBU, rep.ReportType)
	assert.Equal(t, contract.StatusSubmitted, rep.Status)
	assert.Equal(t, "BCAMSP", rep.BankMSPID)
	assert.Equal(t, "alice@bca", rep.SubmittedBy)
	assert.True(t, strings.HasPrefix(rep.ReportID, "014-LBU-2026-04-"))
}

func TestSubmitReport_RejectsRegulators(t *testing.T) {
	stub := NewMockStub("BIMSP", "validator@bi.go.id")
	cc := &contract.AntasenaContract{}

	_, err := cc.SubmitReport(stub.Ctx(), "014", "LBU", "2026-04",
		sha256Hex("x"), "vault://", "BI-LBU-2.1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "regulators cannot submit")
}

func TestSubmitReport_ValidatesInput(t *testing.T) {
	stub := NewMockStub("BCAMSP", "alice@bca")
	cc := &contract.AntasenaContract{}

	tests := []struct {
		name    string
		bank    string
		rtype   string
		period  string
		hash    string
		uri     string
		schema  string
		wantErr string
	}{
		{"bad bank code", "12", "LBU", "2026-04", sha256Hex("x"), "vault://", "v1", "invalid bank code"},
		{"bad period", "014", "LBU", "April-2026", sha256Hex("x"), "vault://", "v1", "invalid reporting period"},
		{"bad hash", "014", "LBU", "2026-04", "not-a-hash", "vault://", "v1", "invalid payload hash"},
		{"bad type", "014", "FOO", "2026-04", sha256Hex("x"), "vault://", "v1", "unknown report type"},
		{"empty uri", "014", "LBU", "2026-04", sha256Hex("x"), "  ", "v1", "payload URI"},
		{"empty schema", "014", "LBU", "2026-04", sha256Hex("x"), "vault://", "", "schema version"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := cc.SubmitReport(stub.Ctx(), tc.bank, tc.rtype, tc.period, tc.hash, tc.uri, tc.schema)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateReport_OnlyBI(t *testing.T) {
	bcaStub := NewMockStub("BCAMSP", "alice@bca")
	cc := &contract.AntasenaContract{}

	rep, err := cc.SubmitReport(bcaStub.Ctx(), "014", "LBU", "2026-04",
		sha256Hex("x"), "vault://", "v1")
	require.NoError(t, err)

	// non-BI cannot validate
	_, err = cc.ValidateReport(bcaStub.Ctx(), rep.ReportID, true, "ok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only Bank Indonesia")

	// BI can validate
	biStub := bcaStub.WithIdentity("BIMSP", "validator@bi.go.id")
	out, err := cc.ValidateReport(biStub.Ctx(), rep.ReportID, true, "passed")
	require.NoError(t, err)
	assert.Equal(t, contract.StatusFinal, out.Status)
	assert.Equal(t, "validator@bi.go.id", out.ValidatedBy)
}

func TestValidateReport_RejectsFinal(t *testing.T) {
	stub := NewMockStub("BCAMSP", "alice@bca")
	cc := &contract.AntasenaContract{}
	rep, err := cc.SubmitReport(stub.Ctx(), "014", "LBU", "2026-04",
		sha256Hex("x"), "vault://", "v1")
	require.NoError(t, err)

	bi := stub.WithIdentity("BIMSP", "validator@bi.go.id")
	_, err = cc.ValidateReport(bi.Ctx(), rep.ReportID, true, "ok")
	require.NoError(t, err)

	// second validation must fail because report is FINAL
	_, err = cc.ValidateReport(bi.Ctx(), rep.ReportID, false, "again")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already FINAL")
}

func TestSubmitReport_AmendsPrevious(t *testing.T) {
	stub := NewMockStub("BCAMSP", "alice@bca")
	cc := &contract.AntasenaContract{}

	r1, err := cc.SubmitReport(stub.Ctx(), "014", "LBU", "2026-04",
		sha256Hex("v1"), "vault://1", "BI-LBU-2.1")
	require.NoError(t, err)

	r2, err := cc.SubmitReport(stub.Ctx(), "014", "LBU", "2026-04",
		sha256Hex("v2"), "vault://2", "BI-LBU-2.1")
	require.NoError(t, err)

	assert.Equal(t, r1.ReportID, r2.PrevReportID)

	// r1 should now be marked AMENDED on the ledger
	got, err := cc.QueryReport(stub.Ctx(), r1.ReportID)
	require.NoError(t, err)
	assert.Equal(t, contract.StatusAmended, got.Status)
}

func TestQueryReport_NotFound(t *testing.T) {
	stub := NewMockStub("BCAMSP", "alice@bca")
	cc := &contract.AntasenaContract{}
	_, err := cc.QueryReport(stub.Ctx(), "nope-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestVerifyPayloadHash(t *testing.T) {
	stub := NewMockStub("BCAMSP", "alice@bca")
	cc := &contract.AntasenaContract{}

	payload := "the-actual-xbrl-bytes"
	rep, err := cc.SubmitReport(stub.Ctx(), "014", "LBU", "2026-04",
		sha256Hex(payload), "vault://", "v1")
	require.NoError(t, err)

	ok, err := cc.VerifyPayloadHash(stub.Ctx(), rep.ReportID, payload)
	require.NoError(t, err)
	assert.True(t, ok, "matching payload should verify")

	bad, err := cc.VerifyPayloadHash(stub.Ctx(), rep.ReportID, "tampered")
	require.NoError(t, err)
	assert.False(t, bad, "tampered payload should fail verification")
}
