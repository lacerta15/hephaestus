#!/usr/bin/env bash
# End-to-end demo: BCA submits an LBU report, BI validates it, OJK queries audit trail.

set -euo pipefail

API="http://localhost:3000/api/v1"

echo "🏦 1. BCA logs in..."
TOKEN=$(curl -s -X POST ${API}/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice@bca","org":"BCAMSP"}' | jq -r .token)

echo "📤 2. BCA submits LBU report for April 2026..."
SUBMIT=$(curl -s -X POST ${API}/reports \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "bankCode": "014",
    "reportType": "LBU",
    "reportingPeriod": "2026-04",
    "payloadHash": "a3f5bd6e1c2d4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b",
    "payloadURI": "vault://bca/2026-04/lbu.xbrl",
    "schemaVersion": "BI-LBU-2.1"
  }')

REPORT_ID=$(echo "${SUBMIT}" | jq -r .reportId)
echo "   ✓ Report ID: ${REPORT_ID}"

echo "🔐 3. BI logs in..."
BI_TOKEN=$(curl -s -X POST ${API}/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"validator@bi.go.id","org":"BIMSP"}' | jq -r .token)

echo "✅ 4. BI validates and finalises the report..."
curl -s -X POST ${API}/reports/${REPORT_ID}/validate \
  -H "Authorization: Bearer ${BI_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"markFinal": true, "notes": "All checks passed."}' | jq

echo "📜 5. OJK fetches the immutable audit trail..."
OJK_TOKEN=$(curl -s -X POST ${API}/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"audit@ojk.go.id","org":"OJKMSP"}' | jq -r .token)

curl -s ${API}/reports/${REPORT_ID}/audit \
  -H "Authorization: Bearer ${OJK_TOKEN}" | jq

echo ""
echo "🎉 Demo complete. Open http://localhost:8080 to explore the dashboard."
