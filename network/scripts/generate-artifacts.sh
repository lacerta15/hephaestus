#!/usr/bin/env bash
# Generate crypto material + genesis block + channel tx for Hephaestus.
#
# Requires: cryptogen, configtxgen (Hyperledger Fabric binaries 2.5+)
# These are typically installed at ./bin/ via the official Fabric install.sh.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NET_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${NET_DIR}"

# 1. Crypto material -----------------------------------------------------------
echo "🔐 Generating crypto material with cryptogen..."
rm -rf organizations
mkdir -p organizations
cryptogen generate --config=configtx/crypto-config.yaml --output=organizations

# Reorganise into the canonical layout the compose file expects
# cryptogen by default creates organizations/{ordererOrganizations,peerOrganizations}
# which is exactly what configtx and docker-compose expect.

# 2. Genesis block + channel tx -----------------------------------------------
echo "📦 Generating system genesis block..."
mkdir -p system-genesis-block channel-artifacts
configtxgen -profile HephaestusGenesis \
            -channelID system-channel \
            -outputBlock system-genesis-block/genesis.block \
            -configPath configtx

echo "📦 Generating antasena-channel transaction..."
configtxgen -profile AntasenaChannel \
            -outputCreateChannelTx channel-artifacts/antasena-channel.tx \
            -channelID antasena-channel \
            -configPath configtx

# Anchor peer txs (one per org)
for org in BCA Mandiri BRI BNI CIMB OJK; do
  echo "📦 Anchor peer tx for ${org}..."
  configtxgen -profile AntasenaChannel \
              -outputAnchorPeersUpdate channel-artifacts/${org}MSPanchors.tx \
              -channelID antasena-channel \
              -asOrg ${org}MSP \
              -configPath configtx
done

echo "✅ All artifacts generated under network/"
