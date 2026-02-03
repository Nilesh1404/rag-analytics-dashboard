import { useState, useRef, useEffect } from "react";
import axios from "axios";
import MessageBubble from "../components/MessageBubble";
import "./chat.css";

export default function Chat() {
  const [messages, setMessages] = useState<any[]>([]);
  const [text, setText] = useState("");
  const [loading, setLoading] = useState(false);
  const messagesRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (messagesRef.current) {
      messagesRef.current.scrollTop = messagesRef.current.scrollHeight;
    }
  }, [messages]);

  async function send() {
    if (!text) return;

    const userMsg = { role: "user", text, time: new Date().toISOString() };
    setMessages((m) => [...m, userMsg]);
    setLoading(true);

    try {
      const res = await axios.post("http://localhost:8080/rag", {
        query: text,
      });
      console.log("Response:", res);
      const assistantMsg = {
        role: "assistant",
        text: res.data.summary,
        chart: res.data.chart,
        data: res.data.data,
        image: res.data.image,
        time: new Date().toISOString(),
      };
      setMessages((m) => [...m, assistantMsg]);
    } catch (err) {
      setMessages((m) => [
        ...m,
        {
          role: "assistant",
          text: "Sorry, something went wrong.",
          time: new Date().toISOString(),
        },
      ]);
      console.error(err);
    } finally {
      setText("");
      setLoading(false);
    }
  }

  return (
    <div className="chat-page">
      <div className="chat-container">
        <div className="header">📊 Analytics Assistant</div>

        <div className="messages" ref={messagesRef}>
          {messages.map((m, i) => (
            <MessageBubble key={i} msg={m} />
          ))}

          {loading && <div className="typing">thinking...</div>}
        </div>

        <div className="input-bar">
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Ask about your sales..."
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                send();
              }
            }}
            rows={1}
            disabled={loading}
            aria-disabled={loading}
          />
          <button
            onClick={send}
            disabled={loading || !text}
            aria-disabled={loading || !text}
          >
            {loading ? "Sending..." : "Send"}
          </button>
        </div>
      </div>
    </div>
  );
}
