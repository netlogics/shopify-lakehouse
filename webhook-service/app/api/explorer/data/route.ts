import { NextRequest, NextResponse } from "next/server"
import { PrismaClient } from "@prisma/client"

const prisma = new PrismaClient()

export const dynamic = "force-dynamic"

// Mapping of Prisma model names to their available filterable fields
const MODEL_FIELDS: Record<string, string[]> = {
  Product: ["id", "title", "vendor", "product_type", "status", "created_at", "updated_at"],
  ProductVariant: ["id", "title", "price", "sku", "created_at", "updated_at"],
  InventoryLevel: ["inventory_item_id", "location_id", "available", "updated_at"],
  OrderDetail: ["order_id", "id", "title", "sku", "price", "quantity", "created_at", "updated_at"],
  Customer: ["id", "email", "first_name", "last_name", "state", "verified_email", "created_at", "updated_at"],
  WebhookEvent: ["id", "topic", "table", "operation", "created_at"],
}

// Date fields that support range filtering
const DATE_FIELDS = ["created_at", "updated_at"]

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams
  const table = searchParams.get("table")
  const page = parseInt(searchParams.get("page") || "1", 10)
  const limit = parseInt(searchParams.get("limit") || "50", 10)
  const sortBy = searchParams.get("sortBy") || undefined
  const sortOrder = searchParams.get("sortOrder") || "desc"
  const filtersRaw = searchParams.get("filters")

  if (!table) {
    return NextResponse.json({ error: "table parameter is required" }, { status: 400 })
  }

  const allowedFields = MODEL_FIELDS[table]
  if (!allowedFields) {
    return NextResponse.json({ error: `Unknown table: ${table}` }, { status: 400 })
  }

  // Parse filters
  const filters: Record<string, any> = {}
  if (filtersRaw) {
    try {
      const parsed = JSON.parse(filtersRaw)
      for (const [field, value] of Object.entries(parsed)) {
        if (!allowedFields.includes(field)) continue

        if (typeof value === "string" && value.trim()) {
          // Text search
          filters[field] = { contains: value, mode: "insensitive" as const }
        } else if (typeof value === "object" && value !== null && "gte" in value && "lte" in value) {
          // Date range
          const gteVal = String((value as any).gte)
          const lteVal = String((value as any).lte)
          const gte = new Date(gteVal)
          const lte = new Date(lteVal)
          if (DATE_FIELDS.includes(field) && !isNaN(gte.getTime()) && !isNaN(lte.getTime())) {
            filters[field] = { gte, lte }
          }
        }
      }
    } catch {
      return NextResponse.json({ error: "Invalid filters JSON" }, { status: 400 })
    }
  }

  // Sort by
  const orderBy: Record<string, "asc" | "desc"> = {}
  if (sortBy && allowedFields.includes(sortBy)) {
    orderBy[sortBy] = (sortOrder as "asc" | "desc") || "desc"
  }

  const skip = (page - 1) * limit
  const prismaTable = (prisma as any)[table]

  try {
    const [rows, total] = await Promise.all([
      prismaTable.findMany({
        skip,
        take: limit,
        where: Object.keys(filters).length > 0 ? filters : undefined,
        orderBy,
      }),
      prismaTable.count({ where: Object.keys(filters).length > 0 ? filters : undefined }),
    ])

    return NextResponse.json({
      rows,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    })
  } catch (err) {
    console.error(`Error querying ${table}:`, err)
    return NextResponse.json({ error: `Failed to query ${table}` }, { status: 500 })
  }
}
