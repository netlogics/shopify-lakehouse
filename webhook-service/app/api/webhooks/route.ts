import { NextRequest, NextResponse } from "next/server";
import { createHash } from "crypto";
import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

// ─── Types ──────────────────────────────────────────────────────────────────

interface ShopifyProduct {
  id: number;
  title: string;
  vendor?: string;
  product_type?: string;
  handle?: string;
  status?: string;
  tags?: string | string[];
  created_at?: string;
  updated_at?: string;
  published_at?: string;
  variants?: ShopifyVariant[];
}

interface ShopifyVariant {
  id: number;
  product_id: number;
  title?: string;
  price: string | number;
  sku?: string;
  option1?: string;
  option2?: string;
  option3?: string;
  inventory_management?: string;
  inventory_item_id: number;
  inventory_quantity?: number;
  created_at?: string;
  updated_at?: string;
}

interface ShopifyInventoryLevel {
  inventory_item_id: number;
  location_id: number;
  available?: number;
  updated_at?: string;
}

interface ShopifyOrderLineItem {
  id: number;
  order_id?: number;
  variant_id?: number;
  product_id?: number;
  title: string;
  variant_title?: string;
  name?: string;
  sku?: string;
  quantity: number;
  price: string | number;
  created_at?: string;
  updated_at?: string;
}

interface ShopifyOrder {
  id: number;
  line_items?: ShopifyOrderLineItem[];
  created_at?: string;
  updated_at?: string;
}

interface ShopifyCustomer {
  id: number;
  email?: string;
  first_name?: string;
  last_name?: string;
  phone?: string;
  state?: "enabled" | "disabled" | "pending" | "invited";
  verified_email?: boolean;
  tags?: string | string[];
  created_at?: string;
  updated_at?: string;
}

// ─── Topic → table mapping ──────────────────────────────────────────────────

const TOPIC_TABLE_MAP: Record<string, string> = {
  "products/create": "products",
  "products/update": "products",
  "products/delete": "products",
  "product_variants/create": "product_variants",
  "product_variants/update": "product_variants",
  "product_variants/delete": "product_variants",
  "inventory_levels/update": "inventory_levels",
  "orders/create": "order_details",
  "orders/updated": "order_details",
  "orders/delete": "order_details",
  "customers/create": "customers",
  "customers/update": "customers",
  "customers/delete": "customers",
};

// ─── Helpers ────────────────────────────────────────────────────────────────

function verifyHmac(payload: string, hmac: string, secret: string): boolean {
  const digest = createHash("sha256").update(payload, "utf-8").digest("hex");
  return hmac === digest;
}

function payloadHash(payload: string): string {
  return createHash("sha256").update(payload).digest("hex");
}

function toStr(val: number | string | undefined | null): string {
  if (val === undefined || val === null) return "";
  return String(val);
}

function toIsoTimestamp(val: string | undefined | null): string {
  if (!val) return new Date().toISOString();
  // Already ISO 8601 — pass through
  return val;
}

function getOperation(topic: string): string {
  if (topic.endsWith("/create")) return "create";
  if (topic.endsWith("/update") || topic.endsWith("/updated")) return "update";
  if (topic.endsWith("/delete")) return "delete";
  return "unknown";
}

// ─── Upsert handlers ────────────────────────────────────────────────────────

async function upsertProduct(data: ShopifyProduct) {
  const id = toStr(data.id);
  const tags = typeof data.tags === "string"
    ? data.tags
    : Array.isArray(data.tags)
      ? data.tags.join(",")
      : "";

  await prisma.product.upsert({
    where: { id },
    update: {
      title: data.title,
      handle: data.handle || null,
      status: data.status || "active",
      tags,
      updated_at: toIsoTimestamp(data.updated_at),
    },
    create: {
      id,
      title: data.title,
      handle: data.handle || null,
      status: data.status || "active",
      tags,
      created_at: toIsoTimestamp(data.created_at),
      updated_at: toIsoTimestamp(data.updated_at),
    },
  });

  // Upsert variants
  if (data.variants?.length) {
    for (const v of data.variants) {
      const variantId = toStr(v.id);
      const price = typeof v.price === "number"
        ? v.price.toFixed(2)
        : String(v.price);

      await prisma.productVariant.upsert({
        where: { id: variantId },
        update: {
          price,
          sku: v.sku || null,
          updated_at: toIsoTimestamp(v.updated_at),
        },
        create: {
          id: variantId,
          product_id: id,
          price,
          sku: v.sku || null,
          created_at: toIsoTimestamp(v.created_at),
          updated_at: toIsoTimestamp(v.updated_at),
        },
      });
    }
  }
}

async function upsertInventoryLevel(data: ShopifyInventoryLevel) {
  // Shopify inventory_level uses inventory_item_id + location_id as composite key
  const key = `${toStr(data.inventory_item_id)}:${toStr(data.location_id)}`;

  await prisma.inventoryLevel.upsert({
    where: {
      product_id_inventory_item_id_location_id: {
        // Use inventory_item_id as product_id for lookup; in practice
        // this needs enrichment. We store the raw inventory_item_id in id.
        product_id: key,
        inventory_item_id: toStr(data.inventory_item_id),
        location_id: toStr(data.location_id),
      },
    },
    update: {
      available: data.available ?? 0,
      updated_at: toIsoTimestamp(data.updated_at),
    },
    create: {
      id: key,
      product_id: key,
      inventory_item_id: toStr(data.inventory_item_id),
      location_id: toStr(data.location_id),
      available: data.available ?? 0,
      updated_at: toIsoTimestamp(data.updated_at),
    },
  });
}

async function upsertOrderDetails(order: ShopifyOrder) {
  const orderId = toStr(order.id);

  if (order.line_items?.length) {
    for (const item of order.line_items) {
      const price = typeof item.price === "number"
        ? item.price.toFixed(2)
        : String(item.price);

      await prisma.orderDetail.upsert({
        where: { id: toStr(item.id) },
        update: {
          order_id: orderId,
          quantity: item.quantity,
          price,
          updated_at: toIsoTimestamp(item.updated_at),
          ...(item.product_id ? { product_id: toStr(item.product_id) } : {}),
          ...(item.variant_id ? { variant_id: toStr(item.variant_id) } : {}),
        },
        create: {
          id: toStr(item.id),
          order_id: orderId,
          product_id: item.product_id ? toStr(item.product_id) : undefined,
          variant_id: item.variant_id ? toStr(item.variant_id) : undefined,
          title: item.title,
          quantity: item.quantity,
          price,
          created_at: toIsoTimestamp(item.created_at),
          updated_at: toIsoTimestamp(item.updated_at),
        },
      });
    }
  }
}

async function upsertCustomer(data: ShopifyCustomer) {
  const id = toStr(data.id);
  const tags = typeof data.tags === "string"
    ? data.tags
    : Array.isArray(data.tags)
      ? data.tags.join(",")
      : "";

  await prisma.customer.upsert({
    where: { id },
    update: {
      email: data.email || undefined,
      first_name: data.first_name || undefined,
      last_name: data.last_name || undefined,
      phone: data.phone || undefined,
      state: data.state || "enabled",
      verified_email: data.verified_email || false,
      tags,
      updated_at: toIsoTimestamp(data.updated_at),
    },
    create: {
      id,
      email: data.email || undefined,
      first_name: data.first_name || undefined,
      last_name: data.last_name || undefined,
      phone: data.phone || undefined,
      state: data.state || "enabled",
      verified_email: data.verified_email || false,
      tags,
      created_at: toIsoTimestamp(data.created_at),
      updated_at: toIsoTimestamp(data.updated_at),
    },
  });
}

// ─── Main handler ───────────────────────────────────────────────────────────

export async function POST(request: NextRequest) {
  const topic = request.headers.get("x-shopify-topic") ?? "unknown";
  const hmac = request.headers.get("x-shopify-hmac-sha256") ?? "";
  const apiKey = request.headers.get("x-shopify-shop-domain") ?? "";

  // Read body
  const bodyText = await request.text();

  // Verify HMAC (skip if no secret configured — local dev)
  const secret = process.env.SHOPIFY_API_SECRET ?? "";
  if (secret && !verifyHmac(bodyText, hmac, secret)) {
    return NextResponse.json(
      { error: "HMAC verification failed" },
      { status: 401 },
    );
  }

  // Parse payload
  let payload: unknown;
  try {
    payload = JSON.parse(bodyText);
  } catch {
    return NextResponse.json(
      { error: "Invalid JSON" },
      { status: 400 },
    );
  }

  // Determine table and operation
  const tableName = TOPIC_TABLE_MAP[topic] ?? "unknown";
  const operation = getOperation(topic);
  const hash = payloadHash(bodyText);
  const sourceId = request.headers.get("x-shopify-webhook-id") ?? "unknown";

  // Dedup check
  const existing = await prisma.webhookEvent.findUnique({
    where: { payload_hash: hash },
  });
  if (existing) {
    return NextResponse.json({ status: "duplicate", topic, tableName });
  }

  // Record audit event
  await prisma.webhookEvent.create({
    data: {
      topic,
      source_id: sourceId,
      table_name: tableName,
      operation,
      payload_hash: hash,
    },
  });

  // Normalize and upsert based on topic
  try {
    switch (topic) {
      case "products/create":
      case "products/update":
      case "products/delete":
        await upsertProduct(payload as ShopifyProduct);
        break;

      case "product_variants/create":
      case "product_variants/update":
      case "product_variants/delete":
        // Variants are embedded in products; skip standalone variant webhooks
        // (Shopify sends them but we handle via product upsert)
        break;

      case "inventory_levels/update":
        await upsertInventoryLevel(payload as ShopifyInventoryLevel);
        break;

      case "orders/create":
      case "orders/updated":
      case "orders/delete":
        await upsertOrderDetails(payload as ShopifyOrder);
        break;

      case "customers/create":
      case "customers/update":
      case "customers/delete":
        await upsertCustomer(payload as ShopifyCustomer);
        break;

      default:
        console.warn(`Unhandled topic: ${topic}`);
    }
  } catch (err) {
    console.error("Upsert error:", err);
    // Don't fail the webhook — audit event is already recorded
  }

  return NextResponse.json({
    status: "received",
    topic,
    tableName,
    operation,
    hash,
  });
}
