/**
 * /api/v1/reports/* — the regulatory reporting surface.
 */
const router = require('express').Router();
const { requireAuth, requireOrg } = require('../middleware/auth');
const { getContract } = require('../fabric');

const BANK_ORGS = ['BCAMSP', 'MandiriMSP', 'BRIMSP', 'BNIMSP', 'CIMBMSP'];

// ---------------------------------------------------------------------------
// POST /reports — submit
// ---------------------------------------------------------------------------
router.post('/', requireAuth, requireOrg(...BANK_ORGS), async (req, res, next) => {
  try {
    const { bankCode, reportType, reportingPeriod, payloadHash, payloadURI, schemaVersion } =
      req.body || {};
    const contract = await getContract(req.user.org, req.user.username);
    const result = await contract.submitTransaction(
      'SubmitReport',
      bankCode,
      reportType,
      reportingPeriod,
      payloadHash,
      payloadURI,
      schemaVersion
    );
    res.status(201).json(JSON.parse(result.toString()));
  } catch (e) {
    next(e);
  }
});

// ---------------------------------------------------------------------------
// POST /reports/:id/validate — BI only
// ---------------------------------------------------------------------------
router.post('/:id/validate', requireAuth, requireOrg('BIMSP'), async (req, res, next) => {
  try {
    const { markFinal = false, notes = '' } = req.body || {};
    const contract = await getContract(req.user.org, req.user.username);
    const result = await contract.submitTransaction(
      'ValidateReport',
      req.params.id,
      String(Boolean(markFinal)),
      notes
    );
    res.json(JSON.parse(result.toString()));
  } catch (e) {
    next(e);
  }
});

// ---------------------------------------------------------------------------
// GET /reports/:id — query single
// ---------------------------------------------------------------------------
router.get('/:id', requireAuth, async (req, res, next) => {
  try {
    const contract = await getContract(req.user.org, req.user.username);
    const result = await contract.evaluateTransaction('QueryReport', req.params.id);
    res.json(JSON.parse(result.toString()));
  } catch (e) {
    next(e);
  }
});

// ---------------------------------------------------------------------------
// GET /reports/:id/audit — immutable history
// ---------------------------------------------------------------------------
router.get('/:id/audit', requireAuth, async (req, res, next) => {
  try {
    const contract = await getContract(req.user.org, req.user.username);
    const result = await contract.evaluateTransaction('GetAuditTrail', req.params.id);
    res.json(JSON.parse(result.toString() || '[]'));
  } catch (e) {
    next(e);
  }
});

// ---------------------------------------------------------------------------
// GET /reports?bankCode=&period=&status= — rich query
// ---------------------------------------------------------------------------
router.get('/', requireAuth, async (req, res, next) => {
  try {
    const { bankCode, period, status } = req.query;
    const contract = await getContract(req.user.org, req.user.username);
    let fn = null;
    let arg = null;
    if (bankCode) {
      fn = 'QueryReportsByBank';
      arg = bankCode;
    } else if (period) {
      fn = 'QueryReportsByPeriod';
      arg = period;
    } else if (status) {
      fn = 'QueryReportsByStatus';
      arg = status;
    } else {
      return res
        .status(400)
        .json({ error: 'provide one of bankCode, period, or status query parameter' });
    }
    const result = await contract.evaluateTransaction(fn, arg);
    res.json(JSON.parse(result.toString() || '[]'));
  } catch (e) {
    next(e);
  }
});

// ---------------------------------------------------------------------------
// POST /reports/:id/verify-payload — checks an off-chain payload against on-chain hash
// ---------------------------------------------------------------------------
router.post('/:id/verify-payload', requireAuth, async (req, res, next) => {
  try {
    const { payload } = req.body || {};
    if (!payload) return res.status(400).json({ error: 'payload required' });
    const contract = await getContract(req.user.org, req.user.username);
    const result = await contract.evaluateTransaction(
      'VerifyPayloadHash',
      req.params.id,
      payload
    );
    res.json({ matches: result.toString() === 'true' });
  } catch (e) {
    next(e);
  }
});

module.exports = router;
