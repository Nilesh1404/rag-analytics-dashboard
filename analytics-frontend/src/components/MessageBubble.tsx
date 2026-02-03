import React, { useState } from "react";
import ChartView from "./ChartView";

function formatTime(iso?: string) {
  if (!iso) return "";
  try {
    const d = new Date(iso);
    return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  } catch {
    return "";
  }
}

export default function MessageBubble({ msg }: any) {
  const [copied, setCopied] = useState(false);

  const copyText = async () => {
    if (!msg.text) return;
    try {
      await navigator.clipboard.writeText(msg.text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (e) {
      console.error(e);
    }
  };

  const exportData = () => {
    if (!msg.data?.length) return;
    const headers = Object.keys(msg.data[0] || {});
    const rows = msg.data.map((r: any) =>
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
    a.download = `${msg.chart || "chart"}-data.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className={`bubble ${msg.role}`}>
      <div className="avatar">{msg.role === "assistant" ? "🤖" : "👤"}</div>

      <div className="bubble-content">
        <div className="bubble-header">
          <div className="bubble-sender">
            {msg.role === "assistant" ? "Assistant" : "You"}
          </div>
          <div className="bubble-meta">{formatTime(msg.time)}</div>
        </div>

        {msg.text && <p className="bubble-text">{msg.text}</p>}

        <div className="bubble-actions">
          {msg.text && (
            <button
              className="action-btn"
              onClick={copyText}
              aria-label="Copy message text"
            >
              {copied ? "Copied" : "Copy"}
            </button>
          )}

          {msg.data?.length > 0 && (
            <button
              className="action-btn"
              onClick={exportData}
              aria-label="Download data"
            >
              Download Data
            </button>
          )}
        </div>

        {msg.data?.length > 0 && <ChartView response={msg} />}

        {msg.image && (
          <img
            src={`data:image/png;base64,${msg.image}`}
            className="bubble-image"
          />
        )}
      </div>
    </div>
  );
}
