/**
 * Demo auth: trades a (username, org) pair for a JWT.
 *
 * In production this endpoint should:
 *   - integrate with corporate SSO (OIDC / SAML)
 *   - verify that the user is enrolled with the org's Fabric CA
 *   - never accept arbitrary identity claims from the client
 *
 * For PoC purposes we accept any username and trust that the wallet has
 * a matching identity pre-enrolled.
 */
const router = require('express').Router();
const { sign } = require('../middleware/auth');

const ALLOWED_ORGS = new Set([
  'BIMSP',
  'OJKMSP',
  'BCAMSP',
  'MandiriMSP',
  'BRIMSP',
  'BNIMSP',
  'CIMBMSP',
]);

router.post('/login', (req, res) => {
  const { username, org } = req.body || {};
  if (!username || !org) {
    return res.status(400).json({ error: 'username and org are required' });
  }
  if (!ALLOWED_ORGS.has(org)) {
    return res.status(400).json({ error: `unknown org ${org}` });
  }
  const token = sign({ username, org });
  res.json({ token, username, org, expiresIn: '12h' });
});

module.exports = router;
