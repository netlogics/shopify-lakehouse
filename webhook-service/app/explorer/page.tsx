"use client"

import { useState, useEffect, useCallback } from "react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  ChevronLeft,
  ChevronRight,
  Loader2,
  Search,
  Table as TableIcon,
  Code2,
  RefreshCw,
  Download,
  X,
} from "lucide-react"

interface TableInfo {
  name: string
  label: string
  count: number
}

interface DataRow {
  [key: string]: unknown
}

interface QueryResult {
  rows: DataRow[]
  columns: string[]
  rowCount: number
}

export default function ExplorerPage() {
  const [tables, setTables] = useState<TableInfo[]>([])
  const [selectedTable, setSelectedTable] = useState<string>("")
  const [rows, setRows] = useState<DataRow[]>([])
  const [loading, setLoading] = useState(false)
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 50,
    total: 0,
    totalPages: 0,
  })
  const [sortBy, setSortBy] = useState<string>("")
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc")
  const [filters, setFilters] = useState<Record<string, string>>({})
  const [dateFilters, setDateFilters] = useState<Record<string, { gte: string; lte: string }>>({})
  const [sql, setSql] = useState("")
  const [queryResult, setQueryResult] = useState<QueryResult | null>(null)
  const [queryLoading, setQueryLoading] = useState(false)
  const [columns, setColumns] = useState<string[]>([])

  // Load table list
  useEffect(() => {
    fetch("/api/explorer/tables")
      .then((r) => r.json())
      .then(setTables)
  }, [])

  const loadTableData = useCallback(async () => {
    if (!selectedTable) return
    setLoading(true)

    const params = new URLSearchParams({
      table: selectedTable,
      page: String(pagination.page),
      limit: String(pagination.limit),
      ...(sortBy && { sortBy }),
      ...(sortOrder && { sortOrder }),
      ...(Object.keys(filters).length > 0 && {
        filters: JSON.stringify(filters),
      }),
    })

    // Add date range filters
    Object.entries(dateFilters).forEach(([field, range]) => {
      params.append("filters", JSON.stringify({ [field]: range }))
    })

    try {
      const res = await fetch(`/api/explorer/data?${params}`)
      const data = await res.json()
      setRows(data.rows)
      setPagination(data.pagination)
      // Extract columns from first row
      if (data.rows.length > 0) {
        setColumns(Object.keys(data.rows[0]))
      }
    } finally {
      setLoading(false)
    }
  }, [selectedTable, pagination.page, pagination.limit, sortBy, sortOrder, filters, dateFilters])

  useEffect(() => {
    loadTableData()
  }, [loadTableData])

  const handleSort = (col: string) => {
    if (sortBy === col) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc")
    } else {
      setSortBy(col)
      setSortOrder("desc")
    }
  }

  const handleFilter = (col: string, value: string) => {
    setFilters((prev) => ({ ...prev, [col]: value }))
    setPagination((prev) => ({ ...prev, page: 1 }))
  }

  const handleDateFilter = (col: string, range: { gte: string; lte: string }) => {
    setDateFilters((prev) => ({ ...prev, [col]: range }))
    setPagination((prev) => ({ ...prev, page: 1 }))
  }

  const clearFilters = () => {
    setFilters({})
    setDateFilters({})
    setPagination((prev) => ({ ...prev, page: 1 }))
  }

  const executeQuery = async () => {
    if (!sql.trim()) return
    setQueryLoading(true)
    try {
      const res = await fetch("/api/explorer/query", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ sql }),
      })
      const data = await res.json()
      if (data.error) {
        alert(`Query error: ${data.error}`)
      } else {
        setQueryResult(data)
        setColumns(data.columns)
      }
    } finally {
      setQueryLoading(false)
    }
  }

  const exportData = (data: DataRow[], cols: string[], format: "json" | "csv") => {
    let content: string
    let filename: string
    let mime: string

    if (format === "json") {
      content = JSON.stringify(data, null, 2)
      filename = `${selectedTable || "query"}_export.json`
      mime = "application/json"
    } else {
      const header = cols.join(",")
      const body = data
        .map((row) =>
          cols
            .map((c) => {
              const val = row[c]
              if (val === null || val === undefined) return ""
              const str = String(val)
              return str.includes(",") || str.includes('"') ? `"${str.replace(/"/g, '""')}"` : str
            })
            .join(",")
        )
        .join("\n")
      content = `${header}\n${body}`
      filename = `${selectedTable || "query"}_export.csv`
      mime = "text/csv"
    }

    const blob = new Blob([content], { type: mime })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }

  const truncate = (val: unknown, max = 60): string => {
    if (val === null || val === undefined) return ""
    const str = typeof val === "object" ? JSON.stringify(val) : String(val)
    return str.length > max ? str.slice(0, max) + "…" : str
  }

  const formatCellValue = (val: unknown): string => {
    if (val === null || val === undefined) return "—"
    if (val instanceof Date) return val.toISOString()
    if (typeof val === "boolean") return val ? "true" : "false"
    if (typeof val === "object") return JSON.stringify(val)
    return String(val)
  }

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Data Explorer</h1>
            <p className="text-muted-foreground text-sm">
              Browse, filter, and query your webhook data
            </p>
          </div>
          <button
            onClick={() => window.location.reload()}
            className="inline-flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm hover:bg-muted"
          >
            <RefreshCw className="h-4 w-4" />
            Refresh
          </button>
        </div>

        {/* Table selector */}
        <div className="flex items-center gap-4">
          <Select value={selectedTable} onValueChange={(v) => { setSelectedTable(v); setPagination((p) => ({ ...p, page: 1 })) }}>
            <SelectTrigger className="w-[280px]">
              <SelectValue placeholder="Select a table" />
            </SelectTrigger>
            <SelectContent>
              {tables.map((t) => (
                <SelectItem key={t.name} value={t.name}>
                  {t.label} ({t.count})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {selectedTable && (
            <span className="text-sm text-muted-foreground">
              {tables.find((t) => t.name === selectedTable)?.count} rows total
            </span>
          )}

          {Object.keys(filters).length > 0 && (
            <button
              onClick={clearFilters}
              className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
            >
              <X className="h-3 w-3" /> Clear filters
            </button>
          )}
        </div>

        <Tabs defaultValue="browse">
          <TabsList>
            <TabsTrigger value="browse">
              <TableIcon className="mr-2 h-4 w-4" />
              Browse
            </TabsTrigger>
            <TabsTrigger value="sql">
              <Code2 className="mr-2 h-4 w-4" />
              SQL Editor
            </TabsTrigger>
          </TabsList>

          {/* Browse tab */}
          <TabsContent value="browse" className="space-y-4">
            {selectedTable && (
              <>
                {/* Filters row */}
                <div className="flex flex-wrap gap-2">
                  {columns.map((col) => {
                    const isDateCol = ["created_at", "updated_at"].includes(col)
                    const hasFilter = filters[col] || dateFilters[col]

                    return (
                      <div key={col} className="flex items-center gap-1">
                        <label className="text-xs text-muted-foreground whitespace-nowrap">{col}</label>
                        {isDateCol ? (
                          <DateRangeFilter
                            value={dateFilters[col]}
                            onChange={(v) => handleDateFilter(col, v)}
                            hasFilter={!!hasFilter}
                          />
                        ) : (
                          <input
                            type="text"
                            placeholder="Search…"
                            value={filters[col] || ""}
                            onChange={(e) => handleFilter(col, e.target.value)}
                            className="h-8 w-40 rounded-md border bg-transparent px-2 text-sm outline-none focus:ring-1 focus:ring-ring"
                          />
                        )}
                        <button
                          onClick={() => handleSort(col)}
                          className="inline-flex items-center gap-1 rounded border px-2 py-1 text-xs hover:bg-muted"
                        >
                          {sortBy === col ? (sortOrder === "asc" ? "↑" : "↓") : "↕"}
                        </button>
                      </div>
                    )
                  })}
                </div>

                {/* Table */}
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        {columns.map((col) => (
                          <TableHead
                            key={col}
                            className="cursor-pointer select-none hover:text-foreground"
                            onClick={() => handleSort(col)}
                          >
                            <span className="flex items-center gap-1">
                              {col}
                              {sortBy === col && (
                                <span className="text-xs">{sortOrder === "asc" ? "↑" : "↓"}</span>
                              )}
                            </span>
                          </TableHead>
                        ))}
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {loading ? (
                        <TableRow>
                          <TableCell colSpan={columns.length} className="text-center py-8">
                            <Loader2 className="mx-auto h-6 w-6 animate-spin text-muted-foreground" />
                          </TableCell>
                        </TableRow>
                      ) : rows.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={columns.length} className="text-center py-8 text-muted-foreground">
                            No data found
                          </TableCell>
                        </TableRow>
                      ) : (
                        rows.map((row, i) => (
                          <TableRow key={i}>
                            {columns.map((col) => (
                              <TableCell key={col} className="max-w-[300px] truncate">
                                <span title={String(row[col] ?? "")}>
                                  {formatCellValue(row[col])}
                                </span>
                              </TableCell>
                            ))}
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                </div>

                {/* Pagination */}
                {pagination.totalPages > 1 && (
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">
                      Page {pagination.page} of {pagination.totalPages} ({pagination.total} rows)
                    </span>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => setPagination((p) => ({ ...p, page: 1 }))}
                        disabled={pagination.page === 1}
                        className="rounded-md border px-2 py-1 text-sm disabled:opacity-50"
                      >
                        «
                      </button>
                      <button
                        onClick={() =>
                          setPagination((p) => ({
                            ...p,
                            page: Math.max(1, p.page - 1),
                          }))
                        }
                        disabled={pagination.page === 1}
                        className="rounded-md border px-2 py-1 text-sm disabled:opacity-50"
                      >
                        <ChevronLeft className="h-4 w-4" />
                      </button>
                      <span className="text-sm">
                        {pagination.page} / {pagination.totalPages}
                      </span>
                      <button
                        onClick={() =>
                          setPagination((p) => ({
                            ...p,
                            page: Math.min(p.totalPages, p.page + 1),
                          }))
                        }
                        disabled={pagination.page === pagination.totalPages}
                        className="rounded-md border px-2 py-1 text-sm disabled:opacity-50"
                      >
                        <ChevronRight className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => setPagination((p) => ({ ...p, page: p.totalPages }))}
                        disabled={pagination.page === pagination.totalPages}
                        className="rounded-md border px-2 py-1 text-sm disabled:opacity-50"
                      >
                        »
                      </button>
                    </div>
                  </div>
                )}

                {/* Export */}
                {rows.length > 0 && (
                  <div className="flex gap-2">
                    <button
                      onClick={() => exportData(rows, columns, "json")}
                      className="inline-flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm hover:bg-muted"
                    >
                      <Download className="h-4 w-4" />
                      Export JSON
                    </button>
                    <button
                      onClick={() => exportData(rows, columns, "csv")}
                      className="inline-flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm hover:bg-muted"
                    >
                      <Download className="h-4 w-4" />
                      Export CSV
                    </button>
                  </div>
                )}
              </>
            )}
          </TabsContent>

          {/* SQL tab */}
          <TabsContent value="sql" className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">SQL Query (SELECT only)</label>
              <div className="relative">
                <textarea
                  value={sql}
                  onChange={(e) => setSql(e.target.value)}
                  placeholder="SELECT * FROM Product LIMIT 10;"
                  rows={6}
                  className="w-full rounded-md border bg-background p-3 font-mono text-sm outline-none focus:ring-1 focus:ring-ring"
                />
              </div>
              <div className="flex gap-2">
                <button
                  onClick={executeQuery}
                  disabled={queryLoading || !sql.trim()}
                  className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
                >
                  {queryLoading ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Code2 className="h-4 w-4" />
                  )}
                  Execute
                </button>
              </div>
            </div>

            {queryResult && (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">
                    {queryResult.rowCount} rows returned
                  </span>
                  <div className="flex gap-2">
                    <button
                      onClick={() => exportData(queryResult.rows, queryResult.columns, "json")}
                      className="inline-flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm hover:bg-muted"
                    >
                      <Download className="h-4 w-4" />
                      Export JSON
                    </button>
                    <button
                      onClick={() => exportData(queryResult.rows, queryResult.columns, "csv")}
                      className="inline-flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm hover:bg-muted"
                    >
                      <Download className="h-4 w-4" />
                      Export CSV
                    </button>
                  </div>
                </div>
                <div className="rounded-md border overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        {queryResult.columns.map((col) => (
                          <TableHead key={col}>{col}</TableHead>
                        ))}
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {queryResult.rows.map((row, i) => (
                        <TableRow key={i}>
                          {queryResult.columns.map((col) => (
                            <TableCell key={col}>
                              <span title={String(row[col] ?? "")}>
                                {truncate(row[col])}
                              </span>
                            </TableCell>
                          ))}
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </div>
            )}
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}

// Date range filter component
function DateRangeFilter({
  value,
  onChange,
  hasFilter,
}: {
  value?: { gte: string; lte: string }
  onChange: (v: { gte: string; lte: string }) => void
  hasFilter: boolean
}) {
  const [open, setOpen] = useState(false)
  const [gte, setGte] = useState(value?.gte || "")
  const [lte, setLte] = useState(value?.lte || "")

  const apply = () => {
    onChange({ gte, lte })
    setOpen(false)
  }

  const clear = () => {
    setGte("")
    setLte("")
    onChange({ gte: "", lte: "" })
    setOpen(false)
  }

  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        className={`inline-flex items-center gap-1 rounded border px-2 py-1 text-xs ${
          hasFilter ? "border-primary bg-primary/10" : "hover:bg-muted"
        }`}
      >
        📅
      </button>
      {open && (
        <div className="absolute top-full left-0 z-50 mt-1 rounded-md border bg-popover p-3 shadow-lg">
          <div className="flex flex-col gap-2">
            <input
              type="datetime-local"
              value={gte}
              onChange={(e) => setGte(e.target.value)}
              className="rounded border px-2 py-1 text-sm"
              placeholder="From"
            />
            <input
              type="datetime-local"
              value={lte}
              onChange={(e) => setLte(e.target.value)}
              className="rounded border px-2 py-1 text-sm"
              placeholder="To"
            />
            <div className="flex gap-2">
              <button onClick={apply} className="rounded bg-primary px-2 py-1 text-xs text-primary-foreground">
                Apply
              </button>
              <button onClick={clear} className="rounded border px-2 py-1 text-xs">
                Clear
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
