import { NextResponse } from "next/server";
import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

// Prometheus format metrics
function formatPrometheusMetrics(metrics: {
  total: number;
  byType: Record<string, number>;
  lastEventTimestamp: string | null;
}): string {
  const lines: string[] = [];

  // Total events counter
  lines.push(`# HELP webhook_events_total Total number of webhook events received`);
  lines.push(`# TYPE webhook_events_total counter`);
  lines.push(`webhook_events_total ${metrics.total}`);
  lines.push("");

  // Events by type
  lines.push(`# HELP webhook_events_by_type Webhook events grouped by type`);
  lines.push(`# TYPE webhook_events_by_type counter`);
  for (const [type, count] of Object.entries(metrics.byType)) {
    lines.push(`webhook_events_by_type{type="${type}"} ${count}`);
  }
  lines.push("");

  // Last event timestamp
  lines.push(`# HELP webhook_last_event_timestamp Timestamp of last received event (Unix seconds)`);
  lines.push(`# TYPE webhook_last_event_timestamp gauge`);
  const timestamp = metrics.lastEventTimestamp
    ? Math.floor(new Date(metrics.lastEventTimestamp).getTime() / 1000)
    : 0;
  lines.push(`webhook_last_event_timestamp ${timestamp}`);

  return lines.join("\n");
}

export async function GET() {
  try {
    // Get total count
    const total = await prisma.webhookEvent.count();

    // Get counts by type
    const eventsByTopic = await prisma.webhookEvent.groupBy({
      by: ["topic"],
      _count: true,
    });

    const byType: Record<string, number> = {};
    for (const event of eventsByTopic) {
      // Normalize topic to type (e.g., "products/create" → "product")
      const type = event.topic.split("/")[0].replace(/s$/, "");
      byType[type] = (byType[type] || 0) + event._count;
    }

    // Get last event timestamp
    const lastEvent = await prisma.webhookEvent.findFirst({
      orderBy: { created_at: "desc" },
      select: { created_at: true },
    });

    const metrics = {
      total,
      byType,
      lastEventTimestamp: lastEvent?.created_at?.toISOString() || null,
    };

    const body = formatPrometheusMetrics(metrics);

    return new NextResponse(body, {
      headers: {
        "Content-Type": "text/plain; charset=utf-8",
      },
    });
  } catch (error) {
    console.error("Metrics error:", error);
    return NextResponse.json(
      { error: "Failed to fetch metrics" },
      { status: 500 }
    );
  }
}
