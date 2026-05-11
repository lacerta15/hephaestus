# Hephaestus API Gateway

Node.js + Express service that exposes a small REST API over the Antasena chaincode. It is the integration point for bank back-office systems and the reference web dashboard.

## Run locally (assumes `make up` is already done)

```bash
cd api-gateway
cp .env.example .env
npm install
npm start
```

OpenAPI docs: <http://localhost:3000/api/docs>

## Architecture

```
┌──────────────┐   HTTPS   ┌──────────────┐   gRPC + TLS   ┌──────────────┐
│ Bank back-   │──────────►│  API gateway │───────────────►│ Fabric peer  │
│ office (CSV) │  + JWT    │  (this svc)  │   fabric-      │  (BCA/Mand/  │
└──────────────┘           └──────────────┘   network SDK   │   BRI/...)   │
                                  │                          └──────────────┘
                                  │
                                  └──► wallet (X.509 + private key per user)
```

## Identity model

Each request carries a JWT. The token's `org` claim selects which organisation's connection profile is loaded; the `username` claim selects which wallet identity is used to sign Fabric proposals.

In production the `/auth/login` endpoint must be replaced with proper SSO (OIDC) or mutual-TLS at the bank's edge.
