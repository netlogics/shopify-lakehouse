import { NextRequest, NextResponse } from "next/server"
import Database from "better-sqlite3"
import path from "path"

export const dynamic = "force-dynamic"

// Only allow SELECT statements
const ALLOWED_PATTERN = /^\s*SELECT\b/im

export async function POST(request: NextRequest) {
  try {
    const { sql } = await request.json()

    if (!sql || typeof sql !== "string") {
      return NextResponse.json({ error: "sql field is required" }, { status: 400 })
    }

    if (!ALLOWED_PATTERN.test(sql.trim())) {
      return NextResponse.json({ error: "Only SELECT queries are allowed" }, { status: 403 })
    }

    const dbPath = path.join(process.cwd(), "data", "webhook.db")
    const db = new Database(dbPath)

    // Enable WAL mode for better read performance
    db.pragma("journal_mode = WAL")

    // Get table names for reference
    const tables = db.prepare("SELECT name FROM sqlite_master WHERE type='table'").all() as { name: string }[]

    const rows = db.prepare(sql).all() as Record<string, unknown>[]
    const columns = rows.length > 0 ? Object.keys(rows[0]) : []

    db.close()

    return NextResponse.json({ rows, columns, rowCount: rows.length })
  } catch (err) {
    console.error("SQL query error:", err)
    const message = err instanceof Error ? err.message : "Unknown error"
    return NextResponse.json({ error: message }, { status: 500 })
  }
}

export async function GET() {
  // Return available tables for auto-completion
  const dbPath = path.join(process.cwd(), "data", "webhook.db")
  try {
    const db = new Database(dbPath)
    const tables = db.prepare(
      "SELECT name, sql FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
    ).all() as { name: string; sql: string }[]
    db.close()
    return NextResponse.json(tables)
  } catch {
    // DB may not exist yet
    return NextResponse.json([])
  }
}
