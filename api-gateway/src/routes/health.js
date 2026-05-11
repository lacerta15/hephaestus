const router = require('express').Router();

router.get('/', (req, res) => {
  res.json({
    status: 'ok',
    service: 'hephaestus-api-gateway',
    version: '1.0.0',
    timestamp: new Date().toISOString(),
  });
});

module.exports = router;
