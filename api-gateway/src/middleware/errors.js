const logger = require('../logger');

/**
 * Centralised error handler. Maps known Fabric errors to sensible HTTP codes.
 */
function errorHandler(err, req, res, _next) {
  logger.error(err.stack || err.message || String(err));

  const msg = err.message || String(err);
  let code = err.status || 500;

  if (/not found/i.test(msg)) code = 404;
  else if (/forbidden|only Bank Indonesia|regulators cannot/i.test(msg)) code = 403;
  else if (/invalid|missing|must be|expected/i.test(msg)) code = 400;
  else if (/already FINAL|already exists/i.test(msg)) code = 409;

  res.status(code).json({ error: msg });
}

module.exports = { errorHandler };
