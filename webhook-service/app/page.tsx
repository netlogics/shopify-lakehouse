import { PrismaClient } from "@prisma/client";

export const dynamic = "force-dynamic";

const prisma = new PrismaClient();

// ─── Helper: fetch resource counts ──────────────────────────────────────────

async function getResourceCounts() {
  const [products, variants, inventory, orders, customers] = await Promise.all([
    prisma.product.count(),
    prisma.productVariant.count(),
    prisma.inventoryLevel.count(),
    prisma.orderDetail.count(),
    prisma.customer.count(),
  ]);

  return { products, variants, inventory, orders, customers };
}

// ─── Helper: fetch webhook health stats ─────────────────────────────────────

async function getHealthStats() {
  const [totalEvents, lastEvent] = await Promise.all([
    prisma.webhookEvent.count(),
    prisma.webhookEvent.findFirst({
      orderBy: { created_at: "desc" },
      select: { created_at: true },
    }),
  ]);

  return { totalEvents, lastReceived: lastEvent?.created_at ?? null };
}

// ─── Helper: fetch recent webhook events ────────────────────────────────────

async function getRecentEvents(limit = 50) {
  return prisma.webhookEvent.findMany({
    take: limit,
    orderBy: { created_at: "desc" },
  });
}

// ─── Dashboard Page ─────────────────────────────────────────────────────────

export default async function DashboardPage() {
  const counts = await getResourceCounts();
  const health = await getHealthStats();
  const events = await getRecentEvents();

  const resourceCards = [
    { label: "Products", count: counts.products, href: "/explorer?table=Product" },
    { label: "Variants", count: counts.variants, href: "/explorer?table=ProductVariant" },
    { label: "Inventory", count: counts.inventory, href: "/explorer?table=InventoryLevel" },
    { label: "Orders", count: counts.orders, href: "/explorer?table=OrderDetail" },
    { label: "Customers", count: counts.customers, href: "/explorer?table=Customer" },
  ];

  // Determine health status
  const isHealthy = health.lastReceived !== null;
  const lastReceivedStr = health.lastReceived
    ? new Date(health.lastReceived).toLocaleString()
    : "—";

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-sm text-gray-500">
            Shopify webhook receiver — data summary and recent activity
          </p>
        </div>
        <a
          href="/explorer"
          className="inline-flex items-center gap-2 rounded-md border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm transition hover:bg-gray-50"
        >
          🔍 Data Explorer
        </a>
      </div>

      {/* Health Status */}
      <section>
        <h2 className="mb-4 text-lg font-semibold text-gray-800">Service Health</h2>
        <div className="grid grid-cols-3 gap-4">
          <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
            <div className="text-sm text-gray-500">Status</div>
            <div className="mt-1 flex items-center gap-2">
              <span
                className={`inline-block h-3 w-3 rounded-full ${
                  isHealthy ? "bg-green-500" : "bg-gray-300"
                }`}
              />
              <span className="text-lg font-semibold text-gray-900">
                {isHealthy ? "Healthy" : "Idle"}
              </span>
            </div>
          </div>
          <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
            <div className="text-sm text-gray-500">Last Received</div>
            <div className="mt-1 text-lg font-semibold text-gray-900">{lastReceivedStr}</div>
          </div>
          <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
            <div className="text-sm text-gray-500">Total Events</div>
            <div className="mt-1 text-lg font-semibold text-gray-900">{health.totalEvents}</div>
          </div>
        </div>
      </section>

      {/* Resource Counts */}
      <section>
        <h2 className="mb-4 text-lg font-semibold text-gray-800">Resource Counts</h2>
        <div className="grid grid-cols-2 gap-4 md:grid-cols-5">
          {resourceCards.map((card) => (
            <a
              key={card.label}
              href={card.href}
              className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition hover:shadow-md"
            >
              <div className="text-sm text-gray-500">{card.label}</div>
              <div className="mt-1 text-2xl font-bold text-gray-900">{card.count}</div>
            </a>
          ))}
        </div>
      </section>

      {/* Recent Webhook Events */}
      <section>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-800">Recent Webhook Events</h2>
          <a
            href="/page"
            className="text-sm text-blue-600 hover:text-blue-800"
          >
            Refresh ↻
          </a>
        </div>
        <div className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Topic
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Table
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Op
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Time
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {events.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-400">
                    No webhook events received yet. Point your Shopify app&apos;s webhook URL to
                    <code className="mx-1 rounded bg-gray-100 px-1">
                      http://localhost:3456/api/webhooks
                    </code>
                  </td>
                </tr>
              ) : (
                events.map((event) => (
                  <tr key={event.id} className="hover:bg-gray-50">
                    <td className="whitespace-nowrap px-4 py-3 text-sm font-medium text-gray-900">
                      {event.topic}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">
                      {event.table_name}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <span
                        className={`inline-flex rounded-full px-2 py-1 text-xs font-semibold ${
                          event.operation === "create"
                            ? "bg-green-100 text-green-800"
                            : event.operation === "update"
                              ? "bg-blue-100 text-blue-800"
                              : "bg-red-100 text-red-800"
                        }`}
                      >
                        {event.operation}
                      </span>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">
                      {new Date(event.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
}
