/**
 * JWT bearer-token middleware.
 * The token carries `username` and `org` claims; downstream handlers use them
 * to look up the matching identity in the Fabric wallet.
 */
const jwt = require('jsonwebtoken');

const SECRET = process.env.JWT_SECRET || 'hephaestus_dev_only_change_me';

function requireAuth(req, res, next) {
  const header = req.headers.authorization || '';
  const token = header.startsWith('Bearer ') ? header.slice(7) : null;
  if (!token) return res.status(401).json({ error: 'missing bearer token' });

  try {
    req.user = jwt.verify(token, SECRET);
    next();
  } catch (e) {
    return res.status(401).json({ error: `invalid token: ${e.message}` });
  }
}

function requireOrg(...allowed) {
  return (req, res, next) => {
    if (!req.user) return res.status(401).json({ error: 'not authenticated' });
    if (!allowed.includes(req.user.org)) {
      return res.status(403).json({
        error: `forbidden: this action requires one of [${allowed.join(', ')}]`,
      });
    }
    next();
  };
}

function sign(payload, ttl = '12h') {
  return jwt.sign(payload, SECRET, { expiresIn: ttl });
}

module.exports = { requireAuth, requireOrg, sign };
