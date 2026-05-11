/**
 * Hephaestus REST API gateway.
 *
 * Sits between bank back-office systems (or the reference dashboard) and the
 * Hyperledger Fabric network. Exposes a small, opinionated HTTP API so that
 * application teams don't need to learn the Fabric SDK to integrate.
 *
 * Endpoints (see src/routes/reports.js for full surface):
 *   POST   /api/v1/auth/login                  - issue a JWT for a known org user
 *   POST   /api/v1/reports                     - submit a new report (bank only)
 *   POST   /api/v1/reports/:id/validate        - validate a report (BI only)
 *   GET    /api/v1/reports/:id                 - read current state
 *   GET    /api/v1/reports/:id/audit           - immutable history
 *   GET    /api/v1/reports?bankCode=...        - rich query
 *   GET    /api/v1/health                      - liveness
 *
 * Auth:
 *   - JWT bearer token. The token's `org` claim drives which Fabric identity
 *     is used to invoke chaincode. The wallet must contain a matching identity.
 */

const express = require('express');
const helmet = require('helmet');
const cors = require('cors');
const morgan = require('morgan');
const swaggerUi = require('swagger-ui-express');
const YAML = require('yaml');
const fs = require('fs');
const path = require('path');

require('dotenv').config();

const logger = require('./logger');
const auth = require('./routes/auth');
const reports = require('./routes/reports');
const health = require('./routes/health');
const { errorHandler } = require('./middleware/errors');

const app = express();
const PORT = process.env.PORT || 3000;

// --- middleware --------------------------------------------------------------
app.use(helmet());
app.use(cors());
app.use(express.json({ limit: '10mb' }));
app.use(
  morgan('combined', {
    stream: { write: (msg) => logger.info(msg.trim()) },
  })
);

// --- routes ------------------------------------------------------------------
app.use('/api/v1/health', health);
app.use('/api/v1/auth', auth);
app.use('/api/v1/reports', reports);

// --- swagger docs ------------------------------------------------------------
const openapiPath = path.join(__dirname, '..', 'openapi.yaml');
if (fs.existsSync(openapiPath)) {
  const spec = YAML.parse(fs.readFileSync(openapiPath, 'utf8'));
  app.use('/api/docs', swaggerUi.serve, swaggerUi.setup(spec));
}

// --- error handler -----------------------------------------------------------
app.use(errorHandler);

app.listen(PORT, () => {
  logger.info(`🔨 Hephaestus API gateway listening on :${PORT}`);
  logger.info(`📚 OpenAPI docs: http://localhost:${PORT}/api/docs`);
});
