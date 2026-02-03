import React, { useMemo } from "react";
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  PieChart,
  Pie,
  Cell,
  CartesianGrid,
  Legend,
  LabelList,
} from "recharts";
import "./ChartView.css";

function formatCurrency(value: number) {
  if (value == null) return "-";
  return new Intl.NumberFormat(undefined, {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(value);
}

function CustomTooltip({ active, payload }: any) {
  if (!active || !payload || !payload.length) return null;
  const p = payload[0];
  return (
    <div
      style={{
        background: "white",
        padding: 8,
        borderRadius: 6,
        boxShadow: "0 6px 18px rgba(34,64,130,0.12)",
      }}
    >
      <div style={{ fontWeight: 600 }}>{p.name || p.payload.productName}</div>
      <div style={{ color: "#334155" }}>{formatCurrency(p.value)}</div>
    </div>
  );
}

const PIE_COLORS = [
  "#60a5fa",
  "#34d399",
  "#f59e0b",
  "#f97316",
  "#ef4444",
  "#7c3aed",
];

export default function ChartView({ response }: any) {
  const data = response.data || [];

  // Note: keep hooks in top-level scope (even if data is empty) so they are
  // called in the same order on every render (fixes react-hooks/rules-of-hooks).

  // try to detect a revenue-like numeric key
  const revenueKey = useMemo(() => {
    const sample = data[0] || {};
    const keys = Object.keys(sample);
    const found = keys.find((k) =>
      /revenue|amount|total|value|sales|price/i.test(k),
    );
    if (found) return found;
    const numKey = keys.find((k) => typeof sample[k] === "number");
    return numKey || keys[0] || "value";
  }, [data]);

  const chartData = useMemo(() => {
    return data.map((d: any) => ({
      ...d,
      revenue: Number(d[revenueKey] ?? 0),
    }));
  }, [data, revenueKey]);

  const totals = useMemo(() => {
    const total = chartData.reduce(
      (s: number, r: any) => s + (Number(r.revenue) || 0),
      0,
    );
    const avg = chartData.length ? total / chartData.length : 0;
    const top = chartData.reduce(
      (best: any, curr: any) =>
        curr.revenue > (best.revenue || 0) ? curr : best,
      {} as any,
    );
    return { total, avg, topProduct: top.productName || top.name || "-" };
  }, [chartData]);

  function exportCSV() {
    const headers = Object.keys(chartData[0] || {});
    const rows = chartData.map((r: any) =>
      headers.map((h) => JSON.stringify(r[h] ?? "")),
    );
    const csv = [
      headers.join(","),
      ...rows.map((r: string[]) => r.join(",")),
    ].join("\n");
    const blob = new Blob([csv], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${response.chart || "chart"}-data.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  const title =
    response.title ||
    (response.chart === "pie" ? "Distribution" : "Revenue by Product");

  // If no data available, render a friendly placeholder (hooks still run above).
  if (!chartData.length) {
    return (
      <div className="chart-card">
        <div className="chart-header">
          <div>
            <div className="chart-title">{title}</div>
            <div className="stats-row" style={{ marginTop: 8 }}>
              <div className="stat-chip">
                Total: {formatCurrency(totals.total)}
              </div>
              <div className="stat-chip">Avg: {formatCurrency(totals.avg)}</div>
              <div className="stat-chip">Top: {totals.topProduct}</div>
            </div>
          </div>

          <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <button className="export-btn" onClick={exportCSV} disabled>
              Export CSV
            </button>
          </div>
        </div>

        <div
          className="chart-area"
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            color: "#64748b",
          }}
        >
          No data available
        </div>
      </div>
    );
  }

  return (
    <div className="chart-card">
      <div className="chart-header">
        <div>
          <div className="chart-title">{title}</div>
          <div className="stats-row" style={{ marginTop: 8 }}>
            <div className="stat-chip">
              Total: {formatCurrency(totals.total)}
            </div>
            <div className="stat-chip">Avg: {formatCurrency(totals.avg)}</div>
            <div className="stat-chip">Top: {totals.topProduct}</div>
          </div>
        </div>

        <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
          <button className="export-btn" onClick={exportCSV}>
            Export CSV
          </button>
        </div>
      </div>

      <div className="chart-area">
        <ResponsiveContainer width="100%" height="100%">
          {response.chart === "pie" ? (
            <PieChart>
              <Pie
                data={chartData}
                dataKey="revenue"
                nameKey={response.nameKey || "productName"}
                outerRadius={100}
                label
              >
                {chartData.map((d: any, i: number) => (
                  <Cell
                    key={`cell-${i}`}
                    fill={PIE_COLORS[i % PIE_COLORS.length]}
                  />
                ))}
              </Pie>
              <Tooltip content={<CustomTooltip />} />
              <Legend />
            </PieChart>
          ) : (
            <BarChart
              data={chartData}
              margin={{ top: 20, right: 20, left: 0, bottom: 40 }}
            >
              <defs>
                <linearGradient id="colorRevenue" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#60a5fa" stopOpacity={0.9} />
                  <stop offset="100%" stopColor="#2563eb" stopOpacity={0.8} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#e6eef9" />
              <XAxis
                dataKey={response.xKey || "productName"}
                angle={-20}
                textAnchor="end"
                height={60}
              />
              <YAxis tickFormatter={formatCurrency} />
              <Tooltip content={<CustomTooltip />} />
              <Legend />
              <Bar dataKey="revenue" fill="url(#colorRevenue)">
                <LabelList
                  dataKey="revenue"
                  formatter={(v: any) => formatCurrency(v)}
                  position="top"
                />
              </Bar>
            </BarChart>
          )}
        </ResponsiveContainer>
      </div>

      {response.chart === "pie" && (
        <div className="pie-legend">
          {chartData.map((d: any, i: number) => (
            <div key={i} className="pie-legend-item">
              <div
                className="color-swatch"
                style={{ background: PIE_COLORS[i % PIE_COLORS.length] }}
              />
              <div style={{ color: "#334155" }}>
                {d.productName || d.name || `Item ${i + 1}`} —{" "}
                {formatCurrency(d.revenue)}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
