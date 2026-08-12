import { NextResponse } from "next/server"
import { PrismaClient } from "@prisma/client"

const prisma = new PrismaClient()

export const dynamic = "force-dynamic"

export async function GET() {
  const tables = [
    { name: "Product", label: "Products" },
    { name: "ProductVariant", label: "Product Variants" },
    { name: "InventoryLevel", label: "Inventory Levels" },
    { name: "OrderDetail", label: "Order Details" },
    { name: "Customer", label: "Customers" },
    { name: "WebhookEvent", label: "Webhook Events" },
  ]

  const withCounts = await Promise.all(
    tables.map(async (t) => {
      try {
        const count = await (prisma as any)[t.name].count()
        return { ...t, count }
      } catch {
        return { ...t, count: 0 }
      }
    })
  )

  return NextResponse.json(withCounts)
}
