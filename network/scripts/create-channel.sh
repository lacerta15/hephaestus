#!/usr/bin/env bash
# Create the antasena-channel and join all peers.
#
# Runs inside the cli container.

set -euo pipefail

CHANNEL_NAME="antasena-channel"
ORDERER_CA="/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/ordererOrganizations/bi.hephaestus.id/orderers/orderer1.bi.hephaestus.id/msp/tlscacerts/tlsca.bi.hephaestus.id-cert.pem"

setEnv() {
  ORG=$1
  DOMAIN=$2
  MSPID=$3
  PORT=$4
  export CORE_PEER_LOCALMSPID="${MSPID}"
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/${DOMAIN}/peers/peer0.${DOMAIN}/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/${DOMAIN}/users/Admin@${DOMAIN}/msp"
  export CORE_PEER_ADDRESS="peer0.${DOMAIN}:${PORT}"
}

run_in_cli() {
  docker exec hephaestus-cli bash -c "$1"
}

# 1. Create channel using BCA admin
echo "🚀 Creating channel ${CHANNEL_NAME}..."
run_in_cli "
  export CORE_PEER_LOCALMSPID=BCAMSP
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/bca.hephaestus.id/peers/peer0.bca.hephaestus.id/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/bca.hephaestus.id/users/Admin@bca.hephaestus.id/msp
  export CORE_PEER_ADDRESS=peer0.bca.hephaestus.id:7051
  peer channel create -o orderer1.bi.hephaestus.id:7050 \
    -c ${CHANNEL_NAME} \
    -f /opt/gopath/src/github.com/hyperledger/fabric/peer/network/channel-artifacts/antasena-channel.tx \
    --tls --cafile ${ORDERER_CA} \
    --outputBlock /opt/gopath/src/github.com/hyperledger/fabric/peer/network/channel-artifacts/${CHANNEL_NAME}.block
"

# 2. Each org joins
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
  echo "🤝 ${org} (${MSPID}) joining channel..."
  run_in_cli "
    export CORE_PEER_LOCALMSPID=${MSPID}
    export CORE_PEER_TLS_ENABLED=true
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/${DOMAIN}/peers/peer0.${DOMAIN}/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/network/organizations/peerOrganizations/${DOMAIN}/users/Admin@${DOMAIN}/msp
    export CORE_PEER_ADDRESS=peer0.${DOMAIN}:7051
    peer channel join -b /opt/gopath/src/github.com/hyperledger/fabric/peer/network/channel-artifacts/${CHANNEL_NAME}.block
  "
done

echo "✅ All peers joined ${CHANNEL_NAME}"
