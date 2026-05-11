/**
 * Thin wrapper over fabric-network's Gateway.
 *
 * Caches one Gateway instance per (org, user) tuple. The wallet is mounted
 * read-only from the host; identities are pre-enrolled with Fabric CA at
 * network bootstrap time.
 */
const path = require('path');
const fs = require('fs');
const { Gateway, Wallets } = require('fabric-network');
const YAML = require('yaml');

const logger = require('./logger');

const CHANNEL = process.env.FABRIC_CHANNEL || 'antasena-channel';
const CHAINCODE = process.env.FABRIC_CHAINCODE || 'antasena';
const WALLET_PATH = process.env.FABRIC_WALLET_PATH || path.resolve(__dirname, '..', 'wallet');
const CONNECTION_DIR =
  process.env.FABRIC_CONNECTION_DIR || path.resolve(__dirname, '..', 'connection');

const gatewayCache = new Map(); // key: `${org}:${user}` -> Gateway

function ccpFor(org) {
  // expects connection-<lowercase-org>.yaml in CONNECTION_DIR
  const filename = `connection-${org.toLowerCase().replace(/msp$/, '')}.yaml`;
  const fp = path.join(CONNECTION_DIR, filename);
  if (!fs.existsSync(fp)) {
    throw new Error(`Connection profile not found for ${org}: ${fp}`);
  }
  return YAML.parse(fs.readFileSync(fp, 'utf8'));
}

async function getContract(org, userId) {
  const key = `${org}:${userId}`;
  let gateway = gatewayCache.get(key);

  if (!gateway) {
    const wallet = await Wallets.newFileSystemWallet(WALLET_PATH);
    const identity = await wallet.get(userId);
    if (!identity) {
      throw new Error(
        `Identity '${userId}' not in wallet. Enroll it with Fabric CA first.`
      );
    }
    gateway = new Gateway();
    await gateway.connect(ccpFor(org), {
      wallet,
      identity: userId,
      discovery: { enabled: true, asLocalhost: process.env.FABRIC_LOCALHOST === 'true' },
    });
    gatewayCache.set(key, gateway);
    logger.info(`Fabric gateway connected for ${key}`);
  }

  const network = await gateway.getNetwork(CHANNEL);
  return network.getContract(CHAINCODE);
}

async function disconnectAll() {
  for (const [k, gw] of gatewayCache.entries()) {
    try {
      gw.disconnect();
    } catch (e) {
      logger.warn(`Error disconnecting ${k}: ${e.message}`);
    }
  }
  gatewayCache.clear();
}

process.on('SIGTERM', disconnectAll);
process.on('SIGINT', disconnectAll);

module.exports = { getContract, disconnectAll };
