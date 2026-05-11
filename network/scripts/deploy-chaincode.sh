#!/usr/bin/env bash
# Package, install, approve and commit the antasena chaincode on antasena-channel.

set -euo pipefail

CC_NAME="antasena"
CC_VERSION="1.0"
CC_SEQUENCE="1"
CHANNEL_NAME="antasena-channel"
CC_LABEL="${CC_NAME}_${CC_VERSION}"
CC_PACKAGE="/tmp/${CC_LABEL}.tar.gz"
ORDERER_CA="/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/ordererOrganizations/bi.hephaestus.id/orderers/orderer1.bi.hephaestus.id/msp/tlscacerts/tlsca.bi.hephaestus.id-cert.pem"

run_in_cli() {
  docker exec hephaestus-cli bash -c "$1"
}

# 1. Package -------------------------------------------------------------------
echo "📦 Packaging chaincode..."
run_in_cli "
  cd /opt/gopath/src/github.com/hyperledger/fabric/peer/chaincode/antasena
  go mod vendor 2>/dev/null || true
  peer lifecycle chaincode package ${CC_PACKAGE} \
    --path /opt/gopath/src/github.com/hyperledger/fabric/peer/chaincode/antasena \
    --lang golang --label ${CC_LABEL}
"

# 2. Install on each org's peer ------------------------------------------------
declare -A ORGS=(
  [BCA]=bca.hephaestus.id:BCAMSP
  [Mandiri]=mandiri.hephaestus.id:MandiriMSP
  [BRI]=bri.hephaestus.id:BRIMSP
  [BNI]=bni.hephaestus.id:BNIMSP
  [CIMB]=cimb.hephaestus.id:CIMBMSP
  [OJK]=ojk.hephaestus.id:OJKMSP
)

for org in "${!ORGS[@]}"; do
  IFS=':' read -r DOMAIN MSPID <<< "${ORGS[$org]}"
  echo "📥 Installing on ${org}..."
  run_in_cli "
    export CORE_PEER_LOCALMSPID=${MSPID}
    export CORE_PEER_TLS_ENABLED=true
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/${DOMAIN}/peers/peer0.${DOMAIN}/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/${DOMAIN}/users/Admin@${DOMAIN}/msp
    export CORE_PEER_ADDRESS=peer0.${DOMAIN}:7051
    peer lifecycle chaincode install ${CC_PACKAGE}
  "
done

# 3. Get package ID
PKG_ID=$(run_in_cli "peer lifecycle chaincode queryinstalled" | grep "${CC_LABEL}" | head -1 | awk '{print $3}' | sed 's/,$//')
echo "📦 Package ID: ${PKG_ID}"

# 4. Each org approves
for org in "${!ORGS[@]}"; do
  IFS=':' read -r DOMAIN MSPID <<< "${ORGS[$org]}"
  echo "✍️  ${org} approving..."
  run_in_cli "
    export CORE_PEER_LOCALMSPID=${MSPID}
    export CORE_PEER_TLS_ENABLED=true
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/${DOMAIN}/peers/peer0.${DOMAIN}/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/${DOMAIN}/users/Admin@${DOMAIN}/msp
    export CORE_PEER_ADDRESS=peer0.${DOMAIN}:7051
    peer lifecycle chaincode approveformyorg \
      -o orderer1.bi.hephaestus.id:7050 \
      --channelID ${CHANNEL_NAME} \
      --name ${CC_NAME} \
      --version ${CC_VERSION} \
      --package-id ${PKG_ID} \
      --sequence ${CC_SEQUENCE} \
      --tls --cafile ${ORDERER_CA}
  "
done

# 5. Commit (BCA acts as committer with peer flags from all orgs)
echo "🚀 Committing chaincode definition..."
run_in_cli "
  export CORE_PEER_LOCALMSPID=BCAMSP
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/bca.hephaestus.id/peers/peer0.bca.hephaestus.id/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/bca.hephaestus.id/users/Admin@bca.hephaestus.id/msp
  export CORE_PEER_ADDRESS=peer0.bca.hephaestus.id:7051
  peer lifecycle chaincode commit \
    -o orderer1.bi.hephaestus.id:7050 \
    --channelID ${CHANNEL_NAME} \
    --name ${CC_NAME} \
    --version ${CC_VERSION} \
    --sequence ${CC_SEQUENCE} \
    --tls --cafile ${ORDERER_CA} \
    --peerAddresses peer0.bca.hephaestus.id:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/bca.hephaestus.id/peers/peer0.bca.hephaestus.id/tls/ca.crt \
    --peerAddresses peer0.mandiri.hephaestus.id:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/mandiri.hephaestus.id/peers/peer0.mandiri.hephaestus.id/tls/ca.crt \
    --peerAddresses peer0.bri.hephaestus.id:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/bri.hephaestus.id/peers/peer0.bri.hephaestus.id/tls/ca.crt
"

# 6. Init ledger (optional)
echo "🌱 Calling InitLedger..."
run_in_cli "
  export CORE_PEER_LOCALMSPID=BCAMSP
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/bca.hephaestus.id/peers/peer0.bca.hephaestus.id/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/bca.hephaestus.id/users/Admin@bca.hephaestus.id/msp
  export CORE_PEER_ADDRESS=peer0.bca.hephaestus.id:7051
  peer chaincode invoke -o orderer1.bi.hephaestus.id:7050 \
    --tls --cafile ${ORDERER_CA} \
    -C ${CHANNEL_NAME} -n ${CC_NAME} \
    -c '{\"function\":\"InitLedger\",\"Args\":[]}' \
    --peerAddresses peer0.bca.hephaestus.id:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/bca.hephaestus.id/peers/peer0.bca.hephaestus.id/tls/ca.crt
"

echo "✅ Chaincode ${CC_LABEL} deployed and initialised on ${CHANNEL_NAME}"
