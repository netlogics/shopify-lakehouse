# Webhook Service

Shopify webhook receiver for the lakehouse pipeline. Receives webhooks, verifies HMAC signatures, normalizes payloads, and persists data to SQLite.

## Quick Start

```bash
# Install dependencies
cd webhook-service && npm install

# Set up environment
cp .env.example .env
# Edit .env with your Shopify API credentials

# Initialize database
npx prisma migrate dev --name init

# Run development server
npm run dev
```

## API

**Webhook endpoint:** `POST /api/webhooks`
- Verifies `x-shopify-hmac-sha256` header
- Parses and normalizes payloads
- Stores audit records in `WebhookEvent` table
- UPSERTs resources into SQLite

## Database

SQLite database at `data/webhook.db` (Docker) or `./webhook.db` (local).

Tables:
- `Product` — Shopify products
- `ProductVariant` — Product variants
- `InventoryLevel` — Inventory counts
- `OrderDetail` — Order line items
- `Customer` — Shopify customers (new)
- `WebhookEvent` — Audit log

## Docker

```bash
# From project root
docker-compose up webhook-service

# Build and run standalone
cd webhook-service && docker build -t webhook-service .
docker run -p 3456:3456 --env-file .env webhook-service
```

## Architecture

```
Shopify → Webhook Service (HMAC verify → normalize → SQLite)
                        → Dashboard (resource counts + event log)
                        → Data Explorer (query SQLite)
```
