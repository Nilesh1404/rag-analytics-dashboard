import { Link } from "react-router-dom";

export default function Home() {
  return (
    <div style={{ padding: 40 }}>
      <h1>Analytics RAG System</h1>
      <p>Ask questions about product sales using natural language.</p>

      <Link to="/chat">
        <button>Go to Analytics Chat</button>
      </Link>
    </div>
  );
}

