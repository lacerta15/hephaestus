package test

// Lightweight mock for fabric-contract-api transaction context.
// Just enough surface area to exercise our contract logic without
// pulling in the full Hyperledger test fixtures.
//
// NOTE: This mock is intentionally simple. For deeper integration
// testing (endorsement policies, MSP signature validation, ordering)
// use the real network spun up by `make up`.

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// MockStub is a tiny in-memory ledger.
type MockStub struct {
	state    map[string][]byte
	history  map[string][]historyEntry
	mspID    string
	cn       string
	txID     int
}

type historyEntry struct {
	txID     string
	ts       time.Time
	value    []byte
	isDelete bool
}

func NewMockStub(mspID, cn string) *MockStub {
	return &MockStub{
		state:   map[string][]byte{},
		history: map[string][]historyEntry{},
		mspID:   mspID,
		cn:      cn,
	}
}

func (m *MockStub) WithIdentity(mspID, cn string) *MockStub {
	clone := *m
	clone.mspID = mspID
	clone.cn = cn
	return &clone
}

func (m *MockStub) Ctx() contractapi.TransactionContextInterface {
	return &mockCtx{stub: m}
}

// ---------------- ctx ----------------

type mockCtx struct {
	contractapi.TransactionContext
	stub *MockStub
}

func (c *mockCtx) GetStub() shim.ChaincodeStubInterface     { return c.stub }
func (c *mockCtx) GetClientIdentity() cid                   { return cid{m: c.stub} }
func (c *mockCtx) SetStub(_ shim.ChaincodeStubInterface)    {}
func (c *mockCtx) SetClientIdentity(_ cid)                  {}

// We need to satisfy the interface contractapi exposes. The full interface is
// large; we embed contractapi.TransactionContext to inherit defaults.
//
// However the methods we actually use (GetStub, GetClientIdentity) are
// overridden above. To keep this file tidy we shim only those.

// Override GetClientIdentity properly. The real interface returns
// contractapi.ClientIdentityInterface; we'll satisfy it via cid below.
//
// (the contractapi default returns nil so we override.)

// ---------------- client identity ----------------

type cid struct{ m *MockStub }

func (c cid) GetID() (string, error)        { return "x509::CN=" + c.m.cn + "::CN=ca", nil }
func (c cid) GetMSPID() (string, error)     { return c.m.mspID, nil }
func (c cid) GetAttributeValue(_ string) (string, bool, error) { return "", false, nil }
func (c cid) AssertAttributeValue(_, _ string) error            { return nil }
func (c cid) GetX509Certificate() (*x509.Certificate, error) {
	return &x509.Certificate{Subject: pkix.Name{CommonName: c.m.cn}}, nil
}

// ---------------- stub ----------------

func (m *MockStub) nextTxID() string {
	m.txID++
	return fmt.Sprintf("tx-%06d", m.txID)
}

func (m *MockStub) GetTxID() string                                { return m.nextTxID() }
func (m *MockStub) GetChannelID() string                           { return "antasena-channel" }
func (m *MockStub) GetState(key string) ([]byte, error)            { return m.state[key], nil }
func (m *MockStub) PutState(key string, val []byte) error {
	m.state[key] = val
	m.history[key] = append(m.history[key], historyEntry{
		txID: m.nextTxID(), ts: time.Now().UTC(), value: append([]byte{}, val...),
	})
	return nil
}
func (m *MockStub) DelState(key string) error {
	delete(m.state, key)
	m.history[key] = append(m.history[key], historyEntry{
		txID: m.nextTxID(), ts: time.Now().UTC(), isDelete: true,
	})
	return nil
}
func (m *MockStub) SetEvent(_ string, _ []byte) error              { return nil }
func (m *MockStub) GetTxTimestamp() (*timestamp.Timestamp, error) {
	now := time.Now().UTC()
	return &timestamp.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())}, nil
}

// GetStateByRange supports the prefix scan used by nextSequence.
func (m *MockStub) GetStateByRange(start, end string) (shim.StateQueryIteratorInterface, error) {
	keys := make([]string, 0, len(m.state))
	for k := range m.state {
		if k >= start && k < end {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return &mockIter{stub: m, keys: keys}, nil
}

// GetQueryResult shims CouchDB rich queries by interpreting a simplified subset.
func (m *MockStub) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	// extremely simple parser — accepts {"selector":{"docType":"report","field":"value"}}
	keys := make([]string, 0)
	for k, v := range m.state {
		var doc map[string]any
		if err := json.Unmarshal(v, &doc); err != nil {
			continue
		}
		match := true
		for _, kv := range parseSelector(query) {
			if fmt.Sprintf("%v", doc[kv[0]]) != kv[1] {
				match = false
				break
			}
		}
		if match {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return &mockIter{stub: m, keys: keys}, nil
}

func (m *MockStub) GetHistoryForKey(key string) (shim.HistoryQueryIteratorInterface, error) {
	return &mockHistIter{entries: m.history[key]}, nil
}

// --- the rest of shim.ChaincodeStubInterface is stubbed with panics; we don't use them ---

func (m *MockStub) GetArgs() [][]byte                                                          { panic("not impl") }
func (m *MockStub) GetStringArgs() []string                                                    { panic("not impl") }
func (m *MockStub) GetFunctionAndParameters() (string, []string)                               { return "", nil }
func (m *MockStub) GetArgsSlice() ([]byte, error)                                              { return nil, nil }
func (m *MockStub) GetCreator() ([]byte, error)                                                { return nil, nil }
func (m *MockStub) GetTransient() (map[string][]byte, error)                                   { return nil, nil }
func (m *MockStub) GetBinding() ([]byte, error)                                                { return nil, nil }
func (m *MockStub) GetDecorations() map[string][]byte                                          { return nil }
func (m *MockStub) GetSignedProposal() (*peer.SignedProposal, error)                           { return nil, nil }
func (m *MockStub) GetTxTimestampUTC() (time.Time, error)                                      { return time.Now().UTC(), nil }
func (m *MockStub) SetStateValidationParameter(_ string, _ []byte) error                       { return nil }
func (m *MockStub) GetStateValidationParameter(_ string) ([]byte, error)                       { return nil, nil }
func (m *MockStub) GetStateByRangeWithPagination(_, _ string, _ int32, _ string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	return nil, nil, nil
}
func (m *MockStub) GetStateByPartialCompositeKey(_ string, _ []string) (shim.StateQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) GetStateByPartialCompositeKeyWithPagination(_ string, _ []string, _ int32, _ string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	return nil, nil, nil
}
func (m *MockStub) CreateCompositeKey(_ string, _ []string) (string, error) { return "", nil }
func (m *MockStub) SplitCompositeKey(_ string) (string, []string, error)    { return "", nil, nil }
func (m *MockStub) GetQueryResultWithPagination(_ string, _ int32, _ string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	return nil, nil, nil
}
func (m *MockStub) GetPrivateData(_, _ string) ([]byte, error)             { return nil, nil }
func (m *MockStub) GetPrivateDataHash(_, _ string) ([]byte, error)         { return nil, nil }
func (m *MockStub) PutPrivateData(_, _ string, _ []byte) error             { return nil }
func (m *MockStub) DelPrivateData(_, _ string) error                       { return nil }
func (m *MockStub) PurgePrivateData(_, _ string) error                     { return nil }
func (m *MockStub) SetPrivateDataValidationParameter(_, _ string, _ []byte) error { return nil }
func (m *MockStub) GetPrivateDataValidationParameter(_, _ string) ([]byte, error) { return nil, nil }
func (m *MockStub) GetPrivateDataByRange(_, _, _ string) (shim.StateQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) GetPrivateDataByPartialCompositeKey(_, _ string, _ []string) (shim.StateQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) GetPrivateDataQueryResult(_, _ string) (shim.StateQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) InvokeChaincode(_ string, _ [][]byte, _ string) peer.Response {
	return peer.Response{}
}

// ---------------- iterators ----------------

type mockIter struct {
	stub *MockStub
	keys []string
	i    int
}

func (it *mockIter) HasNext() bool { return it.i < len(it.keys) }
func (it *mockIter) Close() error  { return nil }
func (it *mockIter) Next() (*queryresult.KV, error) {
	k := it.keys[it.i]
	it.i++
	return &queryresult.KV{Key: k, Value: it.stub.state[k]}, nil
}

type mockHistIter struct {
	entries []historyEntry
	i       int
}

func (it *mockHistIter) HasNext() bool { return it.i < len(it.entries) }
func (it *mockHistIter) Close() error  { return nil }
func (it *mockHistIter) Next() (*queryresult.KeyModification, error) {
	e := it.entries[it.i]
	it.i++
	return &queryresult.KeyModification{
		TxId:      e.txID,
		Value:     e.value,
		Timestamp: &timestamp.Timestamp{Seconds: e.ts.Unix(), Nanos: int32(e.ts.Nanosecond())},
		IsDelete:  e.isDelete,
	}, nil
}

// ---------------- helpers ----------------

// parseSelector pulls "field":"value" pairs out of a Mango-ish JSON string.
// Crude but adequate for our query shapes.
func parseSelector(q string) [][2]string {
	out := [][2]string{}
	q = strings.ReplaceAll(q, " ", "")
	// look for "selector":{...}
	idx := strings.Index(q, `"selector":{`)
	if idx < 0 {
		return out
	}
	body := q[idx+len(`"selector":{`):]
	end := strings.Index(body, "}")
	if end < 0 {
		return out
	}
	body = body[:end]
	for _, pair := range strings.Split(body, ",") {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.Trim(kv[0], `"`)
		v := strings.Trim(kv[1], `"`)
		out = append(out, [2]string{k, v})
	}
	return out
}
