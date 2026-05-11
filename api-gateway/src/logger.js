/**
 * Winston logger with sane defaults for the API gateway.
 * In production, pipe stdout to your central log aggregation (Loki, Splunk, ELK).
 */
const winston = require('winston');

const logger = winston.createLogger({
  level: process.env.LOG_LEVEL || 'info',
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.errors({ stack: true }),
    winston.format.splat(),
    process.env.NODE_ENV === 'production'
      ? winston.format.json()
      : winston.format.combine(
          winston.format.colorize(),
          winston.format.printf(
            ({ level, message, timestamp }) => `${timestamp} ${level}: ${message}`
          )
        )
  ),
  transports: [new winston.transports.Console()],
});

module.exports = logger;
